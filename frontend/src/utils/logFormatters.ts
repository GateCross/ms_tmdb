import type {
  AdminAutoSyncLogDetailEntry,
  AdminProxyAccessLogDetailResp,
  AdminTmdbRequestLogDetailResp,
} from "@/api/admin";
import { formatJsonText } from "@/utils/jsonText";

export type RequestLogDetail = AdminProxyAccessLogDetailResp | AdminTmdbRequestLogDetailResp;

export function formatRequestLogTotal(total: number) {
  return (Math.max(0, Number(total) || 0)).toLocaleString("zh-CN");
}

export function formatDateTime(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "-";
  }
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) {
    return text;
  }
  return date.toLocaleString("zh-CN", { hour12: false });
}

/* 列表时间两行展示：日期弱化、时刻突出 */
export function formatDateTimeParts(value: string) {
  const [date = "-", time = ""] = formatDateTime(value).split(" ");
  return { date, time: time || date };
}

export function formatDuration(durationMs: number) {
  const ms = Number.isFinite(durationMs) ? Math.max(0, Math.trunc(durationMs)) : 0;
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const seconds = ms / 1000;
  return seconds < 60
    ? `${seconds.toFixed(seconds < 10 ? 1 : 0)}s`
    : `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`;
}

export function formatBytes(bytes: number) {
  const size = Number.isFinite(bytes) ? Math.max(0, bytes) : 0;
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(size < 10 * 1024 ? 1 : 0)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

export function formatStatusCode(code: number) {
  return code > 0 ? String(code) : "未返回";
}

export function statusClass(code: number) {
  if (code >= 200 && code < 400) {
    return "bg-success-soft text-success border border-success-line";
  }
  if (code === 0 || code >= 400) {
    return "bg-danger-soft text-danger border border-danger-line";
  }
  return "bg-warn-soft text-warn border border-warn-line";
}

/* 列表状态用色点表达，弱化色块胶囊 */
export function statusDotClass(code: number) {
  if (code >= 200 && code < 400) {
    return "log-status-dot-success";
  }
  if (code === 0 || code >= 400) {
    return "log-status-dot-danger";
  }
  return "log-status-dot-warn";
}

export function formatMode(mode: string) {
  return mode === "overwrite_all" ? "全量覆盖" : "仅更新未在本地修改的字段";
}

export function formatAutoSyncStatus(status: string) {
  switch (status) {
    case "success":
      return "成功";
    case "partial_failed":
      return "部分失败";
    case "panic":
      return "异常";
    case "canceled":
      return "已取消";
    default:
      return status || "-";
  }
}

export function autoSyncStatusClass(status: string) {
  switch (status) {
    case "success":
      return "bg-success-soft text-success border border-success-line";
    case "partial_failed":
      return "bg-warn-soft text-warn border border-warn-line";
    case "panic":
      return "bg-danger-soft text-danger border border-danger-line";
    case "canceled":
      return "bg-raised text-muted border border-line";
    default:
      return "bg-raised text-muted border border-line";
  }
}

export function autoSyncStatusDotClass(status: string) {
  switch (status) {
    case "success":
      return "log-status-dot-success";
    case "partial_failed":
      return "log-status-dot-warn";
    case "panic":
      return "log-status-dot-danger";
    default:
      return "log-status-dot-muted";
  }
}

export function summarizeMessage(message: string) {
  const text = (message ?? "").trim();
  if (!text) {
    return "-";
  }
  if (text.length <= 26) {
    return text;
  }
  return `${text.slice(0, 26)}...`;
}

export function formatMediaType(mediaType: string) {
  switch (mediaType) {
    case "movie":
      return "电影";
    case "tv":
      return "剧集";
    case "person":
      return "人物";
    default:
      return mediaType || "-";
  }
}

export function formatFieldList(fields: string[] | undefined) {
  if (!Array.isArray(fields) || fields.length === 0) {
    return "-";
  }
  return fields.join("、");
}

export function visibleFieldList(fields: string[] | undefined) {
  return Array.isArray(fields) ? fields.filter((field) => !!field) : [];
}

export function hasFieldList(fields: string[] | undefined) {
  return visibleFieldList(fields).length > 0;
}

export function hasLocalFieldSummary(entry: Pick<AdminAutoSyncLogDetailEntry, "changed_fields" | "overwritten_fields" | "kept_local_fields">) {
  return hasFieldList(entry.changed_fields) || hasFieldList(entry.overwritten_fields) || hasFieldList(entry.kept_local_fields);
}

export function fieldChangeCount(changes: Array<{ field: string; diff_type: string; before: string; after: string }> | undefined) {
  return Array.isArray(changes) ? changes.length : 0;
}

export function formatFieldChanges(
  changes: Array<{ field: string; diff_type: string; before: string; after: string }> | undefined,
) {
  if (!Array.isArray(changes) || changes.length === 0) {
    return "-";
  }
  return changes
    .map((item) => `${item.field} [${item.diff_type || "remote"}]\n前: ${item.before || "-"}\n后: ${item.after || "-"}`)
    .join("\n\n");
}

export function trimMiddle(value: string, max = 72) {
  const text = (value ?? "").trim();
  if (!text) return "-";
  if (text.length <= max) return text;
  const head = Math.ceil((max - 3) * 0.6);
  const tail = Math.floor((max - 3) * 0.4);
  return `${text.slice(0, head)}...${text.slice(text.length - tail)}`;
}

export function splitPathAndQuery(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return { path: "-", query: "" };
  }
  const index = text.indexOf("?");
  if (index < 0) {
    return { path: text, query: "" };
  }
  return {
    path: text.slice(0, index) || "/",
    query: text.slice(index + 1),
  };
}

export function accessPath(value: string) {
  return splitPathAndQuery(value).path;
}

export function formatRequestUriForDisplay(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "-";
  }
  try {
    return decodeURIComponent(text);
  } catch {
    return text;
  }
}

export function formatQueryForDisplay(query: string, max: number) {
  const text = (query ?? "").trim().replace(/^\?/, "");
  if (!text) {
    return "无查询参数";
  }
  const params = new window.URLSearchParams(text);
  for (const key of [...params.keys()]) {
    if (key.toLowerCase() === "api_key") {
      params.delete(key);
    }
  }
  const safeQuery = [...params.entries()].map(([key, value]) => `${key}=${value}`).join("&");
  return safeQuery ? `?${trimMiddle(safeQuery, max)}` : "无查询参数";
}

export function accessQuery(value: string) {
  const query = splitPathAndQuery(value).query;
  return formatQueryForDisplay(query, 96);
}

export function upstreamHost(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "-";
  }
  try {
    return new URL(text).origin;
  } catch {
    return "-";
  }
}

export function upstreamQuery(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "无查询参数";
  }
  try {
    const parsed = new URL(text);
    return formatQueryForDisplay(parsed.search, 110);
  } catch {
    return accessQuery(text);
  }
}

/* 完整脱敏查询串：供 tooltip 展示，不做长度截断 */
export function upstreamQueryFull(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "";
  }
  try {
    const params = new URLSearchParams(new URL(text).search);
    params.delete("api_key");
    const safe = [...params.entries()].map(([k, v]) => `${k}=${v}`).join("&");
    return safe ? `?${safe}` : "";
  } catch {
    return "";
  }
}

export function accessQueryFull(value: string) {
  const query = splitPathAndQuery(value).query;
  const text = query.replace(/^\?/, "");
  if (!text) {
    return "";
  }
  const params = new URLSearchParams(text);
  params.delete("api_key");
  const safe = [...params.entries()].map(([k, v]) => `${k}=${v}`).join("&");
  return safe ? `?${safe}` : "";
}

export function bodyText(text: string | undefined) {
  return formatJsonText(text);
}

export function bodyMeta(bytes: number, truncated: boolean) {
  return `${formatBytes(bytes)}${truncated ? " · 已截断" : ""}`;
}

export function isAccessDetail(detail: RequestLogDetail): detail is AdminProxyAccessLogDetailResp {
  return "request_uri" in detail;
}

export function detailEndpointPath(detail: RequestLogDetail) {
  if (isAccessDetail(detail)) {
    return detail.path || accessPath(detail.request_uri);
  }
  return detail.path || "-";
}

export function detailEndpointDisplay(detail: RequestLogDetail) {
  const path = detailEndpointPath(detail);
  const query = detailEndpointQuery(detail);
  return query === "无查询参数" ? path : `${path}${query}`;
}

export function detailEndpointQuery(detail: RequestLogDetail) {
  if (isAccessDetail(detail)) {
    return detail.query ? formatQueryForDisplay(detail.query, 96) : accessQuery(detail.request_uri);
  }
  return upstreamQuery(detail.url);
}

export function detailEndpointSource(detail: RequestLogDetail) {
  if (isAccessDetail(detail)) {
    return detail.client_ip || "-";
  }
  return upstreamHost(detail.url);
}

export function detailEndpointTitle(detail: RequestLogDetail) {
  return detailEndpointDisplay(detail);
}
