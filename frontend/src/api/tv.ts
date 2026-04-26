import http from "./http";
import { clearRequestCache, withRequestCache } from "./requestCache";
import type { MediaGenre, MediaSummary } from "@/types/media";

type PagedResp<T> = {
  page: number;
  total_pages?: number;
  total_results?: number;
  results: T[];
};

type GenreListResp = {
  genres: MediaGenre[];
};

const GENRE_CACHE_TTL = 30 * 60 * 1000;
const AUXILIARY_CACHE_TTL = 10 * 60 * 1000;
const DETAIL_CACHE_TTL = 5 * 60 * 1000;

type DetailOptions = {
  force?: boolean;
};

export function getPopularTV(page = 1, language = "zh-CN") {
  return http.get<PagedResp<MediaSummary>>("/api/v3/tv/popular", { params: { page, language } });
}

export function getTVDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const params = append ? { language, append_to_response: append } : { language };
  const key = `tv:detail:${id}:${language}:${append}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get<MediaSummary & Record<string, unknown>>(`/api/v3/tv/${id}`, { params }),
    DETAIL_CACHE_TTL,
  );
}

export function getTVGenreList(language = "zh-CN") {
  return withRequestCache(
    `genre:tv:${language}`,
    () => http.get<GenreListResp>("/api/v3/genre/tv/list", { params: { language } }),
    GENRE_CACHE_TTL,
  );
}

export function getTVCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const key = `tv:credits:${id}:${language}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get<Record<string, unknown>>(`/api/v3/tv/${id}/credits`, { params: { language } }),
    AUXILIARY_CACHE_TTL,
  );
}

export function getTVSeasonDetail(id: number, seasonNumber: number, language = "zh-CN", append = "") {
  const params = append ? { language, append_to_response: append } : { language };
  return http.get<Record<string, unknown>>(`/api/v3/tv/${id}/season/${seasonNumber}`, {
    params,
  });
}
