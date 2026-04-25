<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import {
  clearAutoSyncLogs,
  getAutoSyncLogDetail,
  getAutoSyncLogs,
  getAutoSyncSettings,
  getProxySettings,
  runAutoSyncNow,
  updateAutoSyncSettings,
  updateProxySettings,
  type AdminAutoSyncLogDetailParams,
  type AdminAutoSyncLogDetailResp,
  type AdminAutoSyncLogItem,
  type AdminAutoSyncMode,
} from "@/api/admin";

const loading = ref(false);

const proxySaving = ref(false);
const proxyError = ref("");
const proxyMessage = ref("");
const proxyEnabled = ref(false);
const proxyURL = ref("");
const proxyLocalWriteEnabled = ref(true);

const syncSaving = ref(false);
const syncError = ref("");
const syncMessage = ref("");
const syncEnabled = ref(true);
const syncCronExpr = ref("*/30 * * * *");
const syncMode = ref<AdminAutoSyncMode>("update_unmodified");
const syncBatchSize = ref(50);
const syncStartDelaySecond = ref(15);
const syncRunning = ref(false);
const syncTriggering = ref(false);

const logsLoading = ref(false);
const logsClearing = ref(false);
const clearLogsConfirmVisible = ref(false);
const logsError = ref("");
const logsMessage = ref("");
const logsStatus = ref("");
const logsPage = ref(1);
const logsPageSize = ref(10);
const logsTotal = ref(0);
const logsItems = ref<AdminAutoSyncLogItem[]>([]);
const detailModalVisible = ref(false);
const detailLoading = ref(false);
const detailError = ref("");
const activeLogDetail = ref<AdminAutoSyncLogDetailResp | null>(null);
const activeLogId = ref<number | null>(null);
const detailSyncedPage = ref(1);
const detailSyncedPageSize = ref(10);
const detailFailedPage = ref(1);
const detailFailedPageSize = ref(10);

const modeOptions: Array<{ label: string; value: AdminAutoSyncMode; hint: string }> = [
  { label: "仅更新未在本地修改的字段", value: "update_unmodified", hint: "保留本地改动，只更新 TMDB 远端变化字段" },
  { label: "全量覆盖", value: "overwrite_all", hint: "使用 TMDB 最新数据覆盖本地字段" },
];

const logStatusOptions: Array<{ label: string; value: string }> = [
  { label: "全部状态", value: "" },
  { label: "成功", value: "success" },
  { label: "部分失败", value: "partial_failed" },
  { label: "异常", value: "panic" },
];

const settingsBusy = computed(() => loading.value || proxySaving.value || syncSaving.value || syncTriggering.value || logsLoading.value || logsClearing.value);
const proxyStatusText = computed(() => (proxyEnabled.value ? "已启用" : "直连"));
const proxyLocalWriteStatusText = computed(() => (proxyLocalWriteEnabled.value ? "自动写入本地" : "仅读已有本地"));
const syncStatusText = computed(() => (syncEnabled.value ? "已启用" : "已关闭"));
const taskRunStatusText = computed(() => (syncRunning.value ? "执行中" : "空闲"));
const latestLog = computed(() => logsItems.value[0] ?? null);
const latestLogStatusText = computed(() => (latestLog.value ? formatStatus(latestLog.value.status) : "暂无记录"));
const latestLogTimeText = computed(() => (latestLog.value ? formatDateTime(latestLog.value.triggered_at) : "等待首次执行"));

function normalizeProxyURL(raw: string) {
  return raw.trim();
}

function normalizeNumber(value: number, min: number, max: number) {
  const next = Number.isFinite(value) ? Math.trunc(value) : min;
  if (next < min) return min;
  if (next > max) return max;
  return next;
}

function formatMode(mode: string) {
  return mode === "overwrite_all" ? "全量覆盖" : "仅更新未在本地修改的字段";
}

function formatStatus(status: string) {
  switch (status) {
    case "success":
      return "成功";
    case "partial_failed":
      return "部分失败";
    case "panic":
      return "异常";
    default:
      return status || "-";
  }
}

function statusClass(status: string) {
  switch (status) {
    case "success":
      return "bg-green-50 text-green-700 border border-green-200";
    case "partial_failed":
      return "bg-amber-50 text-amber-700 border border-amber-200";
    case "panic":
      return "bg-red-50 text-red-700 border border-red-200";
    default:
      return "bg-gray-50 text-gray-600 border border-gray-200";
  }
}

function formatDateTime(value: string) {
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

function formatDuration(durationMs: number) {
  const ms = Number.isFinite(durationMs) ? Math.max(0, Math.trunc(durationMs)) : 0;
  if (ms < 1000) {
    return `${ms}ms`;
  }

  const seconds = ms / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds < 10 ? 1 : 0)}s`;
  }

  const minutes = Math.floor(seconds / 60);
  const remainSeconds = Math.round(seconds % 60);
  return `${minutes}m ${remainSeconds}s`;
}

function summarizeMessage(message: string) {
  const text = (message ?? "").trim();
  if (!text) {
    return "-";
  }
  if (text.length <= 26) {
    return text;
  }
  return `${text.slice(0, 26)}...`;
}

function formatMediaType(mediaType: string) {
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

function formatFieldList(fields: string[] | undefined) {
  if (!Array.isArray(fields) || fields.length === 0) {
    return "-";
  }
  return fields.join("、");
}

function formatFieldChanges(changes: Array<{ field: string; diff_type: string; before: string; after: string }> | undefined) {
  if (!Array.isArray(changes) || changes.length === 0) {
    return "-";
  }
  return changes
    .map((item) => `${item.field} [${item.diff_type || "remote"}]\n前: ${item.before || "-"}\n后: ${item.after || "-"}`)
    .join("\n\n");
}

function logsTotalPages() {
  return Math.max(1, Math.ceil(logsTotal.value / logsPageSize.value));
}

function detailTotalPages(total: number, pageSize: number) {
  const safeTotal = Math.max(0, Number(total) || 0);
  const safePageSize = normalizeNumber(Number(pageSize) || 10, 1, 100);
  return Math.max(1, Math.ceil(safeTotal / safePageSize));
}

function detailSyncedTotalPages() {
  return detailTotalPages(activeLogDetail.value?.synced ?? 0, detailSyncedPageSize.value);
}

function detailFailedTotalPages() {
  return detailTotalPages(activeLogDetail.value?.failed ?? 0, detailFailedPageSize.value);
}

async function loadAutoSyncLogs(page = logsPage.value) {
  logsLoading.value = true;
  logsError.value = "";

  try {
    const safePage = Math.max(1, Math.trunc(page));
    const resp = await getAutoSyncLogs({
      page: safePage,
      page_size: logsPageSize.value,
      status: logsStatus.value || undefined,
    });
    const data = resp.data;
    logsItems.value = Array.isArray(data.results) ? data.results : [];
    logsTotal.value = Math.max(0, Number(data.total) || 0);
    logsPage.value = normalizeNumber(Number(data.page), 1, logsTotalPages());
  } catch (err: any) {
    logsError.value = err.message ?? "读取执行日志失败";
  } finally {
    logsLoading.value = false;
  }
}

async function refreshLogs() {
  await loadAutoSyncLogs(logsPage.value);
}

function openClearLogsConfirm() {
  clearLogsConfirmVisible.value = true;
}

function closeClearLogsConfirm() {
  if (logsClearing.value) {
    return;
  }
  clearLogsConfirmVisible.value = false;
}

async function clearLogs() {
  logsClearing.value = true;
  logsError.value = "";
  logsMessage.value = "";

  try {
    const resp = await clearAutoSyncLogs();
    logsMessage.value = resp.data.message || "执行日志已清空";
    closeLogDetail();
    logsPage.value = 1;
    await loadAutoSyncLogs(1);
    clearLogsConfirmVisible.value = false;
  } catch (err: any) {
    logsError.value = err.message ?? "清空执行日志失败";
  } finally {
    logsClearing.value = false;
  }
}

async function applyLogStatusFilter() {
  logsPage.value = 1;
  await loadAutoSyncLogs(1);
}

async function goToLogsPage(page: number) {
  const target = normalizeNumber(page, 1, logsTotalPages());
  await loadAutoSyncLogs(target);
}

async function loadLogDetail(id: number, params: AdminAutoSyncLogDetailParams = {}, reset = false) {
  detailLoading.value = true;
  detailError.value = "";
  activeLogId.value = id;
  if (reset) {
    activeLogDetail.value = null;
  }

  try {
    const resp = await getAutoSyncLogDetail(id, {
      synced_page: params.synced_page ?? detailSyncedPage.value,
      synced_page_size: params.synced_page_size ?? detailSyncedPageSize.value,
      failed_page: params.failed_page ?? detailFailedPage.value,
      failed_page_size: params.failed_page_size ?? detailFailedPageSize.value,
    });
    const data = resp.data;
    activeLogDetail.value = data;
    detailSyncedPageSize.value = normalizeNumber(Number(data.synced_page_size) || detailSyncedPageSize.value, 1, 100);
    detailFailedPageSize.value = normalizeNumber(Number(data.failed_page_size) || detailFailedPageSize.value, 1, 100);
    detailSyncedPage.value = normalizeNumber(Number(data.synced_page) || 1, 1, detailTotalPages(data.synced, detailSyncedPageSize.value));
    detailFailedPage.value = normalizeNumber(Number(data.failed_page) || 1, 1, detailTotalPages(data.failed, detailFailedPageSize.value));
  } catch (err: any) {
    detailError.value = err.message ?? "读取日志明细失败";
  } finally {
    detailLoading.value = false;
  }
}

async function openLogDetail(item: AdminAutoSyncLogItem) {
  detailModalVisible.value = true;
  detailSyncedPage.value = 1;
  detailFailedPage.value = 1;
  await loadLogDetail(
    item.id,
    {
      synced_page: 1,
      synced_page_size: detailSyncedPageSize.value,
      failed_page: 1,
      failed_page_size: detailFailedPageSize.value,
    },
    true,
  );
}

async function goToDetailSyncedPage(page: number) {
  if (!activeLogId.value) {
    return;
  }
  const target = normalizeNumber(page, 1, detailSyncedTotalPages());
  await loadLogDetail(activeLogId.value, {
    synced_page: target,
    failed_page: detailFailedPage.value,
  });
}

async function goToDetailFailedPage(page: number) {
  if (!activeLogId.value) {
    return;
  }
  const target = normalizeNumber(page, 1, detailFailedTotalPages());
  await loadLogDetail(activeLogId.value, {
    synced_page: detailSyncedPage.value,
    failed_page: target,
  });
}

function closeLogDetail() {
  detailModalVisible.value = false;
  detailLoading.value = false;
  detailError.value = "";
  activeLogDetail.value = null;
  activeLogId.value = null;
  detailSyncedPage.value = 1;
  detailFailedPage.value = 1;
}

async function loadSettings() {
  loading.value = true;
  proxyError.value = "";
  proxyMessage.value = "";
  syncError.value = "";
  syncMessage.value = "";

  try {
    const [proxyResp, autoSyncResp] = await Promise.all([getProxySettings(), getAutoSyncSettings()]);
    const proxyData = proxyResp.data;
    proxyEnabled.value = !!proxyData.enabled;
    proxyURL.value = proxyData.proxy_url ?? "";
    proxyLocalWriteEnabled.value = proxyData.local_write_enabled !== false;

    const syncData = autoSyncResp.data;
    syncEnabled.value = !!syncData.enabled;
    syncCronExpr.value = (syncData.cron_expr ?? "").trim() || "*/30 * * * *";
    syncMode.value = syncData.mode === "overwrite_all" ? "overwrite_all" : "update_unmodified";
    syncBatchSize.value = normalizeNumber(Number(syncData.batch_size), 1, 500);
    syncStartDelaySecond.value = normalizeNumber(Number(syncData.start_delay_second), 0, 3600);
    syncRunning.value = !!syncData.running;
  } catch (err: any) {
    const text = err.message ?? "读取系统设置失败";
    proxyError.value = text;
    syncError.value = text;
  } finally {
    loading.value = false;
  }
}

async function saveProxySettings() {
  proxySaving.value = true;
  proxyError.value = "";
  proxyMessage.value = "";
  try {
    const nextProxyURL = proxyEnabled.value ? normalizeProxyURL(proxyURL.value) : "";
    const resp = await updateProxySettings({
      proxy_url: nextProxyURL,
      local_write_enabled: proxyLocalWriteEnabled.value,
    });
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    proxyEnabled.value = !!data.enabled;
    proxyLocalWriteEnabled.value = data.local_write_enabled !== false;
    proxyMessage.value = proxyEnabled.value ? "代理配置已保存" : "代理已关闭，当前为直连";
  } catch (err: any) {
    proxyError.value = err.message ?? "保存代理设置失败";
  } finally {
    proxySaving.value = false;
  }
}

async function saveAutoSyncSettings() {
  syncSaving.value = true;
  syncError.value = "";
  syncMessage.value = "";
  try {
    const payload = {
      enabled: syncEnabled.value,
      cron_expr: syncCronExpr.value.trim(),
      mode: syncMode.value,
      batch_size: normalizeNumber(syncBatchSize.value, 1, 500),
      start_delay_second: normalizeNumber(syncStartDelaySecond.value, 0, 3600),
    };
    const resp = await updateAutoSyncSettings(payload);
    const data = resp.data;
    syncEnabled.value = !!data.enabled;
    syncCronExpr.value = (data.cron_expr ?? "").trim() || "*/30 * * * *";
    syncMode.value = data.mode === "overwrite_all" ? "overwrite_all" : "update_unmodified";
    syncBatchSize.value = normalizeNumber(Number(data.batch_size), 1, 500);
    syncStartDelaySecond.value = normalizeNumber(Number(data.start_delay_second), 0, 3600);
    syncRunning.value = !!data.running;
    syncMessage.value = syncEnabled.value ? "自动同步配置已保存并生效" : "自动同步已关闭";
  } catch (err: any) {
    syncError.value = err.message ?? "保存自动同步设置失败";
  } finally {
    syncSaving.value = false;
  }
}

async function triggerAutoSyncNow() {
  syncTriggering.value = true;
  syncError.value = "";
  syncMessage.value = "";

  try {
    const resp = await runAutoSyncNow();
    const data = resp.data;
    syncRunning.value = !!data.running;
    syncMessage.value = data.message || "已触发一次立即同步任务";
    await loadAutoSyncLogs(1);
  } catch (err: any) {
    syncError.value = err.message ?? "触发立即同步失败";
  } finally {
    syncTriggering.value = false;
  }
}

async function reloadAll() {
  await Promise.all([loadSettings(), loadAutoSyncLogs(logsPage.value)]);
}

onMounted(reloadAll);
</script>

<template>
  <section class="grid gap-4">
    <section class="settings-toolbar card">
      <div class="min-w-0">
        <p class="section-label">系统设置</p>
        <h2 class="library-toolbar-title">运行配置</h2>
        <p class="mt-1 text-sm text-black/55">统一管理 TMDB 网络代理、库内定时同步任务和执行日志。</p>
      </div>

      <div class="library-toolbar-actions">
        <span class="badge">{{ taskRunStatusText }}</span>
        <button
          class="btn-soft disabled:opacity-60"
          :disabled="settingsBusy"
          @click="reloadAll"
        >
          {{ loading || logsLoading ? "读取中..." : "重新读取" }}
        </button>
      </div>
    </section>

    <section class="settings-summary-grid">
      <article class="settings-summary-card">
        <span class="settings-summary-label">代理访问</span>
        <strong>{{ proxyStatusText }}</strong>
        <p>{{ proxyEnabled ? proxyURL || "已启用，等待代理地址" : "后端直连 TMDB" }} · {{ proxyLocalWriteStatusText }}</p>
      </article>
      <article class="settings-summary-card">
        <span class="settings-summary-label">自动同步</span>
        <strong>{{ syncStatusText }}</strong>
        <p>{{ syncEnabled ? `${syncCronExpr} · ${formatMode(syncMode)}` : "不会自动调度同步任务" }}</p>
      </article>
      <article class="settings-summary-card">
        <span class="settings-summary-label">任务状态</span>
        <strong>{{ taskRunStatusText }}</strong>
        <p>批大小 {{ syncBatchSize }} · 启动延迟 {{ syncStartDelaySecond }} 秒</p>
      </article>
      <article class="settings-summary-card">
        <span class="settings-summary-label">最近执行</span>
        <strong>{{ latestLogStatusText }}</strong>
        <p>{{ latestLogTimeText }}</p>
      </article>
    </section>

    <section class="settings-form-grid">
      <div class="card settings-card">
        <div class="settings-panel-header">
          <div>
            <p class="section-label">Network</p>
            <h3 class="settings-section-title">代理设置</h3>
            <p class="settings-note">配置后端访问 TMDB 时使用的网络代理。</p>
          </div>
          <span class="badge">{{ proxyStatusText }}</span>
        </div>

        <label class="settings-toggle-row">
          <input v-model="proxyEnabled" type="checkbox" class="check-control" />
          <span>
            <strong>启用代理访问 TMDB</strong>
            <small>关闭后恢复为直连，保存后即时生效。</small>
          </span>
        </label>

        <label class="settings-toggle-row">
          <input v-model="proxyLocalWriteEnabled" type="checkbox" class="check-control" :disabled="proxySaving" />
          <span>
            <strong>允许代理自动写入本地库</strong>
            <small>关闭后仍优先读取已有本地数据，回源 TMDB 成功后不再新增或更新本地库。</small>
          </span>
        </label>

        <label class="settings-field-label">
          代理地址
          <input
            v-model="proxyURL"
            type="text"
            class="field-control mt-1 w-full text-sm"
            :disabled="!proxyEnabled || proxySaving"
            placeholder="http://127.0.0.1:7890"
          />
        </label>
        <p class="settings-help-text">支持格式示例：http://127.0.0.1:7890、socks5://127.0.0.1:1080</p>

        <div class="settings-card-actions">
          <button
            class="btn-primary disabled:opacity-60"
            :disabled="proxySaving"
            @click="saveProxySettings"
          >
            {{ proxySaving ? "保存中..." : "保存代理设置" }}
          </button>
        </div>
        <p v-if="proxyMessage" class="settings-feedback settings-feedback-success">{{ proxyMessage }}</p>
        <p v-if="proxyError" class="settings-feedback settings-feedback-error">{{ proxyError }}</p>
      </div>

      <div class="card settings-card">
        <div class="settings-panel-header">
          <div>
            <p class="section-label">Schedule</p>
            <h3 class="settings-section-title">定时同步设置</h3>
            <p class="settings-note">仅支持 cron 表达式调度，保存后即时生效。</p>
          </div>
          <span class="badge">{{ taskRunStatusText }}</span>
        </div>

        <label class="settings-toggle-row">
          <input v-model="syncEnabled" type="checkbox" class="check-control" />
          <span>
            <strong>启用自动同步任务</strong>
            <small>任务会按 Cron 周期检查远端字段变更。</small>
          </span>
        </label>

        <div class="grid gap-3 md:grid-cols-2">
          <label class="settings-field-label md:col-span-2">
            Cron 表达式
            <input
              v-model="syncCronExpr"
              type="text"
              class="field-control mt-1 w-full text-sm"
              :disabled="syncSaving"
              placeholder="*/30 * * * *"
            />
            <span>5 段格式：分 时 日 月 周，例如 0 3 * * *（每天 03:00）。</span>
          </label>

          <label class="settings-field-label md:col-span-2">
            同步策略
            <GlassSelect v-model="syncMode" :options="modeOptions" :disabled="syncSaving" class="mt-1 w-full" />
            <span>{{ modeOptions.find((item) => item.value === syncMode)?.hint }}</span>
          </label>

          <label class="settings-field-label">
            每轮批大小（条）
            <input
              v-model.number="syncBatchSize"
              type="number"
              min="1"
              max="500"
              class="field-control mt-1 w-full text-sm"
              :disabled="syncSaving"
            />
          </label>

          <label class="settings-field-label">
            启动延迟（秒）
            <input
              v-model.number="syncStartDelaySecond"
              type="number"
              min="0"
              max="3600"
              class="field-control mt-1 w-full text-sm"
              :disabled="syncSaving"
            />
          </label>
        </div>

        <div class="settings-card-actions">
          <button
            class="btn-primary disabled:opacity-60"
            :disabled="syncSaving || syncTriggering"
            @click="saveAutoSyncSettings"
          >
            {{ syncSaving ? "保存中..." : "保存定时同步设置" }}
          </button>
          <button
            class="btn-soft disabled:opacity-60"
            :disabled="syncSaving || syncTriggering"
            @click="triggerAutoSyncNow"
          >
            {{ syncTriggering ? "触发中..." : "立即执行一次" }}
          </button>
        </div>
        <p v-if="syncMessage" class="settings-feedback settings-feedback-success">{{ syncMessage }}</p>
        <p v-if="syncError" class="settings-feedback settings-feedback-error">{{ syncError }}</p>
      </div>
    </section>

    <div class="card settings-card-wide settings-log-card">
      <div class="settings-log-header">
        <div>
          <p class="section-label">Logs</p>
          <h3 class="settings-section-title">定时任务执行日志</h3>
          <p class="settings-note">最近执行记录会持久化到数据库，可按状态筛选查看。</p>
        </div>

        <div class="settings-log-actions">
          <label class="settings-log-filter">
            状态
            <GlassSelect
              v-model="logsStatus"
              :options="logStatusOptions"
              :disabled="logsLoading || logsClearing"
              class="min-w-[136px]"
              @change="applyLogStatusFilter"
            />
          </label>

          <button
            class="btn-soft disabled:opacity-60"
            :disabled="logsLoading || logsClearing"
            @click="refreshLogs"
          >
            {{ logsLoading ? "刷新中..." : "刷新日志" }}
          </button>
          <button
            class="btn-danger-soft disabled:opacity-60"
            :disabled="logsLoading || logsClearing"
            @click="openClearLogsConfirm"
          >
            {{ logsClearing ? "清空中..." : "清空日志" }}
          </button>
        </div>
      </div>

      <p v-if="logsMessage" class="settings-feedback settings-feedback-success">{{ logsMessage }}</p>
      <p v-if="logsError" class="settings-feedback settings-feedback-error">{{ logsError }}</p>

      <div class="table-shell settings-table-shell">
        <table class="min-w-full text-sm settings-log-table">
          <thead class="table-head text-left text-black/70">
            <tr>
              <th class="px-3 py-2 font-medium">触发时间</th>
              <th class="px-3 py-2 font-medium">策略</th>
              <th class="px-3 py-2 font-medium">状态</th>
              <th class="px-3 py-2 font-medium">检查/同步/失败</th>
              <th class="px-3 py-2 font-medium">耗时</th>
              <th class="px-3 py-2 font-medium">摘要</th>
              <th class="px-3 py-2 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in logsItems"
              :key="item.id"
              class="table-row-hover"
            >
              <td class="px-3 py-2">
                <p class="settings-table-primary">{{ formatDateTime(item.triggered_at) }}</p>
                <p class="mt-1 text-xs text-black/45">{{ item.cron_expr || "-" }}</p>
              </td>
              <td class="px-3 py-2">
                <p class="settings-table-primary">{{ formatMode(item.mode) }}</p>
                <p class="mt-1 text-xs text-black/45">批大小 {{ item.batch_size }}</p>
              </td>
              <td class="px-3 py-2">
                <span class="settings-status-pill" :class="statusClass(item.status)">
                  {{ formatStatus(item.status) }}
                </span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span class="settings-count-pill">{{ item.checked }}</span>
                <span class="settings-count-pill settings-count-success">{{ item.synced }}</span>
                <span class="settings-count-pill settings-count-danger">{{ item.failed }}</span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">{{ formatDuration(item.duration_ms) }}</td>
              <td class="px-3 py-2 text-black/70">
                <span class="settings-log-summary">{{ summarizeMessage(item.message) }}</span>
              </td>
              <td class="px-3 py-2">
                <button
                  class="btn-soft-xs px-2.5 py-1"
                  @click="openLogDetail(item)"
                >
                  详情
                </button>
              </td>
            </tr>
            <tr v-if="!logsLoading && logsItems.length === 0">
              <td colspan="7" class="px-3 py-6 text-center text-black/55">暂无执行日志</td>
            </tr>
            <tr v-if="logsLoading">
              <td colspan="7" class="px-3 py-6 text-center text-black/55">日志加载中...</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="settings-pagination-row">
        <p>共 {{ logsTotal }} 条，当前第 {{ logsPage }} / {{ logsTotalPages() }} 页</p>
        <div class="flex items-center gap-2">
          <button
            class="btn-soft px-3 py-1.5 disabled:opacity-60"
            :disabled="logsLoading || logsPage <= 1"
            @click="goToLogsPage(logsPage - 1)"
          >
            上一页
          </button>
          <button
            class="btn-soft px-3 py-1.5 disabled:opacity-60"
            :disabled="logsLoading || logsPage >= logsTotalPages()"
            @click="goToLogsPage(logsPage + 1)"
          >
            下一页
          </button>
        </div>
      </div>

    </div>

    <div v-if="detailModalVisible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/55 p-3 sm:p-4" @click.self="closeLogDetail">
      <div class="panel-glass settings-detail-modal max-h-[92vh] w-full max-w-6xl overflow-hidden rounded-2xl">
        <div class="modal-header-dark">
          <div>
            <p class="section-label">Run Detail</p>
            <h4 class="text-base font-semibold">
              执行日志明细
              <span v-if="activeLogDetail" class="text-sm text-black/55">#{{ activeLogDetail.id }}</span>
            </h4>
          </div>
          <button
            class="btn-soft px-3 py-1.5"
            @click="closeLogDetail"
          >
            关闭
          </button>
        </div>

        <div class="max-h-[calc(92vh-72px)] overflow-y-auto px-4 py-4 sm:px-5">
          <p v-if="detailLoading && !activeLogDetail" class="text-sm text-black/60">明细加载中...</p>
          <p v-if="detailError" class="settings-feedback settings-feedback-error mb-3">{{ detailError }}</p>

          <template v-if="activeLogDetail">
            <p v-if="detailLoading" class="mb-3 text-xs text-black/50">分页加载中...</p>
            <div class="settings-detail-summary-grid">
              <article class="settings-detail-summary-item">
                <span>触发时间</span>
                <strong>{{ formatDateTime(activeLogDetail.triggered_at) }}</strong>
                <small>{{ activeLogDetail.cron_expr || "-" }}</small>
              </article>
              <article class="settings-detail-summary-item">
                <span>同步策略</span>
                <strong>{{ formatMode(activeLogDetail.mode) }}</strong>
                <small>批大小 {{ activeLogDetail.batch_size }}</small>
              </article>
              <article class="settings-detail-summary-item">
                <span>状态</span>
                <strong>{{ formatStatus(activeLogDetail.status) }}</strong>
                <small>耗时 {{ formatDuration(activeLogDetail.duration_ms) }}</small>
              </article>
              <article class="settings-detail-summary-item">
                <span>检查 / 同步 / 失败</span>
                <strong>{{ activeLogDetail.checked }} / {{ activeLogDetail.synced }} / {{ activeLogDetail.failed }}</strong>
                <small>{{ activeLogDetail.message || "-" }}</small>
              </article>
            </div>

            <div class="settings-detail-section">
              <div class="settings-detail-section-header">
                <div>
                  <h5 class="text-sm font-semibold text-green-700">同步成功项</h5>
                  <p class="settings-note">展示成功同步条目、远端差异字段和本地字段处理结果。</p>
                </div>
                <span class="badge">{{ activeLogDetail.synced }} 条</span>
              </div>
              <div class="table-shell settings-table-shell">
                <table class="min-w-full text-sm settings-detail-table">
                  <thead class="table-head text-left text-black/70">
                    <tr>
                      <th class="px-3 py-2 font-medium">类型</th>
                      <th class="px-3 py-2 font-medium">名称</th>
                      <th class="px-3 py-2 font-medium">TMDB ID</th>
                      <th class="px-3 py-2 font-medium">远端差异</th>
                      <th class="px-3 py-2 font-medium">本地处理</th>
                      <th class="px-3 py-2 font-medium">字段前后</th>
                      <th class="px-3 py-2 font-medium">信息</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(entry, idx) in activeLogDetail.synced_list"
                      :key="`synced-${idx}-${entry.media_type}-${entry.tmdb_id}`"
                      class="table-row-hover"
                    >
                      <td class="px-3 py-2 whitespace-nowrap">{{ formatMediaType(entry.media_type) }}</td>
                      <td class="px-3 py-2 min-w-40">
                        <p class="settings-table-primary line-clamp-2">{{ entry.name || "-" }}</p>
                      </td>
                      <td class="px-3 py-2 whitespace-nowrap">{{ entry.tmdb_id || "-" }}</td>
                      <td class="px-3 py-2 min-w-44">
                        <p class="settings-chip-list">{{ formatFieldList(entry.remote_diff_fields) }}</p>
                      </td>
                      <td class="px-3 py-2 min-w-56">
                        <div class="settings-field-stack">
                          <p><span>变更</span>{{ formatFieldList(entry.changed_fields) }}</p>
                          <p><span>覆盖</span>{{ formatFieldList(entry.overwritten_fields) }}</p>
                          <p><span>保留</span>{{ formatFieldList(entry.kept_local_fields) }}</p>
                        </div>
                      </td>
                      <td class="px-3 py-2 min-w-72">
                        <pre class="settings-diff-pre">{{ formatFieldChanges(entry.field_changes) }}</pre>
                      </td>
                      <td class="px-3 py-2 min-w-48 text-black/70">{{ entry.message || "-" }}</td>
                    </tr>
                    <tr v-if="activeLogDetail.synced_list.length === 0">
                      <td colspan="7" class="px-3 py-4 text-center text-black/55">无成功同步明细</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="settings-pagination-row settings-pagination-row-sm">
                <p>共 {{ activeLogDetail.synced }} 条，当前第 {{ detailSyncedPage }} / {{ detailSyncedTotalPages() }} 页</p>
                <div class="flex items-center gap-2">
                  <button
                    class="btn-soft px-3 py-1.5 disabled:opacity-60"
                    :disabled="detailLoading || detailSyncedPage <= 1"
                    @click="goToDetailSyncedPage(detailSyncedPage - 1)"
                  >
                    上一页
                  </button>
                  <button
                    class="btn-soft px-3 py-1.5 disabled:opacity-60"
                    :disabled="detailLoading || detailSyncedPage >= detailSyncedTotalPages()"
                    @click="goToDetailSyncedPage(detailSyncedPage + 1)"
                  >
                    下一页
                  </button>
                </div>
              </div>
            </div>

            <div class="settings-detail-section">
              <div class="settings-detail-section-header">
                <div>
                  <h5 class="text-sm font-semibold text-red-700">同步失败项</h5>
                  <p class="settings-note">失败条目会保留原因，便于定位网络、数据或接口异常。</p>
                </div>
                <span class="badge">{{ activeLogDetail.failed }} 条</span>
              </div>
              <div class="table-shell settings-table-shell">
                <table class="min-w-full text-sm settings-detail-table">
                  <thead class="table-head text-left text-black/70">
                    <tr>
                      <th class="px-3 py-2 font-medium">类型</th>
                      <th class="px-3 py-2 font-medium">名称</th>
                      <th class="px-3 py-2 font-medium">TMDB ID</th>
                      <th class="px-3 py-2 font-medium">失败原因</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(entry, idx) in activeLogDetail.failed_list"
                      :key="`failed-${idx}-${entry.media_type}-${entry.tmdb_id}`"
                      class="table-row-hover"
                    >
                      <td class="px-3 py-2 whitespace-nowrap">{{ formatMediaType(entry.media_type) }}</td>
                      <td class="px-3 py-2 min-w-44">
                        <p class="settings-table-primary line-clamp-2">{{ entry.name || "-" }}</p>
                      </td>
                      <td class="px-3 py-2 whitespace-nowrap">{{ entry.tmdb_id || "-" }}</td>
                      <td class="px-3 py-2 min-w-96 text-black/70">{{ entry.message || "-" }}</td>
                    </tr>
                    <tr v-if="activeLogDetail.failed_list.length === 0">
                      <td colspan="4" class="px-3 py-4 text-center text-black/55">无失败明细</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="settings-pagination-row settings-pagination-row-sm">
                <p>共 {{ activeLogDetail.failed }} 条，当前第 {{ detailFailedPage }} / {{ detailFailedTotalPages() }} 页</p>
                <div class="flex items-center gap-2">
                  <button
                    class="btn-soft px-3 py-1.5 disabled:opacity-60"
                    :disabled="detailLoading || detailFailedPage <= 1"
                    @click="goToDetailFailedPage(detailFailedPage - 1)"
                  >
                    上一页
                  </button>
                  <button
                    class="btn-soft px-3 py-1.5 disabled:opacity-60"
                    :disabled="detailLoading || detailFailedPage >= detailFailedTotalPages()"
                    @click="goToDetailFailedPage(detailFailedPage + 1)"
                  >
                    下一页
                  </button>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <div v-if="clearLogsConfirmVisible" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/45 p-4" @click.self="closeClearLogsConfirm">
      <div class="panel-glass w-full max-w-md rounded-2xl p-5">
        <h4 class="text-base font-semibold text-red-700">确认清空执行日志</h4>
        <p class="mt-2 text-sm text-black/70">
          确认要清空所有执行日志吗？清空后无法恢复。
        </p>

        <div class="mt-5 flex items-center justify-end gap-2">
          <button
            class="btn-soft disabled:opacity-60"
            :disabled="logsClearing"
            @click="closeClearLogsConfirm"
          >
            取消
          </button>
          <button
            class="btn-danger-soft disabled:opacity-60"
            :disabled="logsClearing"
            @click="clearLogs"
          >
            {{ logsClearing ? "清空中..." : "确认清空" }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
