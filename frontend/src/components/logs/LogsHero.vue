<script setup lang="ts">
type LogTab = "access" | "tmdb" | "autoSync";

defineProps<{
  activeTab: LogTab;
  accessTotalText: string;
  tmdbTotalText: string;
  autoSyncTotalText: string;
}>();

const emit = defineEmits<{
  "select-tab": [tab: LogTab];
}>();
</script>

<template>
  <section class="logs-hero card">
    <div class="logs-hero-head">
      <div class="min-w-0">
        <p class="section-label">Logs</p>
        <h2 class="library-toolbar-title">请求日志</h2>
        <p class="mt-1 text-sm text-muted">代理访问、TMDB 回源请求与定时同步执行记录。</p>
      </div>
    </div>

    <div class="logs-hero-stats" aria-label="日志分类">
      <button
        type="button"
        class="logs-stat-card"
        :class="activeTab === 'access' ? 'logs-stat-card-active' : ''"
        @click="emit('select-tab', 'access')"
      >
        <span>外部访问</span>
        <strong>{{ accessTotalText }}</strong>
        <small>代理入口请求</small>
      </button>
      <button
        type="button"
        class="logs-stat-card"
        :class="activeTab === 'tmdb' ? 'logs-stat-card-active' : ''"
        @click="emit('select-tab', 'tmdb')"
      >
        <span>TMDB 请求</span>
        <strong>{{ tmdbTotalText }}</strong>
        <small>上游回源记录</small>
      </button>
      <button
        type="button"
        class="logs-stat-card"
        :class="activeTab === 'autoSync' ? 'logs-stat-card-active' : ''"
        @click="emit('select-tab', 'autoSync')"
      >
        <span>定时任务</span>
        <strong>{{ autoSyncTotalText }}</strong>
        <small>同步执行日志</small>
      </button>
    </div>
  </section>
</template>
