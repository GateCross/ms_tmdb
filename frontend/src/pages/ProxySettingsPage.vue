<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getProxySettings, updateProxySettings } from "@/api/admin";

const loading = ref(false);
const saving = ref(false);
const error = ref("");
const message = ref("");
const proxyURL = ref("");
const enabled = ref(false);

function normalizeProxyURL(raw: string) {
  return raw.trim();
}

async function loadSettings() {
  loading.value = true;
  error.value = "";
  message.value = "";
  try {
    const resp = await getProxySettings();
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    enabled.value = !!data.enabled;
  } catch (err: any) {
    error.value = err.message ?? "读取代理设置失败";
  } finally {
    loading.value = false;
  }
}

async function saveSettings() {
  saving.value = true;
  error.value = "";
  message.value = "";
  try {
    const nextProxyURL = enabled.value ? normalizeProxyURL(proxyURL.value) : "";
    const resp = await updateProxySettings({ proxy_url: nextProxyURL });
    const data = resp.data;
    proxyURL.value = data.proxy_url ?? "";
    enabled.value = !!data.enabled;
    message.value = enabled.value ? "代理已启用" : "代理已关闭，当前为直连";
  } catch (err: any) {
    error.value = err.message ?? "保存代理设置失败";
  } finally {
    saving.value = false;
  }
}

onMounted(loadSettings);
</script>

<template>
  <section class="card max-w-2xl">
    <h2 class="text-lg font-semibold">代理设置</h2>
    <p class="mt-1 text-sm text-black/60">
      配置后端访问 TMDB 时使用的网络代理。关闭后将恢复为直连。
    </p>

    <p v-if="loading" class="mt-4 text-sm text-black/55">加载中...</p>

    <template v-else>
      <label class="mt-4 inline-flex items-center gap-2 text-sm">
        <input v-model="enabled" type="checkbox" class="check-control" />
        <span>启用代理访问 TMDB</span>
      </label>

      <label class="mt-3 block text-xs text-black/60">
        代理地址
        <input
          v-model="proxyURL"
          type="text"
          class="field-control mt-1 w-full text-sm"
          :disabled="!enabled || saving"
          placeholder="http://127.0.0.1:7890"
        />
      </label>

      <p class="mt-2 text-xs text-black/50">
        支持格式示例：`http://127.0.0.1:7890`、`socks5://127.0.0.1:1080`
      </p>

      <div class="mt-4 flex items-center gap-3">
        <button
          class="btn-primary disabled:opacity-60"
          :disabled="saving"
          @click="saveSettings"
        >
          {{ saving ? "保存中..." : "保存设置" }}
        </button>
        <button
          class="btn-soft disabled:opacity-60"
          :disabled="saving"
          @click="loadSettings"
        >
          重新读取
        </button>
      </div>
    </template>

    <p v-if="message" class="mt-3 text-sm text-green-700">{{ message }}</p>
    <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
  </section>
</template>
