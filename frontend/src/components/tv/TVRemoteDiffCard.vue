<script setup lang="ts">
import { computed } from "vue";
import DetailSyncPanel from "@/components/DetailSyncPanel.vue";
import type { AdminSyncMode } from "@/api/admin";
import { resolveMediaFieldLabel, summarizeMediaFields } from "@/utils/mediaFieldLabels";
import type { RemoteDiffDecision, RemoteDiffNotice } from "./types";

const props = defineProps<{
  targetId: number;
  checkingRemoteDiff: boolean;
  remoteDiffNotice: RemoteDiffNotice | null;
  remoteDiffMessage: string;
  remoteDiffDecision: RemoteDiffDecision;
  showRemoteDiffDetails: boolean;
  showLocalOverrideDiffDetails: boolean;
  shouldShowSyncPanel: boolean;
  allowedSyncModes: AdminSyncMode[];
  onToggleRemoteDetails: () => void;
  onToggleLocalDetails: () => void;
  onKeepLocal: () => void;
  onSynced: () => void;
}>();

const isCompactStatusOnly = computed(() => {
  return (
    !props.checkingRemoteDiff &&
    !props.remoteDiffNotice &&
    !props.shouldShowSyncPanel &&
    (props.remoteDiffDecision === "no_diff" || Boolean(props.remoteDiffMessage))
  );
});
</script>

<template>
  <div v-if="isCompactStatusOnly" class="detail-alert-compact">
    <span class="text-success">
      {{ remoteDiffDecision === "no_diff" ? "已完成检查，当前未发现远程差异。" : remoteDiffMessage }}
    </span>
  </div>

  <div
    v-else-if="checkingRemoteDiff || remoteDiffNotice || shouldShowSyncPanel"
    class="detail-alert"
  >
    <p v-if="checkingRemoteDiff" class="text-xs text-warn">正在检测远程数据差异...</p>

    <template v-else-if="remoteDiffNotice">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="detail-alert-title">检测到远程剧集数据与本地不一致</p>
          <p class="detail-alert-text">远程变化字段：{{ summarizeMediaFields(remoteDiffNotice.remoteFields) }}</p>
          <p class="detail-alert-text">本地修改字段：{{ summarizeMediaFields(remoteDiffNotice.localOverrideFields) }}</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            v-if="remoteDiffNotice.remoteDetails.length"
            type="button"
            class="btn-soft-xs"
            @click="onToggleRemoteDetails"
          >
            {{ showRemoteDiffDetails ? "收起远程变化明细" : "查看远程变化明细" }}
          </button>
          <button
            v-if="remoteDiffNotice.localOverrideDetails.length"
            type="button"
            class="btn-soft-xs"
            @click="onToggleLocalDetails"
          >
            {{ showLocalOverrideDiffDetails ? "收起本地修改明细" : "查看本地修改明细" }}
          </button>
          <button type="button" class="btn-soft-xs disabled:opacity-60" @click="onKeepLocal">
            暂不处理，保留本地
          </button>
        </div>
      </div>

      <div v-if="showRemoteDiffDetails && remoteDiffNotice.remoteDetails.length" class="detail-diff-list">
        <div v-for="item in remoteDiffNotice.remoteDetails" :key="`remote-${item.field}`" class="detail-diff-item">
          <div class="flex items-baseline justify-between gap-2">
            <span class="text-xs font-semibold text-strong">{{ resolveMediaFieldLabel(item.field) }}</span>
            <span class="font-mono text-[11px] text-muted">{{ item.field }}</span>
          </div>
          <div class="detail-diff-grid">
            <div class="detail-diff-col">
              <span class="block text-[11px] text-muted">本地值</span>
              <p class="mt-0.5 break-words text-ink">{{ item.local }}</p>
            </div>
            <div class="detail-diff-col detail-diff-col-remote">
              <span class="block text-[11px] text-warn">TMDB 远程</span>
              <p class="mt-0.5 break-words text-warn">{{ item.remote }}</p>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showLocalOverrideDiffDetails && remoteDiffNotice.localOverrideDetails.length" class="detail-diff-list">
        <div
          v-for="item in remoteDiffNotice.localOverrideDetails"
          :key="`local-${item.field}`"
          class="detail-diff-item"
        >
          <div class="flex items-baseline justify-between gap-2">
            <span class="text-xs font-semibold text-strong">{{ resolveMediaFieldLabel(item.field) }}</span>
            <span class="font-mono text-[11px] text-muted">{{ item.field }}</span>
          </div>
          <div class="detail-diff-grid">
            <div class="detail-diff-col">
              <span class="block text-[11px] text-muted">本地值</span>
              <p class="mt-0.5 break-words text-ink">{{ item.local }}</p>
            </div>
            <div class="detail-diff-col detail-diff-col-remote">
              <span class="block text-[11px] text-warn">TMDB 远程</span>
              <p class="mt-0.5 break-words text-warn">{{ item.remote }}</p>
            </div>
          </div>
        </div>
      </div>
    </template>

    <DetailSyncPanel
      v-if="shouldShowSyncPanel"
      media-type="tv"
      :target-id="targetId"
      :allowed-modes="allowedSyncModes"
      :preset-changed-fields="remoteDiffNotice?.localOverrideFields ?? []"
      :embedded="true"
      @synced="onSynced"
    />
  </div>
</template>
