<script setup lang="ts">
import BaseDialog from "@/components/common/BaseDialog.vue";
import type { LibraryListItem } from "@/components/library/types";

defineProps<{
  visible: boolean;
  deletingId: number | null;
  pendingDeleteItem: LibraryListItem | null;
  onClose: () => void;
  onConfirm: () => void;
}>();
</script>

<template>
  <BaseDialog
    :visible="visible"
    title="确认删除"
    :busy="deletingId !== null"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="bg-black/65 backdrop-blur-[2px]"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-5 flex justify-end gap-2 px-5 pb-5"
    @close="onClose"
  >
    <template #title>
      <span class="text-base font-semibold text-ink">确认删除</span>
    </template>

    <p class="text-sm text-muted">
      将删除本地数据：
      <span class="font-medium text-ink">{{
        pendingDeleteItem?.title || pendingDeleteItem?.name || `ID ${pendingDeleteItem?.tmdb_id ?? ""}`
      }}</span>
    </p>
    <p class="mt-1 text-xs text-muted">删除后不可恢复。</p>

    <template #footer>
      <button class="btn-soft" :disabled="deletingId !== null" @click="onClose">取消</button>
      <button
        class="btn-danger-soft disabled:opacity-60"
        data-dialog-primary
        :disabled="deletingId !== null"
        @click="onConfirm"
      >
        {{ deletingId !== null ? "删除中..." : "确认删除" }}
      </button>
    </template>
  </BaseDialog>
</template>
