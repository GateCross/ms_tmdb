<script setup lang="ts">
import type { AdminAutoSyncLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
import {
  autoSyncRowToneClass,
  autoSyncStatusDotClass,
  formatAutoSyncStatus,
  formatDateTimeParts,
  formatDuration,
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

// 策略/cron/批大小各行基本恒定，完整配置只在详情弹窗展示，列表留纯结果列
const columns = ["时间", "状态", "检查/同步/失败", "耗时", "摘要", "操作"];
</script>

<template>
  <DataListShell grid-class="logs-grid-auto-sync" :columns="columns" :loading="loading" loading-text="日志加载中...">
    <article
      v-for="item in items"
      :key="item.id"
      class="logs-row logs-grid-auto-sync"
      :class="autoSyncRowToneClass(item.status)"
      @click="emit('open-detail', item)"
    >
      <time class="logs-time">
        <span class="logs-time-date">{{ formatDateTimeParts(item.triggered_at).date }}</span>
        <span>{{ formatDateTimeParts(item.triggered_at).time }}</span>
      </time>

      <div class="logs-status-cell">
        <span class="log-status">
          <i class="log-status-dot" :class="autoSyncStatusDotClass(item.status)" aria-hidden="true"></i>
          {{ formatAutoSyncStatus(item.status) }}
        </span>
      </div>

      <div class="logs-counts">
        <span>检查 {{ item.checked }}</span>
        <span>同步 {{ item.synced }}</span>
        <span :class="item.failed > 0 ? 'logs-failed-count' : undefined">失败 {{ item.failed }}</span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-source">
        <BaseTooltip :content="item.message || '-'">
          <strong>{{ summarizeMessage(item.message) }}</strong>
        </BaseTooltip>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click.stop="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
