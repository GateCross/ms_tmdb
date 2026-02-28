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

export function getPopularTV(page = 1, language = "zh-CN") {
  return http.get<PagedResp<MediaSummary>>("/api/v3/tv/popular", { params: { page, language } });
}

export function getTVDetail(id: number, language = "zh-CN", append = "credits,videos,images") {
  return http.get<MediaSummary & Record<string, unknown>>(`/api/v3/tv/${id}`, {
    params: { language, append_to_response: append },
  });
}

export function getTVGenreList(language = "zh-CN") {
  return http.get<GenreListResp>("/api/v3/genre/tv/list", { params: { language } });
}

export function getTVSeasonDetail(
  id: number,
  seasonNumber: number,
  language = "zh-CN",
  append = "credits,images,videos",
) {
  return http.get<Record<string, unknown>>(`/api/v3/tv/${id}/season/${seasonNumber}`, {
    params: { language, append_to_response: append },
  });
}
