import type { Context } from "hono";
import type { Env, ErrorCode, ErrorResponse } from "./types";

const STATUS_MAP: Record<ErrorCode, number> = {
  INVALID_CONTENT_TYPE: 400,
  INVALID_TTL: 400,
  PAYLOAD_TOO_LARGE: 413,
  RATE_LIMITED: 429,
  NOT_FOUND: 404,
  GONE: 410,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  INTERNAL_ERROR: 500,
} as const;

export function errorResponse(
  c: Context<{ Bindings: Env }>,
  code: ErrorCode,
  message: string,
  retryAfter?: number,
): Response {
  const status = STATUS_MAP[code];
  const body: ErrorResponse = {
    error: {
      code,
      message,
      ...(retryAfter !== undefined && { retry_after: retryAfter }),
    },
  };

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (retryAfter !== undefined) {
    headers["Retry-After"] = String(retryAfter);
  }

  return c.json(body, { status, headers });
}
