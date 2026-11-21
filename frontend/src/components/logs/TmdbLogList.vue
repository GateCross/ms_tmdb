<script setup lang="ts">
import type { AdminTmdbRequestLogItem } from "@/api/admin";
import DataListShell, { type DataListColumn } from "@/components/common/DataListShell.vue";
import {
  bodyMeta,
  cleanLogPath,
  formatDateTimeParts,
  formatDuration,
  formatStatusCode,
  requestRowToneClass,
  statusDotClass,
  trimMiddle,
  upstreamQueryFull,
} from "@/utils/logFormatters";
import BaseTooltip from "@/components/common/BaseTooltip.vue";

defineProps<{
  items: AdminTmdbRequestLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminTmdbRequestLogItem];
}>();

const columns: DataListColumn[] = [
  "时间",
  "上游路径",
  "状态",
  { label: "耗时", align: "right" },
  { label: "响应正文", align: "right" },
  { label: "操作", align: "right" },
];
</script>

<template>
  <DataListShell
    grid-class="logs-grid-tmdb"
    :columns="columns"
    :loading="loading"
    :empty="items.length === 0"
    loading-text="日志加载中..."
  >
    <article
      v-for="item in items"
      :key="item.id"
      class="logs-row logs-grid-tmdb"
      :class="requestRowToneClass(item.status_code)"
      @click="emit('open-detail', item)"
    >
      <time class="logs-time">
        <span class="logs-time-date">{{ formatDateTimeParts(item.created_at).date }}</span>
        <span>{{ formatDateTimeParts(item.created_at).time }}</span>
      </time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">{{ item.method }}</span>
          <BaseTooltip :content="`${cleanLogPath(item.path)}${upstreamQueryFull(item.url)}`">
            <code>{{ cleanLogPath(item.path) }}</code>
          </BaseTooltip>
        </div>
        <BaseTooltip v-if="item.error_message" :content="item.error_message">
          <p class="logs-error-line">{{ trimMiddle(item.error_message, 120) }}</p>
        </BaseTooltip>
      </div>

      <div class="logs-status-cell">
        <span class="log-status">
          <i class="log-status-dot" :class="statusDotClass(item.status_code)" aria-hidden="true"></i>
          {{ formatStatusCode(item.status_code) }}
        </span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-body-cell">
        <span>{{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click.stop="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
