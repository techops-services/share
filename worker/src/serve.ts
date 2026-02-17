import type { Context } from "hono";
import { expiredPage } from "./pages/expired";
import { notFoundPage } from "./pages/not-found";
import type { Env, ExpiredMarker, ShareMetadata } from "./types";

const SECURITY_HEADERS = {
  "Content-Security-Policy":
    "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: blob: https:; font-src data:; media-src data: blob:; connect-src 'none'; form-action 'none'; frame-ancestors 'none'; base-uri 'none'",
  "X-Content-Type-Options": "nosniff",
  "X-Frame-Options": "DENY",
  "Referrer-Policy": "no-referrer",
} as const;

function htmlResponse(
  body: string,
  status: number,
  extraHeaders: Record<string, string> = {},
): Response {
  return new Response(body, {
    status,
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      ...SECURITY_HEADERS,
      ...extraHeaders,
    },
  });
}

async function incrementViewCount(
  meta: KVNamespace,
  id: string,
  metadata: ShareMetadata,
): Promise<void> {
  const updated: ShareMetadata = {
    ...metadata,
    views: metadata.views + 1,
  };

  const remainingTtl = Math.max(
    1,
    Math.floor(
      (new Date(metadata.expires_at).getTime() - Date.now()) / 1000,
    ),
  );

  await meta.put(`share:${id}`, JSON.stringify(updated), {
    expirationTtl: remainingTtl,
  });
}

export async function handleServe(
  c: Context<{ Bindings: Env }>,
): Promise<Response> {
  const id = c.req.param("id");
  if (!id) {
    return htmlResponse(notFoundPage(""), 404);
  }

  const { META, PAGES } = c.env;

  const expiredMarker = await META.get<ExpiredMarker>(
    `expired:${id}`,
    "json",
  );
  if (expiredMarker) {
    return htmlResponse(
      expiredPage(id, expiredMarker.expired_at),
      410,
      { "Cache-Control": "public, max-age=60" },
    );
  }

  const metadata = await META.get<ShareMetadata>(`share:${id}`, "json");
  if (!metadata) {
    return htmlResponse(
      notFoundPage(id),
      404,
      { "Cache-Control": "no-cache" },
    );
  }

  const now = new Date();
  const expiresAt = new Date(metadata.expires_at);
  if (now >= expiresAt) {
    return htmlResponse(
      expiredPage(id, metadata.expires_at),
      410,
      { "Cache-Control": "public, max-age=60" },
    );
  }

  const object = await PAGES.get(`${id}.html`);
  if (!object) {
    return htmlResponse(
      notFoundPage(id),
      404,
      { "Cache-Control": "no-cache" },
    );
  }

  const body = await object.text();

  c.executionCtx.waitUntil(incrementViewCount(META, id, metadata));

  return htmlResponse(body, 200, {
    "Cache-Control": "public, max-age=3600, immutable",
    "X-Share-Expires": metadata.expires_at,
    "X-Share-Id": id,
  });
}
