<script setup lang="ts">
import { computed } from "vue";
import BaseDrawer from "@/components/common/BaseDrawer.vue";
import {
  themeSwatchStyle,
  type AdminAppearance,
  type AdminPreferences,
  type AdminThemeColor,
  type AdminThemeOption,
} from "./adminLayoutConfig";

const appearanceOptions: Array<{ value: AdminAppearance; label: string }> = [
  { value: "light", label: "浅色" },
  { value: "dark", label: "深色" },
  { value: "auto", label: "跟随系统" },
];

const props = defineProps<{
  currentThemeOption: AdminThemeOption;
  preferences: AdminPreferences;
  themeOptions: AdminThemeOption[];
  visible: boolean;
}>();

const emit = defineEmits<{
  close: [];
  reset: [];
  updatePreference: [key: keyof AdminPreferences, value: AdminPreferences[keyof AdminPreferences]];
}>();

const compact = computed({
  get: () => props.preferences.compact,
  set: (value: boolean) => emit("updatePreference", "compact", value),
});
const showTabs = computed({
  get: () => props.preferences.showTabs,
  set: (value: boolean) => emit("updatePreference", "showTabs", value),
});
const sidebarCollapsed = computed({
  get: () => props.preferences.sidebarCollapsed,
  set: (value: boolean) => emit("updatePreference", "sidebarCollapsed", value),
});
const appearance = computed({
  get: () => props.preferences.appearance,
  set: (value: AdminAppearance) => emit("updatePreference", "appearance", value),
});

function setThemeColor(value: AdminThemeColor) {
  emit("updatePreference", "themeColor", value);
}
</script>

<template>
  <BaseDrawer :visible="visible" title="偏好设置" footer-class="mt-auto" @close="emit('close')">
    <section class="admin-preference-group" aria-label="布局">
      <p class="admin-preference-section-title">布局</p>
      <label class="admin-preference-option">
        <span>折叠菜单</span>
        <input v-model="sidebarCollapsed" type="checkbox" class="admin-switch-input" />
        <span class="admin-switch-track" aria-hidden="true">
          <span class="admin-switch-thumb"></span>
        </span>
      </label>
      <label class="admin-preference-option">
        <span>显示标签栏</span>
        <input v-model="showTabs" type="checkbox" class="admin-switch-input" />
        <span class="admin-switch-track" aria-hidden="true">
          <span class="admin-switch-thumb"></span>
        </span>
      </label>
    </section>

    <section class="admin-preference-group" aria-label="外观">
      <p class="admin-preference-section-title">外观</p>
      <p class="admin-preference-label">界面模式</p>
      <div class="admin-choice-grid admin-appearance-switch" role="group" aria-label="界面模式">
        <button
          v-for="option in appearanceOptions"
          :key="option.value"
          type="button"
          class="admin-choice-btn"
          :class="{ 'admin-choice-btn-active': preferences.appearance === option.value }"
          :aria-pressed="preferences.appearance === option.value"
          @click="appearance = option.value"
        >
          <span>{{ option.label }}</span>
        </button>
      </div>
      <p class="admin-preference-label">主题色 · {{ currentThemeOption.label }}</p>
      <div class="admin-theme-grid">
        <button
          v-for="option in themeOptions"
          :key="option.value"
          type="button"
          class="admin-theme-swatch"
          :class="{ 'admin-theme-swatch-active': preferences.themeColor === option.value }"
          :style="themeSwatchStyle(option)"
          :aria-label="`主题色：${option.label}`"
          :title="option.label"
          @click="setThemeColor(option.value)"
        >
          <span v-if="preferences.themeColor === option.value" class="admin-theme-check"></span>
        </button>
      </div>
    </section>

    <label class="admin-preference-option">
      <span>紧凑间距</span>
      <input v-model="compact" type="checkbox" class="admin-switch-input" />
      <span class="admin-switch-track" aria-hidden="true">
        <span class="admin-switch-thumb"></span>
      </span>
    </label>

    <template #footer>
      <button class="btn-soft admin-preference-reset" type="button" @click="emit('reset')">恢复默认</button>
    </template>
  </BaseDrawer>
</template>
