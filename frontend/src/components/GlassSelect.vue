<script setup lang="ts">
import { computed, useAttrs } from "vue";
import {
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectPortal,
  SelectContent,
  SelectViewport,
  SelectItem,
  SelectItemText,
  SelectItemIndicator,
} from "reka-ui";
import { ChevronDown, Check } from "lucide-vue-next";
import { cn } from "@/lib/utils";

defineOptions({ inheritAttrs: false });

type SelectOption = {
  label: string;
  value: string;
};

const props = withDefaults(
  defineProps<{
    modelValue: string;
    options: ReadonlyArray<SelectOption>;
    disabled?: boolean;
    placeholder?: string;
  }>(),
  {
    disabled: false,
    placeholder: "请选择...",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

const attrs = useAttrs();
// reka 的 SelectRoot 不渲染 DOM，外部 attrs（含 class）须转发到 SelectTrigger 才生效
const triggerClass = computed(() =>
  cn(
    "glass-select field-control inline-flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-[var(--field-border)] bg-[var(--field-bg)] px-3 py-1.5 text-sm text-[var(--text-main)] shadow-sm outline-none transition focus:border-[var(--field-border-focus)] focus:ring-1 focus:ring-[var(--field-border-focus)] disabled:cursor-not-allowed disabled:opacity-60",
    attrs.class as string | undefined,
  ),
);
const forwardedAttrs = computed(() => {
  const rest = { ...attrs };
  delete rest.class;
  return rest;
});

const selectedValue = computed({
  get: () => props.modelValue,
  set: (val: string) => {
    if (val !== props.modelValue) {
      emit("update:modelValue", val);
      emit("change", val);
    }
  },
});
</script>

<template>
  <SelectRoot v-model="selectedValue" :disabled="disabled">
    <SelectTrigger v-bind="forwardedAttrs" :class="triggerClass">
      <SelectValue :placeholder="placeholder" class="truncate" />
      <ChevronDown class="h-4 w-4 shrink-0 opacity-60" />
    </SelectTrigger>

    <SelectPortal>
      <SelectContent
        position="popper"
        :side-offset="4"
        class="z-[1400] max-h-64 min-w-[var(--reka-select-trigger-width)] overflow-hidden rounded-lg border border-[var(--border-muted)] bg-[var(--menu-bg)] p-1 text-[var(--text-main)] shadow-xl backdrop-blur-md animate-in fade-in-80"
      >
        <SelectViewport class="p-1">
          <SelectItem
            v-for="option in options"
            :key="option.value"
            :value="option.value"
            class="relative flex cursor-pointer select-none items-center rounded-md py-1.5 pl-8 pr-3 text-sm text-[var(--text-muted)] outline-none transition-colors data-[highlighted]:bg-[var(--menu-hover-bg)] data-[highlighted]:text-[var(--text-main)] data-[state=checked]:font-medium data-[state=checked]:text-[var(--text-main)]"
          >
            <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
              <SelectItemIndicator>
                <Check class="h-3.5 w-3.5 text-[var(--accent)]" />
              </SelectItemIndicator>
            </span>
            <SelectItemText>{{ option.label }}</SelectItemText>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>
