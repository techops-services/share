import {
  env,
  createExecutionContext,
  waitOnExecutionContext,
} from "cloudflare:test";
import { describe, it, expect } from "vitest";
import worker from "../src/index";
import type { Env } from "../src/types";

const BASE_URL = "https://sh.techops.services";

function makeRequest(
  path: string,
  options: RequestInit = {},
): Request {
  return new Request(`${BASE_URL}${path}`, options);
}

async function uploadHtml(html: string): Promise<string> {
  const ctx = createExecutionContext();
  const req = makeRequest("/api/upload", {
    method: "POST",
    headers: { "Content-Type": "text/html" },
    body: html,
  });
  const resp = await worker.fetch(req, env as unknown as Env, ctx);
  await waitOnExecutionContext(ctx);
  const body = await resp.json() as { id: string };
  return body.id;
}

describe("GET /:id", () => {
  it("returns 404 for non-existent share", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/nonexist");

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(404);
    expect(resp.headers.get("Content-Type")).toContain("text/html");
    const body = await resp.text();
    expect(body).toContain("404");
    expect(body).toContain("not found");
  });

  it("serves uploaded HTML with security headers", async () => {
    const html = "<html><body><h1>Secure Page</h1></body></html>";
    const id = await uploadHtml(html);

    const ctx = createExecutionContext();
    const req = makeRequest(`/${id}`);
    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(200);
    expect(resp.headers.get("Content-Type")).toContain("text/html");
    expect(resp.headers.get("Content-Security-Policy")).toBeTruthy();
    expect(resp.headers.get("X-Content-Type-Options")).toBe("nosniff");
    expect(resp.headers.get("X-Frame-Options")).toBe("DENY");
    expect(resp.headers.get("Referrer-Policy")).toBe("no-referrer");
    expect(resp.headers.get("X-Share-Id")).toBe(id);
    expect(resp.headers.get("X-Share-Expires")).toBeTruthy();

    const body = await resp.text();
    expect(body).toBe(html);
  });

  it("returns cache-control header on served content", async () => {
    const id = await uploadHtml("<p>Cacheable</p>");

    const ctx = createExecutionContext();
    const req = makeRequest(`/${id}`);
    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.headers.get("Cache-Control")).toBe(
      "public, max-age=3600, immutable",
    );
  });

  it("returns 410 for expired marker", async () => {
    const typedEnv = env as unknown as Env;
    const expiredId = "expired1";
    await typedEnv.META.put(
      `expired:${expiredId}`,
      JSON.stringify({ expired_at: "2024-01-01T00:00:00.000Z" }),
      { expirationTtl: 604800 },
    );

    const ctx = createExecutionContext();
    const req = makeRequest(`/${expiredId}`);
    const resp = await worker.fetch(req, typedEnv, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(410);
    const body = await resp.text();
    expect(body).toContain("410");
    expect(body).toContain("expired");
  });
});
