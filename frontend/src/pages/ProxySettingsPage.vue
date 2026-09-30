<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { getProxySettings, updateProxySettings } from "@/api/admin";
import LoadState from "@/components/common/LoadState.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import { useToastNotice } from "@/composables/useToastNotice";
import { resolveErrorMessage } from "@/utils/errors";

const loading = ref(false);
const settingsLoaded = ref(false);
const loadError = ref("");
const refreshError = ref("");
const saving = ref(false);
const proxyURL = ref("");
const enabled = ref(false);
const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();
const initialLoading = computed(() => loading.value && !settingsLoaded.value);

function normalizeProxyURL(raw: string) {
  return raw.trim();
}

async function loadSettings() {
  const hadData = settingsLoaded.value;
  loading.value = true;
  loadError.value = "";
  refreshError.value = "";
  try {
    const resp = await getProxySettings();
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    enabled.value = !!data.enabled;
    settingsLoaded.value = true;
    loadError.value = "";
    refreshError.value = "";
  } catch (error) {
    const message = resolveErrorMessage(error, "请求失败，请重试");
    if (hadData) {
      refreshError.value = message;
      loadError.value = "";
    } else {
      loadError.value = message;
      refreshError.value = "";
    }
  } finally {
    loading.value = false;
  }
}

async function saveSettings() {
  saving.value = true;
  try {
    const nextProxyURL = enabled.value ? normalizeProxyURL(proxyURL.value) : "";
    const resp = await updateProxySettings({ proxy_url: nextProxyURL });
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    enabled.value = !!data.enabled;
    showToastNotice(enabled.value ? "代理已启用" : "代理已关闭，当前为直连", enabled.value ? "success" : "info");
  } catch {
    // 错误已由全局请求拦截器提示，这里只恢复保存状态。
  } finally {
    saving.value = false;
  }
}

onMounted(loadSettings);
</script>

<template>
  <section class="card max-w-2xl">
    <h2 class="text-lg font-semibold">代理设置</h2>
    <p class="mt-1 text-sm text-muted">配置后端访问 TMDB 时使用的网络代理。关闭后将恢复为直连。</p>

    <div
      v-if="refreshError && !loadError"
      class="logs-refresh-error mt-4"
      role="status"
      aria-live="polite"
    >
      <span>刷新失败：{{ refreshError }}</span>
      <button type="button" class="btn-soft-xs" :disabled="loading || saving" @click="loadSettings">重试</button>
    </div>

    <LoadState
      class="mt-4"
      :loading="initialLoading"
      loading-text="代理设置加载中..."
    >
      <label class="inline-flex items-center gap-2 text-sm">
        <input v-model="enabled" type="checkbox" class="check-control" />
        <span>启用代理访问 TMDB</span>
      </label>

      <label class="mt-3 block text-xs text-muted">
        代理地址
        <input
          v-model="proxyURL"
          type="text"
          class="field-control mt-1 w-full text-sm"
          :disabled="!enabled || saving"
          placeholder="http://127.0.0.1:7890"
        />
      </label>

      <p class="mt-2 text-xs text-muted">支持格式示例：`http://127.0.0.1:7890`、`socks5://127.0.0.1:1080`</p>

      <div class="mt-4 flex items-center gap-3">
        <button class="btn-primary disabled:opacity-60" :disabled="saving" @click="saveSettings">
          {{ saving ? "保存中..." : "保存设置" }}
        </button>
        <button class="btn-soft disabled:opacity-60" :disabled="saving || loading" @click="loadSettings">
          重新读取
        </button>
      </div>
    </LoadState>

    <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
  </section>
</template>
