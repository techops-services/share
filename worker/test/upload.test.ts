import {
  env,
  createExecutionContext,
  waitOnExecutionContext,
} from "cloudflare:test";
import { describe, it, expect, beforeEach } from "vitest";
import worker from "../src/index";
import type { Env } from "../src/types";

const BASE_URL = "https://sh.techops.services";

function makeRequest(
  path: string,
  options: RequestInit = {},
): Request {
  return new Request(`${BASE_URL}${path}`, options);
}

describe("POST /api/upload", () => {
  it("rejects requests without text/html content type", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "<h1>Hello</h1>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(400);
    const body = await resp.json() as { error: { code: string } };
    expect(body.error.code).toBe("INVALID_CONTENT_TYPE");
  });

  it("rejects empty body", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: { "Content-Type": "text/html" },
      body: "",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(400);
    const body = await resp.json() as { error: { code: string } };
    expect(body.error.code).toBe("INVALID_CONTENT_TYPE");
  });

  it("accepts valid HTML upload and returns 201", async () => {
    const ctx = createExecutionContext();
    const html = "<html><body><h1>Test Page</h1></body></html>";
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: { "Content-Type": "text/html" },
      body: html,
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(201);
    const body = await resp.json() as {
      id: string;
      url: string;
      expires_at: string;
      size: number;
      ttl: number;
    };
    expect(body.id).toBeTruthy();
    expect(body.id).toHaveLength(8);
    expect(body.url).toContain(body.id);
    expect(body.size).toBe(new TextEncoder().encode(html).byteLength);
    expect(body.ttl).toBe(86400);
    expect(body.expires_at).toBeTruthy();
  });

  it("accepts text/html; charset=utf-8 content type", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: { "Content-Type": "text/html; charset=utf-8" },
      body: "<p>UTF-8 content</p>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(201);
  });

  it("rejects invalid TTL below minimum", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: {
        "Content-Type": "text/html",
        "X-TTL": "100",
      },
      body: "<p>Short TTL</p>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(400);
    const body = await resp.json() as { error: { code: string } };
    expect(body.error.code).toBe("INVALID_TTL");
  });

  it("rejects non-numeric TTL", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: {
        "Content-Type": "text/html",
        "X-TTL": "abc",
      },
      body: "<p>Bad TTL</p>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(400);
    const body = await resp.json() as { error: { code: string } };
    expect(body.error.code).toBe("INVALID_TTL");
  });

  it("uses custom TTL when provided", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: {
        "Content-Type": "text/html",
        "X-TTL": "3600",
      },
      body: "<p>Custom TTL</p>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(201);
    const body = await resp.json() as { ttl: number };
    expect(body.ttl).toBe(3600);
  });

  it("includes rate limit headers in response", async () => {
    const ctx = createExecutionContext();
    const req = makeRequest("/api/upload", {
      method: "POST",
      headers: { "Content-Type": "text/html" },
      body: "<p>Rate limit test</p>",
    });

    const resp = await worker.fetch(req, env as unknown as Env, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.headers.get("X-RateLimit-Limit")).toBeTruthy();
    expect(resp.headers.get("X-RateLimit-Remaining")).toBeTruthy();
    expect(resp.headers.get("X-RateLimit-Reset")).toBeTruthy();
  });
});
