import http from "./http";
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

export function getPopularMovies(page = 1, language = "zh-CN") {
  return http.get<PagedResp<MediaSummary>>("/api/v3/movie/popular", { params: { page, language } });
}

export function getMovieDetail(id: number, language = "zh-CN", append = "credits,videos,images") {
  return http.get<MediaSummary & Record<string, unknown>>(`/api/v3/movie/${id}`, {
    params: { language, append_to_response: append },
  });
}

export function getMovieGenreList(language = "zh-CN") {
  return http.get<GenreListResp>("/api/v3/genre/movie/list", { params: { language } });
}
