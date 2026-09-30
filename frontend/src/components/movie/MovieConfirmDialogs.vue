<script setup lang="ts">
import BaseDialog from "@/components/common/BaseDialog.vue";

defineProps<{
  tmdbRiskModalVisible: boolean;
  tmdbRiskCurrentId: number | null;
  tmdbRiskNextId: number | null;
  deleteConfirmModalVisible: boolean;
  deleting: boolean;
  movieTitle: string;
  onCloseTmdbRisk: (confirmed: boolean) => void;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
}>();
</script>

<template>
  <BaseDialog
    :visible="tmdbRiskModalVisible"
    title="修改 TMDB ID 风险确认"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="absolute inset-0 bg-black/45"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-4 flex items-center justify-end gap-2 px-5 pb-5"
    @close="onCloseTmdbRisk(false)"
  >
    <template #title>
      <span class="text-base font-semibold text-warn">修改 TMDB ID 风险确认</span>
    </template>

    <p class="text-sm text-muted">
      你正在修改电影 TMDB ID：
      <span class="font-medium">{{ tmdbRiskCurrentId }}</span>
      ->
      <span class="font-medium">{{ tmdbRiskNextId }}</span>
    </p>
    <div class="mt-3 rounded-lg border border-warn-line bg-warn-soft p-3 text-xs leading-relaxed text-warn">
      <p>1) 这是高风险操作，可能导致与第三方历史引用不一致；</p>
      <p>2) 之后自动/手动同步将继续使用旧 TMDB ID 向 TMDB 拉取；</p>
      <p>3) 对外返回与页面访问将使用新的 TMDB ID。</p>
    </div>

    <template #footer>
      <button class="btn-soft" @click="onCloseTmdbRisk(false)">取消</button>
      <button class="btn-primary" data-dialog-primary @click="onCloseTmdbRisk(true)">确认继续</button>
    </template>
  </BaseDialog>

  <BaseDialog
    :visible="deleteConfirmModalVisible"
    title="删除本地数据确认"
    :busy="deleting"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="absolute inset-0 bg-black/45"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-4 flex items-center justify-end gap-2 px-5 pb-5"
    @close="onCloseDeleteConfirm"
  >
    <template #title>
      <span class="text-base font-semibold text-danger">删除本地数据确认</span>
    </template>

    <p class="text-sm text-muted">
      确认删除电影
      <span class="font-medium">{{ movieTitle }}</span>
      的本地数据吗？
    </p>
    <p class="mt-2 text-xs text-danger">删除后不可恢复。</p>

    <template #footer>
      <button class="btn-soft" :disabled="deleting" @click="onCloseDeleteConfirm">取消</button>
      <button class="btn-danger-soft" data-dialog-primary :disabled="deleting" @click="onConfirmDelete">
        {{ deleting ? "删除中..." : "确认删除" }}
      </button>
    </template>
  </BaseDialog>
</template>
