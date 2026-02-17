import type { Context } from "hono";
import { extractApiKey, getApiKeyHash, validateApiKey } from "./auth";
import { errorResponse } from "./errors";
import type { Env, ShareMetadata } from "./types";

export async function handleDelete(
  c: Context<{ Bindings: Env }>,
): Promise<Response> {
  const id = c.req.param("id");
  if (!id) {
    return errorResponse(c, "NOT_FOUND", "Share not found");
  }

  const { META, PAGES } = c.env;

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

  const metadata = await META.get<ShareMetadata>(`share:${id}`, "json");
  if (!metadata) {
    return errorResponse(c, "NOT_FOUND", "Share not found");
  }

  const requestKeyHash = await getApiKeyHash(apiKey);
  if (metadata.api_key_hash !== requestKeyHash) {
    return errorResponse(
      c,
      "FORBIDDEN",
      "You do not have permission to delete this share",
    );
  }

  await PAGES.delete(`${id}.html`);
  await META.delete(`share:${id}`);

  await META.put(
    `expired:${id}`,
    JSON.stringify({ expired_at: new Date().toISOString() }),
    { expirationTtl: 604800 },
  );

  return new Response(null, { status: 204 });
}
