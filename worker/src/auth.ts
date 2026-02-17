import type { Context } from "hono";
import type { ApiKeyRecord, Env } from "./types";

const API_KEY_PATTERN = /^sk_live_[0-9a-f]{32}$/;

async function hashApiKey(key: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(key);
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  const hashArray = new Uint8Array(hashBuffer);
  return Array.from(hashArray)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

export function extractApiKey(c: Context<{ Bindings: Env }>): string | null {
  const header = c.req.header("X-Api-Key");
  if (!header) {
    return null;
  }

  if (!API_KEY_PATTERN.test(header)) {
    return null;
  }

  return header;
}

export async function validateApiKey(
  meta: KVNamespace,
  apiKey: string,
): Promise<ApiKeyRecord | null> {
  const keyHash = await hashApiKey(apiKey);
  const record = await meta.get<ApiKeyRecord>(`apikey:${keyHash}`, "json");
  return record;
}

export async function getApiKeyHash(apiKey: string): Promise<string> {
  return hashApiKey(apiKey);
}

export function isAuthenticated(
  c: Context<{ Bindings: Env }>,
): { authenticated: true; apiKey: string } | { authenticated: false } {
  const apiKey = extractApiKey(c);
  if (!apiKey) {
    return { authenticated: false };
  }
  return { authenticated: true, apiKey };
}
