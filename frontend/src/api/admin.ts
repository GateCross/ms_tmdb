import http from "./http";
import type { MediaSummary } from "@/types/media";

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
  local_write_enabled: boolean;
};

export type AdminProxyPayload = {
  proxy_url?: string;
  local_write_enabled?: boolean;
};

export type AdminAutoSyncMode = "overwrite_all" | "update_unmodified";

export type AdminAutoSyncResp = {
  enabled: boolean;
  cron_expr: string;
  mode: AdminAutoSyncMode | string;
  batch_size: number;
  start_delay_second: number;
  running: boolean;
};

export type AdminAutoSyncPayload = {
  enabled?: boolean;
  cron_expr?: string;
  mode?: AdminAutoSyncMode;
  batch_size?: number;
  start_delay_second?: number;
};

export type AdminAutoSyncRunResp = {
  started: boolean;
  running: boolean;
  message: string;
};

export type AdminAutoSyncLogClearResp = {
  message: string;
};

export type AdminAutoSyncLogItem = {
  id: number;
  triggered_at: string;
  cron_expr: string;
  mode: string;
  batch_size: number;
  status: string;
  checked: number;
  synced: number;
  failed: number;
  duration_ms: number;
  message: string;
  started_at: string;
  finished_at: string;
  created_at: string;
};

export type AdminAutoSyncLogDetailEntry = {
  media_type: string;
  tmdb_id: number;
  name: string;
  message: string;
  remote_diff_fields: string[];
  field_changes: AdminAutoSyncLogFieldChange[];
  changed_fields: string[];
  overwritten_fields: string[];
  kept_local_fields: string[];
};

export type AdminAutoSyncLogFieldChange = {
  field: string;
  diff_type: string;
  before: string;
  after: string;
};

export type AdminAutoSyncLogDetailResp = {
  id: number;
  triggered_at: string;
  cron_expr: string;
  mode: string;
  batch_size: number;
  status: string;
  checked: number;
  synced: number;
  failed: number;
  duration_ms: number;
  message: string;
  started_at: string;
  finished_at: string;
  created_at: string;
  synced_page: number;
  synced_page_size: number;
  synced_list: AdminAutoSyncLogDetailEntry[];
  failed_page: number;
  failed_page_size: number;
  failed_list: AdminAutoSyncLogDetailEntry[];
};

export type AdminAutoSyncLogDetailParams = {
  synced_page?: number;
  synced_page_size?: number;
  failed_page?: number;
  failed_page_size?: number;
};

export type AdminAutoSyncLogListResp = {
  total: number;
  page: number;
  page_size: number;
  results: AdminAutoSyncLogItem[];
};

export type AdminAutoSyncLogListParams = {
  page?: number;
  page_size?: number;
  status?: string;
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
  sync_tmdb_id?: number;
  is_local: boolean;
  message: string;
};

export type AdminTVSeasonLocalResp = {
  saved: boolean;
  data: Record<string, unknown> | null;
  message?: string;
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

export type AdminMovieListItem = MediaSummary & {
  tmdb_id: number;
  title: string;
  original_title: string;
  poster_path: string;
  vote_average: number;
  release_date: string;
  popularity: number;
  is_modified: boolean;
  genre_names: string[];
};

export type AdminTVListItem = MediaSummary & {
  tmdb_id: number;
  name: string;
  original_name: string;
  poster_path: string;
  vote_average: number;
  first_air_date: string;
  number_of_seasons: number;
  number_of_episodes: number;
  popularity: number;
  is_modified: boolean;
  genre_names: string[];
};

export type AdminListResp<T> = {
  total: number;
  page: number;
  page_size: number;
  results: T[];
};

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
	return http.put(`/api/admin/tv/${id}`, payload);
}

export function listMovies(page = 1, pageSize = 20, keyword = "", searchMode = "contains") {
  return http.get<AdminListResp<AdminMovieListItem>>("/api/admin/movies", {
    params: { page, page_size: pageSize, keyword, search_mode: searchMode },
  });
}

export function listTV(page = 1, pageSize = 20, keyword = "", searchMode = "contains") {
	return http.get<AdminListResp<AdminTVListItem>>("/api/admin/tv-series", {
		params: { page, page_size: pageSize, keyword, search_mode: searchMode },
	});
}

export function createTV(payload: AdminCreateTVPayload) {
  return http.post<AdminCreateResp>("/api/admin/tv", payload);
}

export function deleteTV(id: number) {
  return http.delete(`/api/admin/tv/${id}`);
}

export function getTVSeasonLocal(id: number, seasonNumber: number) {
  return http.get<AdminTVSeasonLocalResp>(`/api/admin/tv/${id}/season/${seasonNumber}/local`);
}

export function saveTVSeasonLocal(id: number, seasonNumber: number, language = "zh-CN") {
  return http.post<AdminTVSeasonLocalResp>(`/api/admin/tv/${id}/season/${seasonNumber}/local`, null, {
    params: { language },
  });
}

export function updateTVSeasonLocal(id: number, seasonNumber: number, payload: Record<string, unknown>) {
  return http.put<AdminTVSeasonLocalResp>(`/api/admin/tv/${id}/season/${seasonNumber}/local`, { payload });
}

export function deleteTVSeasonLocal(id: number, seasonNumber: number) {
  return http.delete<AdminTVSeasonLocalResp>(`/api/admin/tv/${id}/season/${seasonNumber}/local`);
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

export function getAutoSyncSettings() {
  return http.get<AdminAutoSyncResp>("/api/admin/auto-sync");
}

export function updateAutoSyncSettings(payload: AdminAutoSyncPayload) {
  return http.put<AdminAutoSyncResp>("/api/admin/auto-sync", payload);
}

export function runAutoSyncNow() {
  return http.post<AdminAutoSyncRunResp>("/api/admin/auto-sync/run");
}

export function getAutoSyncLogs(params: AdminAutoSyncLogListParams = {}) {
  return http.get<AdminAutoSyncLogListResp>("/api/admin/auto-sync/logs", { params });
}

export function getAutoSyncLogDetail(id: number, params: AdminAutoSyncLogDetailParams = {}) {
  return http.get<AdminAutoSyncLogDetailResp>(`/api/admin/auto-sync/logs/${id}`, { params });
}

export function clearAutoSyncLogs() {
  return http.delete<AdminAutoSyncLogClearResp>("/api/admin/auto-sync/logs");
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
