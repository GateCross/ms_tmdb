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
  message: string;
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

export function listPeople(page = 1, pageSize = 20) {
  return http.get("/api/admin/people", { params: { page, page_size: pageSize } });
}

export function getProxySettings() {
  return http.get<AdminProxyResp>("/api/admin/proxy");
}

export function updateProxySettings(payload: AdminProxyPayload) {
  return http.put<AdminProxyResp>("/api/admin/proxy", payload);
}
