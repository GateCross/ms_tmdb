import http from "./http";

export function getPopularTV(page = 1, language = "zh-CN") {
  return http.get("/api/v3/tv/popular", { params: { page, language } });
}

export function getTVDetail(id: number, language = "zh-CN", append = "credits,videos,images") {
  return http.get(`/api/v3/tv/${id}`, { params: { language, append_to_response: append } });
}

export function getTVGenreList(language = "zh-CN") {
  return http.get("/api/v3/genre/tv/list", { params: { language } });
}

export function getTVSeasonDetail(
  id: number,
  seasonNumber: number,
  language = "zh-CN",
  append = "credits,images,videos",
) {
  return http.get(`/api/v3/tv/${id}/season/${seasonNumber}`, {
    params: { language, append_to_response: append },
  });
}
