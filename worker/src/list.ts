import type { Context } from "hono";
import { extractApiKey, getApiKeyHash, validateApiKey } from "./auth";
import { errorResponse } from "./errors";
import type { Env, ListResponse, ShareMetadata } from "./types";

const DEFAULT_LIMIT = 50;
const MAX_LIMIT = 100;

export async function handleList(
  c: Context<{ Bindings: Env }>,
): Promise<Response> {
  const { META } = c.env;

  const apiKey = extractApiKey(c);
  if (!apiKey) {
    return errorResponse(
      c,
      "UNAUTHORIZED",
      "API key required. Provide X-Api-Key header.",
    );
  }

  const record = await validateApiKey(META, apiKey);
  if (!record) {
    return errorResponse(c, "UNAUTHORIZED", "Invalid API key");
  }

  const apiKeyHash = await getApiKeyHash(apiKey);

  const limitParam = c.req.query("limit");
  let limit = DEFAULT_LIMIT;
  if (limitParam) {
    const parsed = Number.parseInt(limitParam, 10);
    if (!Number.isNaN(parsed) && parsed > 0) {
      limit = Math.min(parsed, MAX_LIMIT);
    }
  }

  const cursor = c.req.query("cursor") ?? undefined;

  const listResult = await META.list({
    prefix: "share:",
    limit,
    cursor,
  });

  const shares: ShareMetadata[] = [];

  for (const key of listResult.keys) {
    const metadata = await META.get<ShareMetadata>(key.name, "json");
    if (!metadata) {
      continue;
    }
    if (metadata.api_key_hash !== apiKeyHash) {
      continue;
    }
    shares.push(metadata);
  }

  const response: ListResponse = {
    shares,
    cursor: listResult.list_complete ? null : (listResult.cursor ?? null),
    total: shares.length,
  };

  return c.json(response, 200);
}
