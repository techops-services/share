import { nanoid } from "nanoid";
import type { Context } from "hono";
import { extractApiKey, getApiKeyHash, validateApiKey } from "./auth";
import { errorResponse } from "./errors";
import {
  checkRateLimit,
  getRateLimitIdentifier,
  incrementRateLimit,
  setRateLimitHeaders,
} from "./ratelimit";
import type { Env, ShareMetadata, UploadResponse } from "./types";

const VALID_CONTENT_TYPES = ["text/html", "text/html; charset=utf-8"] as const;

function isValidContentType(contentType: string | undefined): boolean {
  if (!contentType) {
    return false;
  }
  const normalized = contentType.toLowerCase().replace(/\s/g, "");
  return VALID_CONTENT_TYPES.some(
    (valid) => normalized === valid.replace(/\s/g, ""),
  );
}

function parseTtl(
  ttlHeader: string | undefined,
  defaultTtl: number,
  minTtl: number,
  maxTtl: number,
): { valid: true; ttl: number } | { valid: false; message: string } {
  if (!ttlHeader) {
    return { valid: true, ttl: defaultTtl };
  }

  const parsed = Number.parseInt(ttlHeader, 10);
  if (Number.isNaN(parsed) || !Number.isFinite(parsed) || parsed !== Number(ttlHeader)) {
    return { valid: false, message: "X-TTL must be a valid integer" };
  }

  if (parsed < minTtl) {
    return {
      valid: false,
      message: `X-TTL must be at least ${minTtl} seconds`,
    };
  }

  if (parsed > maxTtl) {
    return {
      valid: false,
      message: `X-TTL must not exceed ${maxTtl} seconds`,
    };
  }

  return { valid: true, ttl: parsed };
}

export async function handleUpload(
  c: Context<{ Bindings: Env }>,
): Promise<Response> {
  const { META, PAGES } = c.env;

  const contentType = c.req.header("Content-Type");
  if (!isValidContentType(contentType)) {
    return errorResponse(
      c,
      "INVALID_CONTENT_TYPE",
      "Content-Type must be text/html or text/html; charset=utf-8",
    );
  }

  const apiKey = extractApiKey(c);
  let apiKeyHash: string | null = null;
  let isAuth = false;

  if (apiKey) {
    const record = await validateApiKey(META, apiKey);
    if (!record) {
      return errorResponse(c, "UNAUTHORIZED", "Invalid API key");
    }
    apiKeyHash = await getApiKeyHash(apiKey);
    isAuth = true;
  }

  const ip = c.req.header("CF-Connecting-IP") ?? c.req.header("X-Forwarded-For");
  const rateLimitId = getRateLimitIdentifier(ip ?? null, apiKeyHash);
  const limit = isAuth
    ? Number.parseInt(c.env.AUTH_RATE_LIMIT, 10)
    : Number.parseInt(c.env.ANON_RATE_LIMIT, 10);

  const rateLimitResult = await checkRateLimit(META, rateLimitId, limit);
  const responseHeaders = new Headers();
  setRateLimitHeaders(responseHeaders, rateLimitResult);

  if (!rateLimitResult.allowed) {
    const retryAfter = rateLimitResult.retryAfter ?? 1;
    const resp = errorResponse(
      c,
      "RATE_LIMITED",
      `Rate limit exceeded. Try again in ${retryAfter} seconds.`,
      retryAfter,
    );
    setRateLimitHeaders(resp.headers, rateLimitResult);
    return resp;
  }

  const maxTtl = isAuth
    ? Number.parseInt(c.env.MAX_TTL_AUTHENTICATED, 10)
    : Number.parseInt(c.env.MAX_TTL_ANONYMOUS, 10);
  const defaultTtl = Number.parseInt(c.env.DEFAULT_TTL, 10);
  const minTtl = Number.parseInt(c.env.MIN_TTL, 10);

  const ttlResult = parseTtl(
    c.req.header("X-TTL"),
    defaultTtl,
    minTtl,
    maxTtl,
  );
  if (!ttlResult.valid) {
    return errorResponse(c, "INVALID_TTL", ttlResult.message);
  }

  const body = await c.req.arrayBuffer();
  const maxSize = Number.parseInt(c.env.MAX_UPLOAD_SIZE, 10);

  if (body.byteLength === 0) {
    return errorResponse(
      c,
      "INVALID_CONTENT_TYPE",
      "Request body must not be empty",
    );
  }

  if (body.byteLength > maxSize) {
    return errorResponse(
      c,
      "PAYLOAD_TOO_LARGE",
      `Payload exceeds maximum size of ${maxSize} bytes`,
    );
  }

  const id = nanoid(8);
  const now = new Date();
  const expiresAt = new Date(now.getTime() + ttlResult.ttl * 1000);

  const metadata: ShareMetadata = {
    id,
    size: body.byteLength,
    ttl: ttlResult.ttl,
    created_at: now.toISOString(),
    expires_at: expiresAt.toISOString(),
    content_type: "text/html",
    api_key_hash: apiKeyHash,
    views: 0,
  };

  await PAGES.put(`${id}.html`, body, {
    httpMetadata: {
      contentType: "text/html; charset=utf-8",
    },
    customMetadata: {
      expires_at: expiresAt.toISOString(),
    },
  });

  await META.put(`share:${id}`, JSON.stringify(metadata), {
    expirationTtl: ttlResult.ttl,
  });

  await incrementRateLimit(META, rateLimitId);

  const response: UploadResponse = {
    id,
    url: `${c.env.BASE_URL}/${id}`,
    expires_at: expiresAt.toISOString(),
    size: body.byteLength,
    ttl: ttlResult.ttl,
  };

  return c.json(response, {
    status: 201,
    headers: Object.fromEntries(responseHeaders.entries()),
  });
}
