// TMDB 图片 URL 工具
const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p";

// 占位图（无海报/头像时显示）
const PLACEHOLDER_POSTER =
  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='185' height='278' fill='%23e7dcc8'%3E%3Crect width='185' height='278'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' fill='%23999' font-size='14'%3ENo Image%3C/text%3E%3C/svg%3E";
const PLACEHOLDER_PROFILE =
  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='185' height='278' fill='%23e7dcc8'%3E%3Crect width='185' height='278' rx='12'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' fill='%23999' font-size='14'%3ENo Photo%3C/text%3E%3C/svg%3E";

export type ImageSize = "w92" | "w154" | "w185" | "w342" | "w500" | "w780" | "original";

function isDirectImagePath(path: string): boolean {
  return (
    path.startsWith("http://") ||
    path.startsWith("https://") ||
    path.startsWith("data:") ||
    path.startsWith("blob:") ||
    path.startsWith("/uploads/")
  );
}

/**
 * 生成 TMDB 图片完整 URL
 * @param path   poster_path / backdrop_path / profile_path
 * @param size   尺寸，默认 w342
 */
export function tmdbImg(path: string | null | undefined, size: ImageSize = "w342"): string {
  if (!path) return PLACEHOLDER_POSTER;
  if (isDirectImagePath(path)) return path;
  return `${TMDB_IMAGE_BASE}/${size}${path}`;
}

/**
 * 人物头像 URL（fallback 不同）
 */
export function profileImg(path: string | null | undefined, size: ImageSize = "w185"): string {
  if (!path) return PLACEHOLDER_PROFILE;
  if (isDirectImagePath(path)) return path;
  return `${TMDB_IMAGE_BASE}/${size}${path}`;
}
