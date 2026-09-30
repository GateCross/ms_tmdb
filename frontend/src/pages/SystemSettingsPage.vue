<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import LoadState from "@/components/common/LoadState.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import {
  getAutoSyncLogs,
  getAutoSyncSettings,
  getLogSettings,
  getProxySettings,
  runAutoSyncNow,
  updateAutoSyncSettings,
  updateLogSettings,
  updateProxySettings,
  type AdminAutoSyncLogItem,
  type AdminAutoSyncMode,
} from "@/api/admin";
import { useToastNotice } from "@/composables/useToastNotice";
const loading = ref(false);
const settingsLoaded = ref(false);
const appVersion = __APP_VERSION__;
const initialLoading = computed(() => loading.value && !settingsLoaded.value);

const proxySaving = ref(false);
const proxyEnabled = ref(false);
const proxyURL = ref("");
const proxyLocalWriteEnabled = ref(true);
const proxyTimeout = ref(30000);
const proxyTimeoutRestartRequired = ref(false);
const proxyPublicBaseURL = ref("");

const logSaving = ref(false);
const logRetentionDays = ref(7);

const syncSaving = ref(false);
const syncEnabled = ref(true);
const syncCronExpr = ref("*/30 * * * *");
const syncMode = ref<AdminAutoSyncMode>("update_unmodified");
const syncBatchSize = ref(50);
const syncStartDelaySecond = ref(15);
const syncRunning = ref(false);
const syncTriggering = ref(false);

const logsLoading = ref(false);
const logsItems = ref<AdminAutoSyncLogItem[]>([]);
const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

const modeOptions: Array<{ label: string; value: AdminAutoSyncMode; hint: string }> = [
  { label: "仅更新未在本地修改的字段", value: "update_unmodified", hint: "保留本地改动，只更新 TMDB 远端变化字段" },
  { label: "全量覆盖", value: "overwrite_all", hint: "使用 TMDB 最新数据覆盖本地字段" },
];

const settingsBusy = computed(
  () => loading.value || proxySaving.value || logSaving.value || syncSaving.value || syncTriggering.value || logsLoading.value,
);
const networkSaving = computed(() => proxySaving.value || logSaving.value);
const proxyStatusText = computed(() => (proxyEnabled.value ? "已启用" : "直连"));
const proxyLocalWriteStatusText = computed(() => (proxyLocalWriteEnabled.value ? "可写本地" : "只读本地"));
const logRetentionStatusText = computed(() => `保留 ${logRetentionDays.value} 天`);
const syncStatusText = computed(() => (syncEnabled.value ? "已启用" : "已关闭"));
const taskRunStatusText = computed(() => (syncRunning.value ? "执行中" : "空闲"));
const latestLog = computed(() => logsItems.value[0] ?? null);
const latestLogStatusText = computed(() => (latestLog.value ? formatStatus(latestLog.value.status) : "暂无记录"));
const latestLogTimeText = computed(() =>
  latestLog.value ? formatDateTime(latestLog.value.triggered_at) : "等待首次执行",
);

function normalizeProxyURL(raw: string) {
  return raw.trim();
}

function normalizeNumber(value: number, min: number, max: number) {
  const next = Number.isFinite(value) ? Math.trunc(value) : min;
  if (next < min) return min;
  if (next > max) return max;
  return next;
}

function normalizeTimeout(value: number) {
  return normalizeNumber(Number(value), 1000, 300000);
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
    case "canceled":
      return "已取消";
    default:
      return status || "-";
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

async function loadAutoSyncLogs() {
  logsLoading.value = true;

  try {
    const resp = await getAutoSyncLogs({ page: 1, page_size: 1 });
    const data = resp.data;
    logsItems.value = Array.isArray(data.results) ? data.results : [];
  } catch {
    // 错误已由全局请求拦截器提示。
  } finally {
    logsLoading.value = false;
  }
}

async function loadSettings() {
  loading.value = true;

  try {
    const [proxyResp, autoSyncResp, logResp] = await Promise.all([
      getProxySettings(),
      getAutoSyncSettings(),
      getLogSettings(),
    ]);
    const proxyData = proxyResp.data;
    proxyEnabled.value = !!proxyData.enabled;
    proxyURL.value = proxyData.proxy_url ?? "";
    proxyLocalWriteEnabled.value = proxyData.local_write_enabled !== false;
    proxyTimeout.value = normalizeTimeout(Number(proxyData.timeout) || 30000);
    proxyPublicBaseURL.value = proxyData.public_base_url ?? "";
    proxyTimeoutRestartRequired.value = !!proxyData.timeout_restart_required;

    const syncData = autoSyncResp.data;
    syncEnabled.value = !!syncData.enabled;
    syncCronExpr.value = (syncData.cron_expr ?? "").trim() || "*/30 * * * *";
    syncMode.value = syncData.mode === "overwrite_all" ? "overwrite_all" : "update_unmodified";
    syncBatchSize.value = normalizeNumber(Number(syncData.batch_size), 1, 500);
    syncStartDelaySecond.value = normalizeNumber(Number(syncData.start_delay_second), 0, 3600);
    syncRunning.value = !!syncData.running;

    const logData = logResp.data;
    logRetentionDays.value = normalizeNumber(Number(logData.retention_days), 1, 365) || 7;
    settingsLoaded.value = true;
  } catch {
    // 错误已由全局请求拦截器提示。
  } finally {
    loading.value = false;
  }
}

async function saveNetworkSettings() {
  proxySaving.value = true;
  logSaving.value = true;
  try {
    const nextProxyURL = proxyEnabled.value ? normalizeProxyURL(proxyURL.value) : "";
    const nextRetentionDays = normalizeNumber(logRetentionDays.value, 1, 365);
    const [proxyResp, logResp] = await Promise.all([
      updateProxySettings({
        proxy_url: nextProxyURL,
        local_write_enabled: proxyLocalWriteEnabled.value,
        timeout: normalizeTimeout(proxyTimeout.value),
        public_base_url: normalizeProxyURL(proxyPublicBaseURL.value),
      }),
      updateLogSettings({
        retention_days: nextRetentionDays,
      }),
    ]);

    const proxyData = proxyResp.data;
    proxyURL.value = proxyData.proxy_url ?? "";
    proxyEnabled.value = !!proxyData.enabled;
    proxyLocalWriteEnabled.value = proxyData.local_write_enabled !== false;
    proxyTimeout.value = normalizeTimeout(Number(proxyData.timeout) || proxyTimeout.value);
    proxyPublicBaseURL.value = proxyData.public_base_url ?? "";
    proxyTimeoutRestartRequired.value = !!proxyData.timeout_restart_required;

    const logData = logResp.data;
    logRetentionDays.value = normalizeNumber(Number(logData.retention_days), 1, 365) || 7;

    if (proxyTimeoutRestartRequired.value) {
      showToastNotice(
        `网络与日志配置已保存；日志保留 ${logRetentionDays.value} 天。TMDB 请求超时已即时生效，重启后端可同步外层请求超时`,
        "success",
      );
      return;
    }

    showToastNotice(
      proxyEnabled.value
        ? `网络与日志配置已保存；日志保留 ${logRetentionDays.value} 天`
        : `代理已关闭为直连；日志保留 ${logRetentionDays.value} 天`,
      proxyEnabled.value ? "success" : "info",
    );
  } catch {
    // 错误已由全局请求拦截器提示。
  } finally {
    proxySaving.value = false;
    logSaving.value = false;
  }
}

async function saveAutoSyncSettings() {
  syncSaving.value = true;
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
    showToastNotice(
      syncEnabled.value ? "自动同步配置已保存并生效" : "自动同步已关闭",
      syncEnabled.value ? "success" : "info",
    );
  } catch {
    // errors shown via global toast
  } finally {
    syncSaving.value = false;
  }
}

async function triggerAutoSyncNow() {
  syncTriggering.value = true;

  try {
    const resp = await runAutoSyncNow();
    const data = resp.data;
    syncRunning.value = !!data.running;
    showToastNotice(data.message || "已触发一次立即同步任务", "success");
    await loadAutoSyncLogs();
  } catch {
    // errors shown via global toast
  } finally {
    syncTriggering.value = false;
  }
}

async function reloadAll() {
  await Promise.all([loadSettings(), loadAutoSyncLogs()]);
}

onMounted(reloadAll);
</script>

<template>
  <section class="grid min-w-0 gap-4">
    <section class="settings-toolbar card">
      <div class="min-w-0">
        <p class="section-label">系统设置</p>
        <h2 class="library-toolbar-title">运行配置</h2>
        <p class="mt-1 text-sm text-muted">统一管理 TMDB 网络代理、日志保留和库内定时同步任务。</p>
      </div>

      <div class="flex items-center gap-3">
        <span class="badge">{{ taskRunStatusText }}</span>
        <button class="btn-soft-xs disabled:opacity-60" :disabled="settingsBusy" @click="reloadAll">
          {{ loading || logsLoading ? "读取中..." : "重新读取" }}
        </button>
      </div>
    </section>

    <LoadState
      class="grid min-w-0 gap-4"
      :loading="initialLoading"
      loading-text="系统设置加载中..."
    >
      <section class="settings-summary-grid">
        <article class="settings-summary-card">
          <span class="settings-summary-label">代理访问</span>
          <strong>{{ proxyStatusText }} · {{ proxyLocalWriteStatusText }}</strong>
          <p>{{ proxyEnabled ? proxyURL || "等待代理地址" : "直连 TMDB" }}</p>
        </article>
        <article class="settings-summary-card">
          <span class="settings-summary-label">日志保留</span>
          <strong>{{ logRetentionStatusText }}</strong>
          <p>访问 / 上游 / 自动同步日志</p>
        </article>
        <article class="settings-summary-card">
          <span class="settings-summary-label">自动同步</span>
          <strong>{{ syncStatusText }}</strong>
          <p>{{ syncEnabled ? `${syncCronExpr} · ${formatMode(syncMode)}` : "未启用调度" }}</p>
        </article>
        <article class="settings-summary-card">
          <span class="settings-summary-label">任务状态</span>
          <strong>{{ taskRunStatusText }}</strong>
          <p>批大小 {{ syncBatchSize }} · 延迟 {{ syncStartDelaySecond }}s</p>
        </article>
        <article class="settings-summary-card">
          <span class="settings-summary-label">最近执行</span>
          <strong>{{ latestLogStatusText }}</strong>
          <p>{{ latestLogTimeText }}</p>
        </article>
        <article class="settings-summary-card">
          <span class="settings-summary-label">当前版本</span>
          <strong>v{{ appVersion || "-" }}</strong>
          <p>前端构建版本</p>
        </article>
      </section>

      <section class="settings-form-grid">
      <div class="card settings-card">
        <div class="settings-panel-header">
          <div>
            <p class="section-label">Network</p>
            <h3 class="settings-section-title">代理与日志</h3>
            <p class="settings-note">配置 TMDB 网络代理，并统一控制三类日志的自动清理周期。</p>
          </div>
        </div>

        <label class="settings-toggle-row">
          <input v-model="proxyEnabled" type="checkbox" class="check-control" />
          <span>
            <strong>启用代理访问 TMDB</strong>
            <small>关闭后恢复为直连，保存后即时生效。</small>
          </span>
        </label>

        <label class="settings-toggle-row">
          <input v-model="proxyLocalWriteEnabled" type="checkbox" class="check-control" :disabled="networkSaving" />
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
            :disabled="!proxyEnabled || networkSaving"
            placeholder="http://127.0.0.1:7890"
          />
        </label>
        <p class="settings-help-text">支持格式示例：http://127.0.0.1:7890、socks5://127.0.0.1:1080</p>

        <label class="settings-field-label">
          对外访问地址
          <input
            v-model="proxyPublicBaseURL"
            type="text"
            class="field-control mt-1 w-full text-sm"
            :disabled="networkSaving"
            placeholder="http://192.168.50.10:18080"
          />
          <span>用于把本地上传图片 /uploads/xxx 改写为完整 URL，供 media-saber 等外部系统跨域访问；留空保持同源相对路径。</span>
        </label>

        <label class="settings-field-label">
          请求超时（毫秒）
          <input
            v-model.number="proxyTimeout"
            type="number"
            min="1000"
            max="300000"
            step="1000"
            class="field-control mt-1 w-full text-sm"
            :disabled="networkSaving"
          />
          <span>可设置 1000-300000 毫秒；TMDB 请求超时保存后即时生效，外层请求处理超时重启后同步。</span>
        </label>
        <p v-if="proxyTimeoutRestartRequired" class="settings-feedback settings-feedback-warning">
          当前外层请求处理超时配置已变更，重启后端后完全同步。
        </p>

        <label class="settings-field-label">
          日志保留天数
          <input
            v-model.number="logRetentionDays"
            type="number"
            min="1"
            max="365"
            class="field-control mt-1 w-full text-sm"
            :disabled="networkSaving"
          />
          <span>读取当前运行配置；可设置 1-365 天，默认 7 天。覆盖访问日志、上游请求日志与自动同步日志。</span>
        </label>

        <div class="settings-card-actions">
          <button class="btn-primary disabled:opacity-60" :disabled="networkSaving" @click="saveNetworkSettings">
            {{ networkSaving ? "保存中..." : "保存设置" }}
          </button>
        </div>
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
      </div>
      </section>
    </LoadState>

    <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
  </section>
</template>
