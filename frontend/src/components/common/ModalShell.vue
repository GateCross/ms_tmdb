<script setup lang="ts">
import { computed } from "vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    maxWidthClass?: string;
    contentClass?: string;
    footerClass?: string;
    variant?: "glass" | "vben";
  }>(),
  {
    maxWidthClass: "max-w-5xl",
    contentClass: "modal-scroll-content max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6",
    footerClass: "",
    variant: "glass",
  },
);

const emit = defineEmits<{
  close: [];
}>();

const panelClass = computed(() =>
  props.variant === "vben"
    ? "bg-[var(--surface)] text-[var(--text-main)] border border-[var(--border-muted)] shadow-[0_12px_28px_rgba(0,0,0,0.24),0_24px_56px_rgba(0,0,0,0.28)]"
    : "panel-glass",
);

const headerClass = computed(() =>
  props.variant === "vben"
    ? "sticky top-0 z-10 flex min-h-[48px] items-center justify-between border-b border-[var(--border-muted)] bg-[var(--surface)] px-5 py-0"
    : "sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/10 bg-black/35 px-4 py-3 backdrop-blur sm:px-6",
);

const resolvedFooterClass = computed(() => {
  if (props.variant === "vben") {
    return cn(
      "flex items-center justify-end gap-2 border-t border-[var(--border-muted)] bg-[var(--surface)] px-5 py-3",
      props.footerClass,
    );
  }
  return props.footerClass;
});

const titleClass = computed(() =>
  props.variant === "vben"
    ? "text-[15px] font-semibold leading-[1.4] text-[var(--text-main)]"
    : "text-sm font-semibold",
);

const closeButtonClass = computed(() =>
  props.variant === "vben"
    ? "inline-flex h-[30px] w-[30px] items-center justify-center rounded-md border-0 bg-transparent text-[22px] leading-none text-[var(--text-muted)] transition-colors hover:bg-[var(--surface-muted)] hover:text-[var(--text-main)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-ring)] disabled:opacity-60"
    : "btn-soft px-3 py-1.5 text-xs disabled:opacity-60",
);

const closeButtonText = computed(() => (props.variant === "vben" ? "×" : "关闭"));
</script>

<template>
  <BaseDialog
    :visible="visible"
    :title="title"
    :max-width-class="maxWidthClass"
    :content-class="contentClass"
    :footer-class="resolvedFooterClass"
    :panel-class="panelClass"
    :header-class="headerClass"
    :close-button-class="closeButtonClass"
    :close-button-text="closeButtonText"
    @close="emit('close')"
  >
    <template #title>
      <span :class="titleClass">{{ title }}</span>
    </template>

    <slot />

    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </BaseDialog>
</template>
