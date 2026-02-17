import type { Env, RateLimitEntry, RateLimitResult } from "./types";

function getHourBucket(): string {
  const now = new Date();
  return `${now.getUTCFullYear()}${String(now.getUTCMonth() + 1).padStart(2, "0")}${String(now.getUTCDate()).padStart(2, "0")}${String(now.getUTCHours()).padStart(2, "0")}`;
}

function getHourResetTimestamp(): number {
  const now = new Date();
  const reset = new Date(now);
  reset.setUTCMinutes(0, 0, 0);
  reset.setUTCHours(reset.getUTCHours() + 1);
  return Math.floor(reset.getTime() / 1000);
}

export async function checkRateLimit(
  meta: KVNamespace,
  identifier: string,
  limit: number,
): Promise<RateLimitResult> {
  const bucket = getHourBucket();
  const key = `ratelimit:${identifier}:${bucket}`;
  const reset = getHourResetTimestamp();

  const entry = await meta.get<RateLimitEntry>(key, "json");
  const currentCount = entry?.count ?? 0;

  if (currentCount >= limit) {
    const retryAfter = reset - Math.floor(Date.now() / 1000);
    return {
      allowed: false,
      limit,
      remaining: 0,
      reset,
      retryAfter: Math.max(retryAfter, 1),
    };
  }

  return {
    allowed: true,
    limit,
    remaining: limit - currentCount - 1,
    reset,
  };
}

export async function incrementRateLimit(
  meta: KVNamespace,
  identifier: string,
): Promise<void> {
  const bucket = getHourBucket();
  const key = `ratelimit:${identifier}:${bucket}`;

  const entry = await meta.get<RateLimitEntry>(key, "json");
  const currentCount = entry?.count ?? 0;

  await meta.put(key, JSON.stringify({ count: currentCount + 1 }), {
    expirationTtl: 3600,
  });
}

export function setRateLimitHeaders(
  headers: Headers,
  result: RateLimitResult,
): void {
  headers.set("X-RateLimit-Limit", String(result.limit));
  headers.set("X-RateLimit-Remaining", String(Math.max(result.remaining, 0)));
  headers.set("X-RateLimit-Reset", String(result.reset));
}

export function getRateLimitIdentifier(
  ip: string | null,
  apiKeyHash: string | null,
): string {
  if (apiKeyHash) {
    return `key:${apiKeyHash}`;
  }
  return `ip:${ip ?? "unknown"}`;
}
