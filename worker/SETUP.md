# Share Worker - Setup Guide

## Prerequisites

- Node.js 18+
- pnpm (or npm/yarn)
- Cloudflare account with Workers, R2, and KV enabled
- Wrangler CLI (`npm install -g wrangler`)

## 1. Install Dependencies

```bash
cd worker
pnpm install
```

## 2. Authenticate with Cloudflare

```bash
wrangler login
```

## 3. Create R2 Bucket

```bash
wrangler r2 bucket create share-pages
```

## 4. Create KV Namespace

```bash
# Production
wrangler kv namespace create META

# Preview (for local dev)
wrangler kv namespace create META --preview
```

Update `wrangler.toml` with the returned namespace IDs:

```toml
[[kv_namespaces]]
binding = "META"
preview_id = "<preview-id-from-above>"
id = "<production-id-from-above>"
```

## 5. Provision API Keys

API keys follow the format `sk_live_<32-char-hex>`. To create one:

```bash
# Generate a key
API_KEY="sk_live_$(openssl rand -hex 16)"
echo "Your API key: $API_KEY"

# Hash it for storage
KEY_HASH=$(echo -n "$API_KEY" | shasum -a 256 | cut -d ' ' -f 1)

# Store in KV (production)
wrangler kv key put --namespace-id=<your-kv-id> \
  "apikey:$KEY_HASH" \
  '{"key_prefix":"sk_live_'"${API_KEY:8:12}"'","created_at":"'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'","label":"my-first-key"}'
```

## 6. Configure Custom Domain (Optional)

In the Cloudflare dashboard:

1. Go to Workers & Pages > share-worker > Settings > Domains & Routes
2. Add custom domain: `sh.techops.services`
3. Cloudflare will auto-provision the SSL certificate

## 7. Local Development

```bash
pnpm dev
```

The worker runs at `http://localhost:8787`.

### Test upload locally:

```bash
curl -X POST http://localhost:8787/api/upload \
  -H "Content-Type: text/html" \
  -d "<h1>Hello World</h1>"
```

### Health check:

```bash
curl http://localhost:8787/api/health
```

## 8. Run Tests

```bash
pnpm test
```

## 9. Deploy

```bash
pnpm deploy
```

## 10. Verify Production

```bash
# Health check
curl https://sh.techops.services/api/health

# Upload
curl -X POST https://sh.techops.services/api/upload \
  -H "Content-Type: text/html" \
  -d "<h1>Live Test</h1>"
```

## Environment Variables

All environment variables are configured in `wrangler.toml` under `[vars]`. No secrets are needed beyond the KV-stored API keys.

| Variable | Default | Description |
|---|---|---|
| BASE_URL | https://sh.techops.services | Public base URL |
| MAX_UPLOAD_SIZE | 5242880 | Max upload size in bytes (5MB) |
| DEFAULT_TTL | 86400 | Default TTL in seconds (24h) |
| MAX_TTL_ANONYMOUS | 86400 | Max TTL for anonymous uploads (24h) |
| MAX_TTL_AUTHENTICATED | 2592000 | Max TTL for authenticated uploads (30d) |
| MIN_TTL | 300 | Minimum TTL (5 minutes) |
| ANON_RATE_LIMIT | 10 | Anonymous uploads per hour |
| AUTH_RATE_LIMIT | 100 | Authenticated uploads per hour |

## Cron Schedule

The cleanup cron runs every hour (`0 * * * *`) to remove expired shares from R2 and write expired markers to KV for 410 responses.
