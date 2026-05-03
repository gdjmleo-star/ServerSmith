// Frontend timezone utilities for ServerSmith.
// All timestamps from API are in UTC ISO8601 format.
// Display must always be in Asia/Shanghai (Beijing time).

const BEIJING_TZ = "Asia/Shanghai";

/**
 * Format a UTC ISO8601 string to Beijing time display string.
 * Example: "2026-05-03T14:00:00Z" → "2026-05-03 22:00:00"
 */
export function formatUTCToBeijing(utcStr: string | null | undefined): string {
  if (!utcStr) return "-";
  try {
    const date = new Date(utcStr);
    return date.toLocaleString("zh-CN", {
      timeZone: BEIJING_TZ,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    });
  } catch {
    return utcStr;
  }
}

/**
 * Format UTC to Beijing date only (no time).
 * Example: "2026-05-02T15:59:59Z" → "2026-05-03"
 */
export function formatUTCToBeijingDate(utcStr: string | null | undefined): string {
  if (!utcStr) return "-";
  try {
    const date = new Date(utcStr);
    return date.toLocaleDateString("zh-CN", {
      timeZone: BEIJING_TZ,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  } catch {
    return utcStr;
  }
}

/**
 * Format UTC to relative time (e.g. "2分钟前", "3小时前").
 */
export function timeAgo(utcStr: string | null | undefined): string {
  if (!utcStr) return "-";
  try {
    const now = Date.now();
    const then = new Date(utcStr).getTime();
    const diffSec = Math.floor((now - then) / 1000);

    if (diffSec < 0) return "刚刚";
    if (diffSec < 60) return `${diffSec}秒前`;
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}分钟前`;
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}小时前`;
    return formatUTCToBeijingDate(utcStr);
  } catch {
    return utcStr;
  }
}
