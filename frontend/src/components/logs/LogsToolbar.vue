<script setup lang="ts">
import GlassSelect from "@/components/GlassSelect.vue";

type LogTab = "access" | "tmdb" | "autoSync";
type SelectOption = {
  label: string;
  value: string;
};

defineProps<{
  activeTab: LogTab;
  busy: boolean;
  panelLabel: string;
  panelTitle: string;
  panelNote: string;
  requestStatusOptions: ReadonlyArray<SelectOption>;
  autoSyncStatusOptions: ReadonlyArray<SelectOption>;
  accessKeyword: string;
  tmdbKeyword: string;
  accessStatus: string;
  tmdbStatus: string;
  autoSyncStatus: string;
  currentKeyword: string;
}>();

const emit = defineEmits<{
  "update:accessKeyword": [value: string];
  "update:tmdbKeyword": [value: string];
  "update:accessStatus": [value: string];
  "update:tmdbStatus": [value: string];
  "update:autoSyncStatus": [value: string];
  search: [];
  "clear-keyword": [];
  "status-change": [];
  refresh: [];
  clear: [];
}>();

// 「全部」选项的哨兵值，对应查询参数的空串
const ALL_STATUS = "all";

function toSelectValue(status: string) {
  return status || ALL_STATUS;
}

function updateAccessKeyword(event: Event) {
  emit("update:accessKeyword", (event.target as HTMLInputElement).value.trim());
}

function updateTmdbKeyword(event: Event) {
  emit("update:tmdbKeyword", (event.target as HTMLInputElement).value.trim());
}

function updateStatus(tab: LogTab, value: string) {
  const status = value === ALL_STATUS ? "" : value;
  if (tab === "access") {
    emit("update:accessStatus", status);
  } else if (tab === "tmdb") {
    emit("update:tmdbStatus", status);
  } else {
    emit("update:autoSyncStatus", status);
  }
  emit("status-change");
}
</script>

<template>
  <div class="settings-log-header">
    <div>
      <p class="section-label">{{ panelLabel }}</p>
      <h3 class="settings-section-title">{{ panelTitle }}</h3>
      <p class="settings-note">{{ panelNote }}</p>
    </div>

    <div class="settings-log-actions">
      <form v-if="activeTab !== 'autoSync'" class="settings-log-search" @submit.prevent="emit('search')">
        <label class="settings-log-filter">
          关键字
          <input
            v-if="activeTab === 'access'"
            :value="accessKeyword"
            class="field-control settings-log-search-input"
            type="search"
            placeholder="搜索路径、查询、IP 或 UA"
            :disabled="busy"
            @input="updateAccessKeyword"
          />
          <input
            v-else
            :value="tmdbKeyword"
            class="field-control settings-log-search-input"
            type="search"
            placeholder="搜索上游路径或 URL"
            :disabled="busy"
            @input="updateTmdbKeyword"
          />
        </label>
        <button class="btn-soft disabled:opacity-60" type="submit" :disabled="busy">搜索</button>
        <button
          v-if="currentKeyword"
          class="btn-soft disabled:opacity-60"
          type="button"
          :disabled="busy"
          @click="emit('clear-keyword')"
        >
          清空关键字
        </button>
      </form>

      <label class="settings-log-filter">
        状态
        <GlassSelect
          v-if="activeTab === 'access'"
          :model-value="toSelectValue(accessStatus)"
          :options="requestStatusOptions"
          :disabled="busy"
          class="min-w-[136px]"
          @update:model-value="(value) => updateStatus('access', value)"
        />
        <GlassSelect
          v-else-if="activeTab === 'tmdb'"
          :model-value="toSelectValue(tmdbStatus)"
          :options="requestStatusOptions"
          :disabled="busy"
          class="min-w-[136px]"
          @update:model-value="(value) => updateStatus('tmdb', value)"
        />
        <GlassSelect
          v-else
          :model-value="toSelectValue(autoSyncStatus)"
          :options="autoSyncStatusOptions"
          :disabled="busy"
          class="min-w-[136px]"
          @update:model-value="(value) => updateStatus('autoSync', value)"
        />
      </label>

      <button class="btn-soft disabled:opacity-60" :disabled="busy" @click="emit('refresh')">
        {{ busy ? "刷新中..." : "刷新日志" }}
      </button>
      <button class="btn-danger-soft disabled:opacity-60" :disabled="busy" @click="emit('clear')">
        {{ busy ? "清空中..." : "清空日志" }}
      </button>
    </div>
  </div>
</template>
