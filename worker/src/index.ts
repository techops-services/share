import { Hono } from "hono";
import { cors } from "hono/cors";
import { handleCleanup } from "./cleanup";
import { handleDelete } from "./delete";
import { errorResponse } from "./errors";
import { handleHealth } from "./health";
import { handleList } from "./list";
import { handleServe } from "./serve";
import type { Env } from "./types";
import { handleUpload } from "./upload";

const app = new Hono<{ Bindings: Env }>();

app.use(
  "/api/*",
  cors({
    origin: "*",
    allowMethods: ["GET", "POST", "DELETE", "OPTIONS"],
    allowHeaders: ["Content-Type", "X-Api-Key", "X-TTL"],
    exposeHeaders: [
      "X-RateLimit-Limit",
      "X-RateLimit-Remaining",
      "X-RateLimit-Reset",
      "X-Share-Expires",
      "X-Share-Id",
    ],
    maxAge: 86400,
  }),
);

app.get("/api/health", handleHealth);
app.post("/api/upload", handleUpload);
app.get("/api/list", handleList);
app.delete("/api/:id", handleDelete);
app.get("/:id", handleServe);

app.onError((err, c) => {
  if (err instanceof Error && err.message.includes("Body is unusable")) {
    return errorResponse(
      c,
      "INVALID_CONTENT_TYPE",
      "Request body is missing or unreadable",
    );
  }

  return errorResponse(
    c,
    "INTERNAL_ERROR",
    "An unexpected error occurred",
  );
});

app.notFound((c) => {
  return errorResponse(c, "NOT_FOUND", "Endpoint not found");
});

export default {
  fetch: app.fetch,
  async scheduled(
    _event: ScheduledEvent,
    env: Env,
    ctx: ExecutionContext,
  ): Promise<void> {
    ctx.waitUntil(handleCleanup(env));
  },
};
