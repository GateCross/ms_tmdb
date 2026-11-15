import http, { type RequestOptions } from "./http";
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
  timeout: number;
  timeout_restart_required: boolean;
  public_base_url: string;
};

export type AdminProxyPayload = {
  proxy_url?: string;
  local_write_enabled?: boolean;
  timeout?: number;
  public_base_url?: string;
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

export type AdminLogSettingsResp = {
  retention_days: number;
};

export type AdminLogSettingsPayload = {
  retention_days?: number;
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

export type AdminRequestLogListParams = {
  page?: number;
  page_size?: number;
  status?: string;
  keyword?: string;
};

export type AdminRequestLogClearResp = {
  message: string;
};

export type AdminProxyAccessLogItem = {
  id: number;
  request_id: string;
  method: string;
  path: string;
  query: string;
  request_uri: string;
  client_ip: string;
  user_agent: string;
  status_code: number;
  duration_ms: number;
  error_message: string;
  request_body_bytes: number;
  request_body_truncated: boolean;
  response_body_bytes: number;
  response_body_truncated: boolean;
  created_at: string;
};

export type AdminProxyAccessLogDetailResp = AdminProxyAccessLogItem & {
  request_body: string;
  response_body: string;
};

export type AdminProxyAccessLogListResp = {
  total: number;
  page: number;
  page_size: number;
  results: AdminProxyAccessLogItem[];
};

export type AdminTmdbRequestLogItem = {
  id: number;
  request_id: string;
  method: string;
  path: string;
  url: string;
  status_code: number;
  duration_ms: number;
  error_message: string;
  request_body_bytes: number;
  request_body_truncated: boolean;
  response_body_bytes: number;
  response_body_truncated: boolean;
  created_at: string;
};

export type AdminTmdbRequestLogDetailResp = AdminTmdbRequestLogItem & {
  request_body: string;
  response_body: string;
};

export type AdminTmdbRequestLogListResp = {
  total: number;
  page: number;
  page_size: number;
  results: AdminTmdbRequestLogItem[];
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
  // 仅剧集返回；电影恒为空，仅为列表 union 访问兼容
  skip_sync?: boolean;
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
  skip_sync: boolean;
  genre_names: string[];
};

export type AdminHomeMediaItem = {
  media_type: "movie" | "tv";
  tmdb_id: number;
  title: string;
  original_title: string;
  overview: string;
  poster_path: string;
  vote_average: number;
  air_date: string;
  popularity: number;
  visit_count: number;
  created_at: string;
};

export type AdminHomeResp = {
  latest: AdminHomeMediaItem[];
  hot: AdminHomeMediaItem[];
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

export function compareMovieRemote(id: number, options?: RequestOptions) {
  return http.get<AdminCompareResp>(`/api/admin/compare/movie/${id}`, { ...options });
}

export function compareTVRemote(id: number, options?: RequestOptions) {
  return http.get<AdminCompareResp>(`/api/admin/compare/tv/${id}`, { ...options });
}

export function comparePersonRemote(id: number, options?: RequestOptions) {
  return http.get<AdminCompareResp>(`/api/admin/compare/person/${id}`, { ...options });
}

export function getHomeDashboard(limit?: number, options?: RequestOptions) {
  return http.get<AdminHomeResp>("/api/admin/home", {
    params: typeof limit === "number" ? { limit } : undefined,
    ...options,
  });
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

export function listMovies(
  page = 1,
  pageSize = 20,
  keyword = "",
  searchMode = "contains",
  options?: RequestOptions,
) {
  return http.get<AdminListResp<AdminMovieListItem>>("/api/admin/movies", {
    params: { page, page_size: pageSize, keyword, search_mode: searchMode },
    ...options,
  });
}

export function listTV(
  page = 1,
  pageSize = 20,
  keyword = "",
  searchMode = "contains",
  options?: RequestOptions,
) {
  return http.get<AdminListResp<AdminTVListItem>>("/api/admin/tv-series", {
    params: { page, page_size: pageSize, keyword, search_mode: searchMode },
    ...options,
  });
}

export function createTV(payload: AdminCreateTVPayload) {
  return http.post<AdminCreateResp>("/api/admin/tv", payload);
}

export function deleteTV(id: number) {
  return http.delete(`/api/admin/tv/${id}`);
}

export type AdminTvSkipSyncResp = {
  skip_sync: boolean;
  message?: string;
};

export function getTVSkipSync(id: number, options?: RequestOptions) {
  return http.get<AdminTvSkipSyncResp>(`/api/admin/tv/${id}/skip-sync`, {
    ...options,
  });
}

export function updateTVSkipSync(id: number, skipSync: boolean) {
  return http.put<AdminTvSkipSyncResp>(`/api/admin/tv/${id}/skip-sync`, { skip_sync: skipSync });
}

export function getTVSeasonLocal(id: number, seasonNumber: number, options?: RequestOptions) {
  return http.get<AdminTVSeasonLocalResp>(`/api/admin/tv/${id}/season/${seasonNumber}/local`, {
    ...options,
  });
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

export function getProxySettings(options?: RequestOptions) {
  return http.get<AdminProxyResp>("/api/admin/proxy", { ...options });
}

export function updateProxySettings(payload: AdminProxyPayload) {
  return http.put<AdminProxyResp>("/api/admin/proxy", payload);
}

export function getAutoSyncSettings(options?: RequestOptions) {
  return http.get<AdminAutoSyncResp>("/api/admin/auto-sync", { ...options });
}

export function updateAutoSyncSettings(payload: AdminAutoSyncPayload) {
  return http.put<AdminAutoSyncResp>("/api/admin/auto-sync", payload);
}

export function getLogSettings(options?: RequestOptions) {
  return http.get<AdminLogSettingsResp>("/api/admin/log-settings", { ...options });
}

export function updateLogSettings(payload: AdminLogSettingsPayload) {
  return http.put<AdminLogSettingsResp>("/api/admin/log-settings", payload);
}

export function runAutoSyncNow() {
  return http.post<AdminAutoSyncRunResp>("/api/admin/auto-sync/run");
}

export function getAutoSyncLogs(params: AdminAutoSyncLogListParams = {}, options?: RequestOptions) {
  return http.get<AdminAutoSyncLogListResp>("/api/admin/auto-sync/logs", {
    params,
    ...options,
  });
}

export function getAutoSyncLogDetail(
  id: number,
  params: AdminAutoSyncLogDetailParams = {},
  options?: RequestOptions,
) {
  return http.get<AdminAutoSyncLogDetailResp>(`/api/admin/auto-sync/logs/${id}`, {
    params,
    ...options,
  });
}

export function clearAutoSyncLogs() {
  return http.delete<AdminAutoSyncLogClearResp>("/api/admin/auto-sync/logs");
}

export function getProxyAccessLogs(params: AdminRequestLogListParams = {}, options?: RequestOptions) {
  return http.get<AdminProxyAccessLogListResp>("/api/admin/logs/access", {
    params,
    ...options,
  });
}

export function getProxyAccessLogDetail(id: number, options?: RequestOptions) {
  return http.get<AdminProxyAccessLogDetailResp>(`/api/admin/logs/access/${id}`, { ...options });
}

export function clearProxyAccessLogs() {
  return http.delete<AdminRequestLogClearResp>("/api/admin/logs/access");
}

export function getTmdbRequestLogs(params: AdminRequestLogListParams = {}, options?: RequestOptions) {
  return http.get<AdminTmdbRequestLogListResp>("/api/admin/logs/tmdb", {
    params,
    ...options,
  });
}

export function getTmdbRequestLogDetail(id: number, options?: RequestOptions) {
  return http.get<AdminTmdbRequestLogDetailResp>(`/api/admin/logs/tmdb/${id}`, { ...options });
}

export function clearTmdbRequestLogs() {
  return http.delete<AdminRequestLogClearResp>("/api/admin/logs/tmdb");
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
