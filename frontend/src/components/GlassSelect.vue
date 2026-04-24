<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

type SelectOption = {
  label: string;
  value: string;
};

const props = withDefaults(
  defineProps<{
    modelValue: string;
    options: ReadonlyArray<SelectOption>;
    disabled?: boolean;
  }>(),
  {
    disabled: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

const rootRef = ref<HTMLElement | null>(null);
const open = ref(false);

const selectedOption = computed(() => {
  return props.options.find((item) => item.value === props.modelValue) ?? props.options[0];
});

function closeMenu() {
  open.value = false;
}

function toggleMenu() {
  if (props.disabled) return;
  open.value = !open.value;
}

function selectOption(value: string) {
  if (props.disabled) return;
  if (props.modelValue !== value) {
    emit("update:modelValue", value);
    emit("change", value);
  }
  closeMenu();
}

function handleClickOutside(event: MouseEvent) {
  if (!open.value) return;
  const target = event.target;
  if (!(target instanceof Node)) return;
  if (rootRef.value?.contains(target)) return;
  closeMenu();
}

function handleWindowBlur() {
  closeMenu();
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  window.addEventListener("blur", handleWindowBlur);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
  window.removeEventListener("blur", handleWindowBlur);
});
</script>

<template>
  <div ref="rootRef" class="glass-select relative">
    <button
      type="button"
      class="field-control glass-select-trigger flex w-full items-center justify-between gap-2 text-left text-sm"
      :class="disabled ? 'cursor-not-allowed opacity-80' : 'cursor-pointer'"
      :disabled="disabled"
      @click="toggleMenu"
    >
      <span class="truncate">{{ selectedOption?.label ?? "-" }}</span>
      <svg
        viewBox="0 0 20 20"
        class="h-4 w-4 flex-none transition"
        :class="open ? 'rotate-180' : ''"
        aria-hidden="true"
      >
        <path
          d="M5 7.5L10 12.5L15 7.5"
          fill="none"
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.8"
        />
      </svg>
    </button>

    <Transition name="glass-select-fade">
      <div
        v-if="open"
        class="glass-select-menu absolute left-0 right-0 top-[calc(100%+0.35rem)] z-[1200] max-h-56 overflow-y-auto rounded-xl border p-1 shadow-[0_14px_30px_rgba(19,38,52,0.18)]"
      >
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="glass-select-option block w-full rounded-lg px-3 py-2 text-left text-sm transition"
          :class="option.value === modelValue ? 'glass-select-option-active' : ''"
          @click="selectOption(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.glass-select-trigger {
  background: var(--field-bg) !important;
  border-color: var(--field-border) !important;
  color: var(--text-main) !important;
}

.glass-select-trigger svg {
  color: var(--text-muted);
}

.glass-select-menu {
  background: var(--glass-bg-strong);
  border-color: var(--field-border);
  box-shadow: var(--glass-shadow-soft);
  backdrop-filter: blur(16px) saturate(145%);
}

.glass-select-option {
  color: var(--text-main);
}

.glass-select-option:hover {
  background: rgba(56, 189, 248, 0.12);
}

.glass-select-option-active {
  color: #f8fbff;
  background: linear-gradient(135deg, rgba(1, 180, 228, 0.42) 0%, rgba(2, 132, 199, 0.38) 100%);
}

.glass-select-fade-enter-active,
.glass-select-fade-leave-active {
  transition: all 0.14s ease;
}

.glass-select-fade-enter-from,
.glass-select-fade-leave-to {
  opacity: 0;
  transform: translateY(-3px);
}
</style>
