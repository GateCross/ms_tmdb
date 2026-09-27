<script setup lang="ts">
import { computed } from "vue";
import {
  DialogRoot,
  DialogPortal,
  DialogOverlay,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from "radix-vue";

const props = withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    description?: string;
    /** 提交中禁止关闭，优先于 closeOnEscape / closeOnOverlay */
    busy?: boolean;
    closeOnOverlay?: boolean;
    closeOnEscape?: boolean;
    initialFocus?: "close" | "primary" | "first";
    showCloseButton?: boolean;
    closeButtonClass?: string;
    closeButtonText?: string;
    maxWidthClass?: string;
    panelClass?: string;
    headerClass?: string;
    contentClass?: string;
    footerClass?: string;
    overlayClass?: string;
    rootClass?: string;
  }>(),
  {
    description: undefined,
    busy: false,
    closeOnOverlay: true,
    closeOnEscape: true,
    initialFocus: "first",
    showCloseButton: true,
    closeButtonClass: "btn-soft px-3 py-1.5 text-xs disabled:opacity-60",
    closeButtonText: "关闭",
    maxWidthClass: "max-w-5xl",
    panelClass: "panel-glass",
    headerClass:
      "sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/10 bg-black/35 px-4 py-3 backdrop-blur sm:px-6",
    contentClass: "modal-scroll-content max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6",
    footerClass: "",
    overlayClass: "fixed inset-0 z-[1300] bg-black/60 backdrop-blur-[2px]",
    rootClass: "fixed inset-0 z-[1300] flex items-center justify-center p-3 sm:p-6",
  },
);

const emit = defineEmits<{
  close: [];
}>();

const open = computed({
  get: () => props.visible,
  set: (val: boolean) => {
    if (!val && !props.busy) {
      emit("close");
    }
  },
});

function handleInteractOutside(event: Event) {
  if (props.busy || !props.closeOnOverlay) {
    event.preventDefault();
  }
}

function handleEscapeKeyDown(event: KeyboardEvent) {
  if (props.busy || !props.closeOnEscape) {
    event.preventDefault();
  }
}

function handleOpenAutoFocus(event: Event) {
  if (props.initialFocus !== "primary") {
    // close / first 交给 radix 默认初始焦点
    return;
  }
  // 查询限定在当前弹窗容器内，避免叠开弹窗时命中下层弹窗的按钮
  const el = (event.currentTarget as HTMLElement | null)?.querySelector<HTMLElement>("[data-dialog-primary]") ?? null;
  if (el) {
    event.preventDefault();
    el.focus();
  }
}
</script>

<template>
  <DialogRoot v-model:open="open">
    <!-- radix 的 Portal 不随 open 卸载，须手动门控，否则全屏容器关闭后残留挡住页面点击 -->
    <DialogPortal v-if="open">
      <DialogOverlay :class="overlayClass" />
      <div :class="rootClass">
        <DialogContent
          :class="[panelClass, 'relative z-10 w-full overflow-hidden rounded-lg outline-none', maxWidthClass]"
          @pointer-down-outside="handleInteractOutside"
          @focus-outside="handleInteractOutside"
          @escape-key-down="handleEscapeKeyDown"
          @open-auto-focus="handleOpenAutoFocus"
        >
          <header :class="headerClass">
            <div class="min-w-0">
              <DialogTitle as-child>
                <slot name="title">
                  <h3 class="min-w-0 truncate text-sm font-semibold">
                    {{ title }}
                  </h3>
                </slot>
              </DialogTitle>
              <DialogDescription v-if="description" class="mt-0.5 text-xs text-black/60 dark:text-slate-400">
                {{ description }}
              </DialogDescription>
            </div>

            <DialogClose v-if="showCloseButton" as-child :disabled="busy">
              <button type="button" :class="closeButtonClass" :disabled="busy" aria-label="关闭">
                {{ closeButtonText }}
              </button>
            </DialogClose>
          </header>

          <div :class="contentClass">
            <slot />
          </div>

          <div v-if="$slots.footer" :class="footerClass">
            <slot name="footer" />
          </div>
        </DialogContent>
      </div>
    </DialogPortal>
  </DialogRoot>
</template>
