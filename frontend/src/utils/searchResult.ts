import { profileImg, tmdbImg } from "@/api/tmdb";
import type { SearchType } from "@/api/search";
import type { SearchResultItem } from "@/types/media";

const MEDIA_TYPE_LABELS: Record<string, string> = {
  movie: "电影",
  tv: "剧集",
  person: "人物",
};

export function getSearchResultMediaType(item: SearchResultItem, fallbackType: SearchType) {
  return item.media_type ?? fallbackType;
}

export function getSearchResultKey(item: SearchResultItem, fallbackType: SearchType) {
  return `${getSearchResultMediaType(item, fallbackType)}-${item.id}`;
}

export function getSearchResultRoute(item: SearchResultItem, fallbackType: SearchType) {
  const mediaType = getSearchResultMediaType(item, fallbackType);
  if (mediaType === "movie") return `/movie/${item.id}`;
  if (mediaType === "tv") return `/tv/${item.id}`;
  if (mediaType === "person") return `/person/${item.id}`;
  return "/search";
}

export function getSearchResultThumb(item: SearchResultItem, fallbackType: SearchType) {
  const mediaType = getSearchResultMediaType(item, fallbackType);
  if (mediaType === "person") return profileImg(item.profile_path, "w92");
  return tmdbImg(item.poster_path, "w92");
}

export function getSearchResultTitle(item: SearchResultItem) {
  return item.title || item.name || item.original_title || `ID ${item.id}`;
}

export function getSearchResultSubtitle(item: SearchResultItem, fallbackType: SearchType) {
  const mediaType = getSearchResultMediaType(item, fallbackType);
  const tag = MEDIA_TYPE_LABELS[mediaType] ?? mediaType;
  const date = item.release_date || item.first_air_date || "";
  return date ? `${tag} · ${date}` : tag;
}
