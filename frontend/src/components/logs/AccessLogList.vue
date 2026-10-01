<script setup lang="ts">
import type { AdminProxyAccessLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
import {
  accessPath,
  accessQuery,
  accessQueryFull,
  bodyMeta,
  formatDateTimeParts,
  formatDuration,
  formatStatusCode,
  statusDotClass,
  trimMiddle,
} from "@/utils/logFormatters";
import BaseTooltip from "@/components/common/BaseTooltip.vue";

defineProps<{
  items: AdminProxyAccessLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminProxyAccessLogItem];
}>();

// 查询参数内联展示填补宽屏空档；accessQuery 已脱敏 api_key，无参数返回空串
function pathQueryString(item: AdminProxyAccessLogItem): string {
  const query = accessQuery(item.request_uri);
  return query.startsWith("?") ? query : "";
}

const columns = ["时间", "请求", "状态", "耗时", "正文", "来源", "操作"];
</script>

<template>
  <DataListShell
    grid-class="logs-grid-access"
    :columns="columns"
    :loading="loading"
    :empty="!loading && items.length === 0"
    empty-text="暂无外部访问日志"
    loading-text="日志加载中..."
  >
    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-access">
      <time class="logs-time">
        <span class="logs-time-date">{{ formatDateTimeParts(item.created_at).date }}</span>
        <span>{{ formatDateTimeParts(item.created_at).time }}</span>
      </time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">{{ item.method }}</span>
          <BaseTooltip :content="`${item.path || accessPath(item.request_uri)}${accessQueryFull(item.request_uri)}`">
            <code>{{ item.path || accessPath(item.request_uri) }}<span class="logs-path-query">{{ pathQueryString(item) }}</span></code>
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
        <span>响应 {{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
        <small>请求 {{ bodyMeta(item.request_body_bytes, item.request_body_truncated) }}</small>
      </div>

      <div class="logs-source">
        <strong>{{ item.client_ip || "-" }}</strong>
        <BaseTooltip :content="item.user_agent || '-'">
          <span>{{ item.user_agent || "-" }}</span>
        </BaseTooltip>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
