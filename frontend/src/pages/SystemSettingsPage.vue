<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  getAutoSyncLogs,
  getAutoSyncSettings,
  getProxySettings,
  updateAutoSyncSettings,
  updateProxySettings,
  type AdminAutoSyncLogItem,
  type AdminAutoSyncMode,
} from "@/api/admin";

const loading = ref(false);

const proxySaving = ref(false);
const proxyError = ref("");
const proxyMessage = ref("");
const proxyEnabled = ref(false);
const proxyURL = ref("");

const syncSaving = ref(false);
const syncError = ref("");
const syncMessage = ref("");
const syncEnabled = ref(true);
const syncCronExpr = ref("*/30 * * * *");
const syncMode = ref<AdminAutoSyncMode>("update_unmodified");
const syncBatchSize = ref(50);
const syncStartDelaySecond = ref(15);
const syncRunning = ref(false);

const logsLoading = ref(false);
const logsError = ref("");
const logsStatus = ref("");
const logsPage = ref(1);
const logsPageSize = ref(10);
const logsTotal = ref(0);
const logsItems = ref<AdminAutoSyncLogItem[]>([]);

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

function logsTotalPages() {
  return Math.max(1, Math.ceil(logsTotal.value / logsPageSize.value));
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

async function applyLogStatusFilter() {
  logsPage.value = 1;
  await loadAutoSyncLogs(1);
}

async function goToLogsPage(page: number) {
  const target = normalizeNumber(page, 1, logsTotalPages());
  await loadAutoSyncLogs(target);
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
    const resp = await updateProxySettings({ proxy_url: nextProxyURL });
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    proxyEnabled.value = !!data.enabled;
    proxyMessage.value = proxyEnabled.value ? "代理已启用" : "代理已关闭，当前为直连";
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

async function reloadAll() {
  await Promise.all([loadSettings(), loadAutoSyncLogs(logsPage.value)]);
}

onMounted(reloadAll);
</script>

<template>
  <section class="grid gap-4">
    <div class="card max-w-2xl">
      <h2 class="text-lg font-semibold">系统设置</h2>
      <p class="mt-1 text-sm text-black/60">统一管理代理访问和库内定时同步任务。</p>
      <p v-if="loading" class="mt-4 text-sm text-black/55">加载中...</p>
      <div v-else class="mt-3">
        <button
          class="rounded-xl border border-black/10 bg-white px-4 py-2 text-sm hover:bg-sand/50 disabled:opacity-60"
          :disabled="proxySaving || syncSaving || logsLoading"
          @click="reloadAll"
        >
          重新读取全部设置
        </button>
      </div>
    </div>

    <div class="card max-w-2xl">
      <h3 class="text-base font-semibold">代理设置</h3>
      <p class="mt-1 text-sm text-black/60">配置后端访问 TMDB 时使用的网络代理。</p>

      <label class="mt-4 inline-flex items-center gap-2 text-sm">
        <input v-model="proxyEnabled" type="checkbox" />
        <span>启用代理访问 TMDB</span>
      </label>

      <label class="mt-3 block text-xs text-black/60">
        代理地址
        <input
          v-model="proxyURL"
          type="text"
          class="mt-1 w-full rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
          :disabled="!proxyEnabled || proxySaving"
          placeholder="http://127.0.0.1:7890"
        />
      </label>
      <p class="mt-2 text-xs text-black/50">支持格式示例：`http://127.0.0.1:7890`、`socks5://127.0.0.1:1080`</p>

      <div class="mt-4 flex items-center gap-3">
        <button
          class="rounded-xl bg-pine px-4 py-2 text-sm font-medium text-white hover:bg-pine/90 disabled:opacity-60"
          :disabled="proxySaving"
          @click="saveProxySettings"
        >
          {{ proxySaving ? "保存中..." : "保存代理设置" }}
        </button>
      </div>
      <p v-if="proxyMessage" class="mt-3 text-sm text-green-700">{{ proxyMessage }}</p>
      <p v-if="proxyError" class="mt-3 text-sm text-red-600">{{ proxyError }}</p>
    </div>

    <div class="card max-w-2xl">
      <h3 class="text-base font-semibold">定时同步设置</h3>
      <p class="mt-1 text-sm text-black/60">仅支持 cron 表达式调度，保存后即时生效。</p>

      <label class="mt-4 inline-flex items-center gap-2 text-sm">
        <input v-model="syncEnabled" type="checkbox" />
        <span>启用自动同步任务</span>
      </label>
      <p class="mt-2 text-xs text-black/50">当前运行状态：{{ syncRunning ? "执行中" : "空闲" }}</p>

      <label class="mt-3 block text-xs text-black/60">
        Cron 表达式
        <input
          v-model="syncCronExpr"
          type="text"
          class="mt-1 w-full rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
          :disabled="syncSaving"
          placeholder="*/30 * * * *"
        />
      </label>
      <p class="mt-1 text-xs text-black/50">5 段格式：分 时 日 月 周，例如 `0 3 * * *`（每天 03:00）。</p>

      <label class="mt-3 block text-xs text-black/60">
        同步策略
        <select
          v-model="syncMode"
          class="mt-1 w-full rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
          :disabled="syncSaving"
        >
          <option
            v-for="item in modeOptions"
            :key="item.value"
            :value="item.value"
          >
            {{ item.label }}
          </option>
        </select>
      </label>
      <p class="mt-1 text-xs text-black/50">{{ modeOptions.find((item) => item.value === syncMode)?.hint }}</p>

      <label class="mt-3 block text-xs text-black/60">
        每轮批大小（条）
        <input
          v-model.number="syncBatchSize"
          type="number"
          min="1"
          max="500"
          class="mt-1 w-full rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
          :disabled="syncSaving"
        />
      </label>

      <label class="mt-3 block text-xs text-black/60">
        启动延迟（秒）
        <input
          v-model.number="syncStartDelaySecond"
          type="number"
          min="0"
          max="3600"
          class="mt-1 w-full rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
          :disabled="syncSaving"
        />
      </label>

      <div class="mt-4 flex items-center gap-3">
        <button
          class="rounded-xl bg-pine px-4 py-2 text-sm font-medium text-white hover:bg-pine/90 disabled:opacity-60"
          :disabled="syncSaving"
          @click="saveAutoSyncSettings"
        >
          {{ syncSaving ? "保存中..." : "保存定时同步设置" }}
        </button>
      </div>
      <p v-if="syncMessage" class="mt-3 text-sm text-green-700">{{ syncMessage }}</p>
      <p v-if="syncError" class="mt-3 text-sm text-red-600">{{ syncError }}</p>
    </div>

    <div class="card">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h3 class="text-base font-semibold">定时任务执行日志</h3>
          <p class="mt-1 text-sm text-black/60">最近执行记录会持久化到数据库，可按状态筛选查看。</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="text-xs text-black/60">
            状态
            <select
              v-model="logsStatus"
              class="ml-2 rounded-xl border border-black/10 bg-white px-3 py-2 text-sm"
              :disabled="logsLoading"
              @change="applyLogStatusFilter"
            >
              <option
                v-for="item in logStatusOptions"
                :key="item.value"
                :value="item.value"
              >
                {{ item.label }}
              </option>
            </select>
          </label>

          <button
            class="rounded-xl border border-black/10 bg-white px-4 py-2 text-sm hover:bg-sand/50 disabled:opacity-60"
            :disabled="logsLoading"
            @click="refreshLogs"
          >
            {{ logsLoading ? "刷新中..." : "刷新日志" }}
          </button>
        </div>
      </div>

      <p v-if="logsError" class="mt-3 text-sm text-red-600">{{ logsError }}</p>

      <div class="mt-4 overflow-x-auto rounded-xl border border-black/10 bg-white">
        <table class="min-w-full text-sm">
          <thead class="bg-black/5 text-left text-black/70">
            <tr>
              <th class="px-3 py-2 font-medium">触发时间</th>
              <th class="px-3 py-2 font-medium">策略</th>
              <th class="px-3 py-2 font-medium">状态</th>
              <th class="px-3 py-2 font-medium">检查/同步/失败</th>
              <th class="px-3 py-2 font-medium">耗时</th>
              <th class="px-3 py-2 font-medium">信息</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in logsItems"
              :key="item.id"
              class="border-t border-black/5"
            >
              <td class="px-3 py-2">
                <p>{{ formatDateTime(item.triggered_at) }}</p>
                <p class="mt-1 text-xs text-black/45">{{ item.cron_expr || "-" }}</p>
              </td>
              <td class="px-3 py-2">
                <p>{{ formatMode(item.mode) }}</p>
                <p class="mt-1 text-xs text-black/45">批大小 {{ item.batch_size }}</p>
              </td>
              <td class="px-3 py-2">
                <span class="inline-flex rounded-full px-2 py-1 text-xs" :class="statusClass(item.status)">
                  {{ formatStatus(item.status) }}
                </span>
              </td>
              <td class="px-3 py-2">
                {{ item.checked }} / {{ item.synced }} / {{ item.failed }}
              </td>
              <td class="px-3 py-2">{{ formatDuration(item.duration_ms) }}</td>
              <td class="px-3 py-2 text-black/70">{{ item.message || "-" }}</td>
            </tr>
            <tr v-if="!logsLoading && logsItems.length === 0">
              <td colspan="6" class="px-3 py-6 text-center text-black/55">暂无执行日志</td>
            </tr>
            <tr v-if="logsLoading">
              <td colspan="6" class="px-3 py-6 text-center text-black/55">日志加载中...</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-3 flex items-center justify-between text-sm text-black/65">
        <p>共 {{ logsTotal }} 条，当前第 {{ logsPage }} / {{ logsTotalPages() }} 页</p>
        <div class="flex items-center gap-2">
          <button
            class="rounded-xl border border-black/10 bg-white px-3 py-1.5 hover:bg-sand/50 disabled:opacity-60"
            :disabled="logsLoading || logsPage <= 1"
            @click="goToLogsPage(logsPage - 1)"
          >
            上一页
          </button>
          <button
            class="rounded-xl border border-black/10 bg-white px-3 py-1.5 hover:bg-sand/50 disabled:opacity-60"
            :disabled="logsLoading || logsPage >= logsTotalPages()"
            @click="goToLogsPage(logsPage + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
