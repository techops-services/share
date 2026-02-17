import { env } from "cloudflare:test";
import { describe, it, expect } from "vitest";
import {
  checkRateLimit,
  incrementRateLimit,
  getRateLimitIdentifier,
  setRateLimitHeaders,
} from "../src/ratelimit";
import type { Env } from "../src/types";

const meta = (env as unknown as Env).META;

describe("Rate Limiting", () => {
  describe("getRateLimitIdentifier", () => {
    it("uses API key hash when available", () => {
      const id = getRateLimitIdentifier("1.2.3.4", "abc123hash");
      expect(id).toBe("key:abc123hash");
    });

    it("falls back to IP when no API key", () => {
      const id = getRateLimitIdentifier("1.2.3.4", null);
      expect(id).toBe("ip:1.2.3.4");
    });

    it("uses unknown when no IP and no key", () => {
      const id = getRateLimitIdentifier(null, null);
      expect(id).toBe("ip:unknown");
    });
  });

  describe("checkRateLimit", () => {
    it("allows first request", async () => {
      const result = await checkRateLimit(meta, "test-fresh-ip", 10);
      expect(result.allowed).toBe(true);
      expect(result.limit).toBe(10);
      expect(result.remaining).toBe(9);
    });

    it("decrements remaining after increment", async () => {
      const identifier = "test-decrement-ip";
      await incrementRateLimit(meta, identifier);
      await incrementRateLimit(meta, identifier);

      const result = await checkRateLimit(meta, identifier, 10);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(7);
    });

    it("blocks when limit is reached", async () => {
      const identifier = "test-blocked-ip";
      for (let i = 0; i < 5; i++) {
        await incrementRateLimit(meta, identifier);
      }

      const result = await checkRateLimit(meta, identifier, 5);
      expect(result.allowed).toBe(false);
      expect(result.remaining).toBe(0);
      expect(result.retryAfter).toBeGreaterThan(0);
    });
  });

  describe("setRateLimitHeaders", () => {
    it("sets all three rate limit headers", () => {
      const headers = new Headers();
      setRateLimitHeaders(headers, {
        allowed: true,
        limit: 10,
        remaining: 7,
        reset: 1708300800,
      });

      expect(headers.get("X-RateLimit-Limit")).toBe("10");
      expect(headers.get("X-RateLimit-Remaining")).toBe("7");
      expect(headers.get("X-RateLimit-Reset")).toBe("1708300800");
    });

    it("clamps remaining to zero minimum", () => {
      const headers = new Headers();
      setRateLimitHeaders(headers, {
        allowed: false,
        limit: 10,
        remaining: -1,
        reset: 1708300800,
      });

      expect(headers.get("X-RateLimit-Remaining")).toBe("0");
    });
  });
});
