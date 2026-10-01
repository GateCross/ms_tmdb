<script setup lang="ts">
withDefaults(
  defineProps<{
    loading?: boolean;
    error?: string;
    empty?: boolean;
    loadingText?: string;
    emptyText?: string;
  }>(),
  {
    loading: false,
    error: "",
    empty: false,
    loadingText: "加载中...",
    emptyText: "暂无数据",
  },
);</script>

<template>
  <div class="load-state">
    <div v-if="loading" class="load-state-status" role="status" aria-live="polite">
      {{ loadingText }}
    </div>
    <div v-else-if="empty" class="load-state-status" role="status">
      {{ emptyText }}
    </div>
    <slot v-else />
  </div>
</template>

<style scoped>
.load-state-status {
  @apply px-4 py-8 text-center text-sm;
  color: var(--text-placeholder);
}
</style>
