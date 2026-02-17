import {
  env,
  createExecutionContext,
  waitOnExecutionContext,
} from "cloudflare:test";
import { describe, it, expect } from "vitest";
import { handleCleanup } from "../src/cleanup";
import worker from "../src/index";
import type { Env, ShareMetadata } from "../src/types";

const BASE_URL = "https://share.techops.services";
const typedEnv = env as unknown as Env;

function makeRequest(
  path: string,
  options: RequestInit = {},
): Request {
  return new Request(`${BASE_URL}${path}`, options);
}

describe("Cron Cleanup", () => {
  it("removes expired shares and creates expired markers", async () => {
    const expiredMeta: ShareMetadata = {
      id: "cleanup1",
      size: 100,
      ttl: 1,
      created_at: "2024-01-01T00:00:00.000Z",
      expires_at: "2024-01-01T00:00:01.000Z",
      content_type: "text/html",
      api_key_hash: null,
      views: 5,
    };

    await typedEnv.META.put(
      "share:cleanup1",
      JSON.stringify(expiredMeta),
    );
    await typedEnv.PAGES.put("cleanup1.html", "<p>Old page</p>");

    await handleCleanup(typedEnv);

    const metaAfter = await typedEnv.META.get("share:cleanup1");
    expect(metaAfter).toBeNull();

    const r2After = await typedEnv.PAGES.get("cleanup1.html");
    expect(r2After).toBeNull();

    const expiredMarker = await typedEnv.META.get("expired:cleanup1");
    expect(expiredMarker).toBeTruthy();
  });

  it("does not remove non-expired shares", async () => {
    const futureMeta: ShareMetadata = {
      id: "cleanup2",
      size: 200,
      ttl: 86400,
      created_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 86400 * 1000).toISOString(),
      content_type: "text/html",
      api_key_hash: null,
      views: 0,
    };

    await typedEnv.META.put(
      "share:cleanup2",
      JSON.stringify(futureMeta),
    );
    await typedEnv.PAGES.put("cleanup2.html", "<p>Future page</p>");

    await handleCleanup(typedEnv);

    const metaAfter = await typedEnv.META.get("share:cleanup2");
    expect(metaAfter).toBeTruthy();

    const r2After = await typedEnv.PAGES.get("cleanup2.html");
    expect(r2After).toBeTruthy();
  });

  it("cleaned share returns 410 on serve", async () => {
    const expiredMeta: ShareMetadata = {
      id: "cleanup3",
      size: 50,
      ttl: 1,
      created_at: "2024-01-01T00:00:00.000Z",
      expires_at: "2024-01-01T00:00:01.000Z",
      content_type: "text/html",
      api_key_hash: null,
      views: 0,
    };

    await typedEnv.META.put(
      "share:cleanup3",
      JSON.stringify(expiredMeta),
    );
    await typedEnv.PAGES.put("cleanup3.html", "<p>Will expire</p>");

    await handleCleanup(typedEnv);

    const ctx = createExecutionContext();
    const req = makeRequest("/cleanup3");
    const resp = await worker.fetch(req, typedEnv, ctx);
    await waitOnExecutionContext(ctx);

    expect(resp.status).toBe(410);
    const body = await resp.text();
    expect(body).toContain("expired");
  });
});
