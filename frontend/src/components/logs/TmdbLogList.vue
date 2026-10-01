<script setup lang="ts">
import type { AdminTmdbRequestLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
import {
  bodyMeta,
  formatDateTimeParts,
  formatDuration,
  formatStatusCode,
  statusDotClass,
  trimMiddle,
} from "@/utils/logFormatters";

defineProps<{
  items: AdminTmdbRequestLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminTmdbRequestLogItem];
}>();

const columns = ["时间", "上游路径", "状态", "耗时", "响应正文", "操作"];
</script>

<template>
  <DataListShell
    grid-class="logs-grid-tmdb"
    :columns="columns"
    :loading="loading"
    :empty="!loading && items.length === 0"
    empty-text="暂无 TMDB 请求日志"
    loading-text="日志加载中..."
  >
    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-tmdb">
      <time class="logs-time">
        <span class="logs-time-date">{{ formatDateTimeParts(item.created_at).date }}</span>
        <span>{{ formatDateTimeParts(item.created_at).time }}</span>
      </time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">{{ item.method }}</span>
          <!-- 域名各行恒定，收进悬浮提示；完整 URL 在详情弹窗 -->
          <code :title="item.url || item.path || '-'">{{ item.path || "-" }}</code>
        </div>
        <p v-if="item.error_message" class="logs-error-line" :title="item.error_message">
          {{ trimMiddle(item.error_message, 120) }}
        </p>
      </div>

      <div>
        <span class="log-status">
          <i class="log-status-dot" :class="statusDotClass(item.status_code)" aria-hidden="true"></i>
          {{ formatStatusCode(item.status_code) }}
        </span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-body-cell">
        <span>{{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
