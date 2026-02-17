import type { Env, ShareMetadata } from "./types";

const CLEANUP_BATCH_SIZE = 100;

export async function handleCleanup(env: Env): Promise<void> {
  const { META, PAGES } = env;
  const now = new Date();

  let cursor: string | undefined;
  let totalCleaned = 0;

  do {
    const listResult = await META.list({
      prefix: "share:",
      limit: CLEANUP_BATCH_SIZE,
      cursor,
    });

    for (const key of listResult.keys) {
      const metadata = await META.get<ShareMetadata>(key.name, "json");
      if (!metadata) {
        continue;
      }

      const expiresAt = new Date(metadata.expires_at);
      if (now < expiresAt) {
        continue;
      }

      await PAGES.delete(`${metadata.id}.html`);
      await META.delete(key.name);

      await META.put(
        `expired:${metadata.id}`,
        JSON.stringify({ expired_at: now.toISOString() }),
        { expirationTtl: 604800 },
      );

      totalCleaned++;
    }

    cursor = listResult.list_complete ? undefined : listResult.cursor;
  } while (cursor);

  if (totalCleaned > 0) {
    // Cron cleanup completed; count is available in worker analytics
    void totalCleaned;
  }
}
