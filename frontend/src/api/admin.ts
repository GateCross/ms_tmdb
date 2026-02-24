import http from "./http";

export type AdminSyncMode = "overwrite_all" | "update_unmodified" | "selective" | "preview";

export type AdminSyncPayload = {
  mode?: AdminSyncMode;
  overwrite_fields?: string[];
};

export type AdminSyncResp = {
  mode: AdminSyncMode;
  changed_fields: string[];
  overwritten_fields: string[];
  kept_local_fields: string[];
  is_modified: boolean;
  message: string;
};

export type AdminProxyResp = {
  proxy_url: string;
  enabled: boolean;
};

export type AdminProxyPayload = {
  proxy_url?: string;
};

export type AdminCompareResp = {
  has_diff: boolean;
  diff_fields: string[];
  local_override_diff_fields: string[];
  diff_details: AdminCompareFieldDetail[];
  message: string;
};

export type AdminCompareFieldDetail = {
  field: string;
  diff_type: "remote" | "local_override" | string;
  local: string;
  remote: string;
};

export type AdminCreateResp = {
  tmdb_id: number;
  is_local: boolean;
  message: string;
};

export type AdminCreateMoviePayload = {
  title: string;
  original_title?: string;
  genre_names?: string[];
  tagline?: string;
  release_date?: string;
  status?: string;
  runtime?: number;
  original_language?: string;
  homepage?: string;
  poster_path?: string;
  backdrop_path?: string;
  vote_average?: number;
  popularity?: number;
  overview?: string;
};

export type AdminCreateTVPayload = {
  name: string;
  original_name?: string;
  genre_names?: string[];
  type?: string;
  tagline?: string;
  first_air_date?: string;
  status?: string;
  number_of_seasons?: number;
  number_of_episodes?: number;
  original_language?: string;
  homepage?: string;
  poster_path?: string;
  backdrop_path?: string;
  vote_average?: number;
  popularity?: number;
  overview?: string;
};

export type AdminUploadResp = {
  path: string;
};

export function getStats() {
  return http.get("/api/admin/stats");
}

export function syncMovie(id: number, payload: AdminSyncPayload = {}) {
  return http.post<AdminSyncResp>(`/api/admin/sync/movie/${id}`, payload);
}

export function syncTV(id: number, payload: AdminSyncPayload = {}) {
  return http.post<AdminSyncResp>(`/api/admin/sync/tv/${id}`, payload);
}

export function syncPerson(id: number, payload: AdminSyncPayload = {}) {
  return http.post<AdminSyncResp>(`/api/admin/sync/person/${id}`, payload);
}

export function compareMovieRemote(id: number) {
  return http.get<AdminCompareResp>(`/api/admin/compare/movie/${id}`);
}

export function compareTVRemote(id: number) {
  return http.get<AdminCompareResp>(`/api/admin/compare/tv/${id}`);
}

export function comparePersonRemote(id: number) {
  return http.get<AdminCompareResp>(`/api/admin/compare/person/${id}`);
}

export function updateMovie(id: number, payload: Record<string, unknown>) {
  return http.put(`/api/admin/movie/${id}`, payload);
}

export function createMovie(payload: AdminCreateMoviePayload) {
  return http.post<AdminCreateResp>("/api/admin/movie", payload);
}

export function deleteMovie(id: number) {
  return http.delete(`/api/admin/movie/${id}`);
}

export function updateTV(id: number, payload: Record<string, unknown>) {
  return http.put(`/api/admin/tv/${id}`, payload).catch((err: any) => {
    if (String(err?.message ?? "").includes("404")) {
      return http.put(`/api/admin/tv-series/${id}`, payload);
    }
    throw err;
  });
}

export function listMovies(page = 1, pageSize = 20, keyword = "", searchMode = "contains") {
  return http.get("/api/admin/movies", { params: { page, page_size: pageSize, keyword, search_mode: searchMode } });
}

export function listTV(page = 1, pageSize = 20, keyword = "", searchMode = "contains") {
  return http
    .get("/api/admin/tv-series", { params: { page, page_size: pageSize, keyword, search_mode: searchMode } })
    .catch((err: any) => {
      if (String(err?.message ?? "").includes("404")) {
        return http.get("/api/admin/tv", { params: { page, page_size: pageSize, keyword, search_mode: searchMode } });
      }
      throw err;
    });
}

export function createTV(payload: AdminCreateTVPayload) {
  return http.post<AdminCreateResp>("/api/admin/tv", payload);
}

export function deleteTV(id: number) {
  return http.delete(`/api/admin/tv/${id}`);
}

export function listPeople(page = 1, pageSize = 20) {
  return http.get("/api/admin/people", { params: { page, page_size: pageSize } });
}

export function getProxySettings() {
  return http.get<AdminProxyResp>("/api/admin/proxy");
}

export function updateProxySettings(payload: AdminProxyPayload) {
  return http.put<AdminProxyResp>("/api/admin/proxy", payload);
}

export function uploadAdminImage(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  return http.post<AdminUploadResp>("/api/admin/upload/image", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}
