export const PAGE_SIZE = 20;
export const DEFAULT_HISTORY_DAYS = 30;
// In production (Go embed), frontend and backend are served from the same origin,
// so API calls use relative paths. Set NEXT_PUBLIC_API_URL for cross-origin dev.
export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";
export const API_TIMEOUT_MS = 10000;
export const TOPEMBY_URL = process.env.NEXT_PUBLIC_TOPEMBY_URL || "http://localhost:3000";
export const ALERT_POLL_MS = 30000;
