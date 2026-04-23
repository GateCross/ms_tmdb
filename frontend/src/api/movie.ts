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

export function getPopularMovies(page = 1, language = "zh-CN") {
  return http.get<PagedResp<MediaSummary>>("/api/v3/movie/popular", { params: { page, language } });
}

export function getMovieDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const params = append ? { language, append_to_response: append } : { language };
  const key = `movie:detail:${id}:${language}:${append}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get<MediaSummary & Record<string, unknown>>(`/api/v3/movie/${id}`, { params }),
    DETAIL_CACHE_TTL,
  );
}

export function getMovieGenreList(language = "zh-CN") {
  return withRequestCache(
    `genre:movie:${language}`,
    () => http.get<GenreListResp>("/api/v3/genre/movie/list", { params: { language } }),
    GENRE_CACHE_TTL,
  );
}

export function getMovieCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const key = `movie:credits:${id}:${language}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get<Record<string, unknown>>(`/api/v3/movie/${id}/credits`, { params: { language } }),
    AUXILIARY_CACHE_TTL,
  );
}
