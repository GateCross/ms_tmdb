<script setup lang="ts">
import type { AdminAutoSyncLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
import {
  autoSyncStatusDotClass,
  formatAutoSyncStatus,
  formatDateTime,
  formatDateTimeParts,
  formatDuration,
  formatMode,
  summarizeMessage,
} from "@/utils/logFormatters";
import BaseTooltip from "@/components/common/BaseTooltip.vue";

defineProps<{
  items: AdminAutoSyncLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminAutoSyncLogItem];
}>();

// cron/批大小各行基本恒定，列表只留策略单行，完整配置收进悬浮提示与详情
function cronTitle(item: AdminAutoSyncLogItem): string {
  return `${item.cron_expr || "-"} · 批大小 ${item.batch_size}`;
}

const columns = ["时间", "策略", "状态", "耗时", "检查/同步/失败", "摘要", "操作"];
</script>

<template>
  <DataListShell
    grid-class="logs-grid-auto-sync"
    :columns="columns"
    :loading="loading"
    :empty="!loading && items.length === 0"
    empty-text="暂无执行日志"
    loading-text="日志加载中..."
  >
    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-auto-sync">
      <time class="logs-time">
        <span class="logs-time-date">{{ formatDateTimeParts(item.triggered_at).date }}</span>
        <span>{{ formatDateTimeParts(item.triggered_at).time }}</span>
      </time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">SYNC</span>
          <BaseTooltip :content="cronTitle(item)">
            <code>{{ formatMode(item.mode) }}</code>
          </BaseTooltip>
        </div>
      </div>

      <div>
        <span class="log-status">
          <i class="log-status-dot" :class="autoSyncStatusDotClass(item.status)" aria-hidden="true"></i>
          {{ formatAutoSyncStatus(item.status) }}
        </span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-body-cell">
        <span>检查 {{ item.checked }}</span>
        <small>同步 {{ item.synced }} · <span :class="item.failed > 0 ? 'logs-failed-count' : undefined">失败 {{ item.failed }}</span></small>
      </div>

      <div class="logs-source">
        <BaseTooltip :content="item.message || '-'">
          <strong>{{ summarizeMessage(item.message) }}</strong>
        </BaseTooltip>
        <span>{{ formatDateTime(item.finished_at || item.created_at) }}</span>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
