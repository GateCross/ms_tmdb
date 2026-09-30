<script setup lang="ts">
import BaseDialog from "@/components/common/BaseDialog.vue";
import LogsPagination from "@/components/logs/LogsPagination.vue";
import type { AdminAutoSyncLogDetailResp } from "@/api/admin";
import {
  fieldChangeCount,
  formatAutoSyncStatus,
  formatDateTime,
  formatDuration,
  formatFieldChanges,
  formatFieldList,
  formatMediaType,
  formatMode,
  hasFieldList,
  hasLocalFieldSummary,
  visibleFieldList,
} from "@/utils/logFormatters";

defineProps<{
  visible: boolean;
  loading: boolean;
  detail: AdminAutoSyncLogDetailResp | null;
  syncedPage: number;
  syncedTotalPages: number;
  failedPage: number;
  failedTotalPages: number;
}>();

const emit = defineEmits<{
  close: [];
  "change-synced-page": [page: number];
  "change-failed-page": [page: number];
}>();
</script>

<template>
  <BaseDialog
    :visible="visible"
    title="执行日志明细"
    max-width-class="max-w-6xl"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-3 sm:p-4"
    overlay-class="absolute inset-0 bg-black/55"
    panel-class="panel-glass settings-detail-modal max-h-[92vh]"
    header-class="modal-header-dark"
    content-class="settings-detail-scroll max-h-[calc(92vh-72px)] overflow-y-auto px-4 py-4 sm:px-5"
    initial-focus="close"
    @close="emit('close')"
  >
    <template #title>
      <span class="block min-w-0">
        <span class="section-label block">Run Detail</span>
        <span class="block text-base font-semibold">
          执行日志明细
          <span v-if="detail" class="text-sm text-muted">#{{ detail.id }}</span>
        </span>
      </span>
    </template>

    <p v-if="loading && !detail" class="text-sm text-muted">明细加载中...</p>

    <template v-if="detail">
      <p v-if="loading" class="mb-3 text-xs text-muted">分页加载中...</p>
      <div class="settings-detail-summary-grid">
        <article class="settings-detail-summary-item">
          <span>触发时间</span>
          <strong>{{ formatDateTime(detail.triggered_at) }}</strong>
          <small>{{ detail.cron_expr || "-" }}</small>
        </article>
        <article class="settings-detail-summary-item">
          <span>同步策略</span>
          <strong>{{ formatMode(detail.mode) }}</strong>
          <small>批大小 {{ detail.batch_size }}</small>
        </article>
        <article class="settings-detail-summary-item">
          <span>状态</span>
          <strong>{{ formatAutoSyncStatus(detail.status) }}</strong>
          <small>耗时 {{ formatDuration(detail.duration_ms) }}</small>
        </article>
        <article class="settings-detail-summary-item">
          <span>检查 / 同步 / 失败</span>
          <strong>{{ detail.checked }} / {{ detail.synced }} / {{ detail.failed }}</strong>
          <small>{{ detail.message || "-" }}</small>
        </article>
      </div>

      <div class="settings-detail-section">
        <div class="settings-detail-section-header">
          <div>
            <h5 class="text-sm font-semibold text-success">同步成功项</h5>
            <p class="settings-note">展示成功同步条目、远端差异字段和本地字段处理结果。</p>
          </div>
          <span class="badge">{{ detail.synced }} 条</span>
        </div>
        <div class="table-shell settings-table-shell">
          <table class="min-w-full text-sm settings-detail-table settings-detail-table-fixed settings-detail-success-table">
            <colgroup>
              <col class="settings-detail-col-media" />
              <col class="settings-detail-col-remote" />
              <col class="settings-detail-col-local" />
              <col class="settings-detail-col-message" />
            </colgroup>
            <thead class="table-head text-left text-muted">
              <tr>
                <th class="px-3 py-2 font-medium">媒体</th>
                <th class="px-3 py-2 font-medium">远端差异</th>
                <th class="px-3 py-2 font-medium">本地处理</th>
                <th class="px-3 py-2 font-medium">信息</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(entry, idx) in detail.synced_list"
                :key="`synced-${idx}-${entry.media_type}-${entry.tmdb_id}`"
                class="table-row-hover"
              >
                <td class="px-3 py-2">
                  <div class="settings-media-cell">
                    <span class="settings-media-type">{{ formatMediaType(entry.media_type) }}</span>
                    <div>
                      <p class="settings-table-primary line-clamp-2">{{ entry.name || "-" }}</p>
                      <p class="settings-table-meta">TMDB ID {{ entry.tmdb_id || "-" }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2">
                  <div v-if="visibleFieldList(entry.remote_diff_fields).length" class="settings-chip-list">
                    <span v-for="field in visibleFieldList(entry.remote_diff_fields)" :key="field">{{ field }}</span>
                  </div>
                  <span v-else class="settings-empty-value">-</span>
                  <details v-if="fieldChangeCount(entry.field_changes)" class="settings-field-detail">
                    <summary>字段明细 {{ fieldChangeCount(entry.field_changes) }} 项</summary>
                    <pre class="settings-diff-pre settings-diff-pre-compact">{{ formatFieldChanges(entry.field_changes) }}</pre>
                  </details>
                </td>
                <td class="px-3 py-2">
                  <div v-if="hasLocalFieldSummary(entry)" class="settings-field-stack">
                    <div v-if="hasFieldList(entry.changed_fields)">
                      <span>变更</span>
                      <p>{{ formatFieldList(entry.changed_fields) }}</p>
                    </div>
                    <div v-if="hasFieldList(entry.overwritten_fields)">
                      <span>覆盖</span>
                      <p>{{ formatFieldList(entry.overwritten_fields) }}</p>
                    </div>
                    <div v-if="hasFieldList(entry.kept_local_fields)">
                      <span>保留</span>
                      <p>{{ formatFieldList(entry.kept_local_fields) }}</p>
                    </div>
                  </div>
                  <span v-else class="settings-empty-value">-</span>
                </td>
                <td class="px-3 py-2 text-muted">
                  <p class="settings-detail-message">{{ entry.message || "-" }}</p>
                </td>
              </tr>
              <tr v-if="detail.synced_list.length === 0">
                <td colspan="4" class="px-3 py-4 text-center text-muted">无成功同步明细</td>
              </tr>
            </tbody>
          </table>
        </div>
        <LogsPagination
          :total="detail.synced"
          :page="syncedPage"
          :total-pages="syncedTotalPages"
          :busy="loading"
          small
          @change-page="(page) => emit('change-synced-page', page)"
        />
      </div>

      <div class="settings-detail-section">
        <div class="settings-detail-section-header">
          <div>
            <h5 class="text-sm font-semibold text-danger">同步失败项</h5>
            <p class="settings-note">失败条目会保留原因，便于定位网络、数据或接口异常。</p>
          </div>
          <span class="badge">{{ detail.failed }} 条</span>
        </div>
        <div class="table-shell settings-table-shell">
          <table class="min-w-full text-sm settings-detail-table settings-detail-table-fixed settings-detail-failed-table">
            <colgroup>
              <col class="settings-detail-col-media" />
              <col class="settings-detail-col-failure" />
            </colgroup>
            <thead class="table-head text-left text-muted">
              <tr>
                <th class="px-3 py-2 font-medium">媒体</th>
                <th class="px-3 py-2 font-medium">失败原因</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(entry, idx) in detail.failed_list"
                :key="`failed-${idx}-${entry.media_type}-${entry.tmdb_id}`"
                class="table-row-hover"
              >
                <td class="px-3 py-2">
                  <div class="settings-media-cell">
                    <span class="settings-media-type">{{ formatMediaType(entry.media_type) }}</span>
                    <div>
                      <p class="settings-table-primary line-clamp-2">{{ entry.name || "-" }}</p>
                      <p class="settings-table-meta">TMDB ID {{ entry.tmdb_id || "-" }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2 text-muted">
                  <p class="settings-detail-message">{{ entry.message || "-" }}</p>
                </td>
              </tr>
              <tr v-if="detail.failed_list.length === 0">
                <td colspan="2" class="px-3 py-4 text-center text-muted">无失败明细</td>
              </tr>
            </tbody>
          </table>
        </div>
        <LogsPagination
          :total="detail.failed"
          :page="failedPage"
          :total-pages="failedTotalPages"
          :busy="loading"
          small
          @change-page="(page) => emit('change-failed-page', page)"
        />
      </div>
    </template>
  </BaseDialog>
</template>
