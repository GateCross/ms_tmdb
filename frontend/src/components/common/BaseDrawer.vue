<script setup lang="ts">
import { computed } from "vue";
import { DialogRoot, DialogPortal, DialogOverlay, DialogContent, DialogTitle, DialogDescription } from "radix-vue";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    description?: string;
    /** 提交中禁止关闭，优先于 closeOnEscape / closeOnOverlay */
    busy?: boolean;
    closeOnOverlay?: boolean;
    closeOnEscape?: boolean;
    showCloseButton?: boolean;
    side?: "right" | "left";
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
    showCloseButton: true,
    side: "right",
    panelClass: "admin-preference-drawer",
    headerClass: "admin-preference-header",
    contentClass: "admin-preference-options",
    footerClass: "",
    overlayClass: "admin-preference-mask",
    rootClass: "fixed inset-0 z-[80]",
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

const panelSideClass = computed(() =>
  props.side === "left"
    ? "inset-y-0 left-0 right-auto border-l-0 border-r border-[var(--border-muted)] shadow-[10px_0_28px_rgba(15,23,42,0.14)]"
    : "inset-y-0 right-0 left-auto",
);

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

function requestClose() {
  if (props.busy) return;
  emit("close");
}
</script>

<template>
  <DialogRoot v-model:open="open">
    <!-- radix 的 Portal 不随 open 卸载，须手动门控，否则全屏容器关闭后残留挡住页面点击 -->
    <DialogPortal v-if="open">
      <div :class="rootClass">
        <DialogOverlay :class="['fixed inset-0 bg-black/45 backdrop-blur-[2px] transition-opacity', overlayClass]" />

        <DialogContent
          :class="cn('fixed z-10 flex flex-col outline-none transition-transform', panelClass, panelSideClass)"
          @pointer-down-outside="handleInteractOutside"
          @focus-outside="handleInteractOutside"
          @escape-key-down="handleEscapeKeyDown"
        >
          <header :class="headerClass">
            <div class="min-w-0">
              <DialogTitle as-child>
                <slot name="title">
                  <h2>{{ title }}</h2>
                </slot>
              </DialogTitle>
              <DialogDescription v-if="description" class="mt-0.5 text-xs text-black/55 dark:text-slate-400">
                {{ description }}
              </DialogDescription>
            </div>

            <button
              v-if="showCloseButton"
              type="button"
              class="admin-drawer-close"
              :disabled="busy"
              aria-label="关闭"
              @click="requestClose"
            />
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
