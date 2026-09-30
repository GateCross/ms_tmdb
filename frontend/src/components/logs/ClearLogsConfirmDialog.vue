<script setup lang="ts">
import BaseDialog from "@/components/common/BaseDialog.vue";

defineProps<{
  visible: boolean;
  busy: boolean;
  label: string;
}>();

const emit = defineEmits<{
  close: [];
  confirm: [];
}>();
</script>

<template>
  <BaseDialog
    :visible="visible"
    title="确认清空日志"
    :busy="busy"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="absolute inset-0 bg-black/45"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="flex items-center justify-end gap-2 px-5 pb-5 pt-5"
    @close="emit('close')"
  >
    <template #title>
      <span class="text-base font-semibold text-danger">确认清空日志</span>
    </template>

    <p class="text-sm text-muted">将清空当前视图的{{ label }}，清空后无法恢复。</p>

    <template #footer>
      <button class="btn-soft disabled:opacity-60" :disabled="busy" @click="emit('close')">取消</button>
      <button
        class="btn-danger-soft disabled:opacity-60"
        data-dialog-primary
        :disabled="busy"
        @click="emit('confirm')"
      >
        {{ busy ? "清空中..." : "确认清空" }}
      </button>
    </template>
  </BaseDialog>
</template>
