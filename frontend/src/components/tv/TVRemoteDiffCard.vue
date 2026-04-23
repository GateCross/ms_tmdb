<script setup lang="ts">
import DetailSyncPanel from "@/components/DetailSyncPanel.vue";
import type { AdminSyncMode } from "@/api/admin";
import type { RemoteDiffDecision, RemoteDiffNotice } from "./types";

const props = defineProps<{
  targetId: number;
  checkingRemoteDiff: boolean;
  remoteDiffNotice: RemoteDiffNotice | null;
  remoteDiffMessage: string;
  remoteDiffError: string;
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
</script>

<template>
  <div
    v-if="checkingRemoteDiff || remoteDiffNotice || remoteDiffMessage || remoteDiffError || remoteDiffDecision === 'no_diff'"
    class="mt-4 rounded-xl border border-amber-200 bg-amber-50/80 p-4"
  >
    <p v-if="checkingRemoteDiff" class="text-xs text-amber-700">
      正在检测远程数据差异...
    </p>

    <template v-else-if="remoteDiffNotice">
      <p class="mt-3 text-sm font-medium text-amber-800">
        检测到远程剧集数据与本地不一致
      </p>
      <p class="mt-1 text-xs text-amber-700">
        远程变化字段：{{ remoteDiffNotice.remoteSummary }}
      </p>
      <p class="mt-1 text-xs text-amber-700">
        本地修改字段：{{ remoteDiffNotice.localOverrideSummary }}
      </p>
      <div class="mt-2 flex flex-wrap items-center gap-2">
        <button
          v-if="remoteDiffNotice.remoteDetails.length"
          class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100"
          @click="onToggleRemoteDetails"
        >
          {{ showRemoteDiffDetails ? "收起远程变化明细" : "查看远程变化明细" }}
        </button>
        <button
          v-if="remoteDiffNotice.localOverrideDetails.length"
          class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100"
          @click="onToggleLocalDetails"
        >
          {{ showLocalOverrideDiffDetails ? "收起本地修改明细" : "查看本地修改明细" }}
        </button>
        <button
          class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100 disabled:opacity-60"
          @click="onKeepLocal"
        >
          暂不处理，保留本地
        </button>
      </div>

      <div
        v-if="showRemoteDiffDetails && remoteDiffNotice.remoteDetails.length"
        class="mt-2 space-y-2 rounded-lg border border-amber-200 bg-white/70 p-2"
      >
        <div
          v-for="item in remoteDiffNotice.remoteDetails"
          :key="`remote-${item.field}`"
          class="rounded-md bg-amber-50/70 p-2"
        >
          <p class="text-xs font-semibold text-amber-900">{{ item.field }}</p>
          <p class="mt-1 text-xs text-amber-800">本地：{{ item.local }}</p>
          <p class="mt-1 text-xs text-amber-800">远程：{{ item.remote }}</p>
        </div>
      </div>

      <div
        v-if="showLocalOverrideDiffDetails && remoteDiffNotice.localOverrideDetails.length"
        class="mt-2 space-y-2 rounded-lg border border-amber-200 bg-white/70 p-2"
      >
        <div
          v-for="item in remoteDiffNotice.localOverrideDetails"
          :key="`local-${item.field}`"
          class="rounded-md bg-amber-50/70 p-2"
        >
          <p class="text-xs font-semibold text-amber-900">{{ item.field }}</p>
          <p class="mt-1 text-xs text-amber-800">本地：{{ item.local }}</p>
          <p class="mt-1 text-xs text-amber-800">远程：{{ item.remote }}</p>
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

    <p v-if="!checkingRemoteDiff && !remoteDiffNotice && remoteDiffDecision === 'no_diff'" class="mt-3 text-xs text-green-700">
      已完成检查，当前未发现远程差异。
    </p>
    <p v-if="!checkingRemoteDiff && !remoteDiffNotice && remoteDiffMessage" class="mt-3 text-xs text-green-700">
      {{ remoteDiffMessage }}
    </p>
    <p v-if="remoteDiffError" class="mt-1 text-xs text-red-600">
      {{ remoteDiffError }}
    </p>
  </div>
</template>
