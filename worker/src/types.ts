export interface Env {
  PAGES: R2Bucket;
  META: KVNamespace;
  BASE_URL: string;
  MAX_UPLOAD_SIZE: string;
  DEFAULT_TTL: string;
  MAX_TTL_ANONYMOUS: string;
  MAX_TTL_AUTHENTICATED: string;
  MIN_TTL: string;
  ANON_RATE_LIMIT: string;
  AUTH_RATE_LIMIT: string;
}

export interface ShareMetadata {
  id: string;
  size: number;
  ttl: number;
  created_at: string;
  expires_at: string;
  content_type: string;
  api_key_hash: string | null;
  views: number;
}

export interface ApiKeyRecord {
  key_prefix: string;
  created_at: string;
  label: string;
}

export interface RateLimitEntry {
  count: number;
}

export interface ExpiredMarker {
  expired_at: string;
}

export interface UploadResponse {
  id: string;
  url: string;
  expires_at: string;
  size: number;
  ttl: number;
}

export interface ListResponse {
  shares: ShareMetadata[];
  cursor: string | null;
  total: number;
}

export interface HealthResponse {
  status: "ok";
  version: string;
}

export type ErrorCode =
  | "INVALID_CONTENT_TYPE"
  | "PAYLOAD_TOO_LARGE"
  | "RATE_LIMITED"
  | "NOT_FOUND"
  | "GONE"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "INTERNAL_ERROR"
  | "INVALID_TTL";

export interface ErrorResponse {
  error: {
    code: ErrorCode;
    message: string;
    retry_after?: number;
  };
}

export interface RateLimitResult {
  allowed: boolean;
  limit: number;
  remaining: number;
  reset: number;
  retryAfter?: number;
}
