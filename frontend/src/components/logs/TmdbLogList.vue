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
  upstreamQuery,
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

// 查询参数是路径列的主要可变内容，内联展示填补宽屏空档；无参数返回空串
function pathQuery(item: AdminTmdbRequestLogItem): string {
  const query = upstreamQuery(item.url);
  return query.startsWith("?") ? query : "";
}

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
          <BaseTooltip :content="`${item.path || '-'}${upstreamQueryFull(item.url)}`">
            <code>{{ item.path || "-" }}<span class="logs-path-query">{{ pathQuery(item) }}</span></code>
          </BaseTooltip>
        </div>
        <BaseTooltip v-if="item.error_message" :content="item.error_message">
          <p class="logs-error-line">{{ trimMiddle(item.error_message, 120) }}</p>
        </BaseTooltip>
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
