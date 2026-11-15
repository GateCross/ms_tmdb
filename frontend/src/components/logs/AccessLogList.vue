<script setup lang="ts">
import type { AdminProxyAccessLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
import {
  accessPath,
  accessQueryFull,
  bodyMeta,
  cleanLogPath,
  formatDateTimeParts,
  formatDuration,
  formatStatusCode,
  requestRowToneClass,
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

const columns = ["时间", "请求", "状态", "耗时", "正文", "来源", "操作"];
</script>

<template>
  <DataListShell grid-class="logs-grid-access" :columns="columns" :loading="loading" loading-text="日志加载中...">
    <article
      v-for="item in items"
      :key="item.id"
      class="logs-row logs-grid-access"
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
          <BaseTooltip
            :content="`${cleanLogPath(item.path || accessPath(item.request_uri))}${accessQueryFull(item.request_uri)}`"
          >
            <code>{{ cleanLogPath(item.path || accessPath(item.request_uri)) }}</code>
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
        <span
          >响应 {{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }} · 请求
          {{ bodyMeta(item.request_body_bytes, item.request_body_truncated) }}</span
        >
      </div>

      <div class="logs-source">
        <strong>{{ item.client_ip || "-" }}</strong>
        <BaseTooltip :content="item.user_agent || '-'">
          <span>{{ item.user_agent || "-" }}</span>
        </BaseTooltip>
      </div>

      <button class="btn-soft-xs logs-action" type="button" @click.stop="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
