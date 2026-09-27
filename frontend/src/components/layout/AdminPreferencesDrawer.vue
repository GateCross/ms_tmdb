<script setup lang="ts">
import { computed } from "vue";
import BaseDrawer from "@/components/common/BaseDrawer.vue";
import {
  sidebarControlStyle,
  themeSwatchStyle,
  type AdminPreferences,
  type AdminSidebarColor,
  type AdminSidebarOption,
  type AdminThemeColor,
  type AdminThemeOption,
} from "./adminLayoutConfig";

const props = defineProps<{
  currentSidebarOption: AdminSidebarOption;
  currentThemeOption: AdminThemeOption;
  preferences: AdminPreferences;
  sidebarOptions: AdminSidebarOption[];
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

function setThemeColor(value: AdminThemeColor) {
  emit("updatePreference", "themeColor", value);
}

function setSidebarColor(value: AdminSidebarColor) {
  emit("updatePreference", "sidebarColor", value);
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
      <p class="admin-preference-label">侧栏色 · {{ currentSidebarOption.label }}</p>
      <div class="admin-sidebar-control" role="group" aria-label="侧栏色">
        <button
          v-for="option in sidebarOptions"
          :key="option.value"
          type="button"
          class="admin-sidebar-control-btn"
          :class="{ 'admin-sidebar-control-btn-active': preferences.sidebarColor === option.value }"
          :style="sidebarControlStyle(option)"
          :aria-label="`侧栏色：${option.label}`"
          :title="option.label"
          @click="setSidebarColor(option.value)"
        >
          <span class="admin-sidebar-control-preview"></span>
          <span>{{ option.label }}</span>
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
