import type { Context } from "hono";
import type { Env, HealthResponse } from "./types";

const VERSION = "1.0.0";

export async function handleHealth(
  c: Context<{ Bindings: Env }>,
): Promise<Response> {
  const body: HealthResponse = {
    status: "ok",
    version: VERSION,
  };
  return c.json(body, 200);
}
