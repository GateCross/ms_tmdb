<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";
import { cancelPrefetch, prefetchMediaDetail, schedulePrefetch } from "@/api/prefetch";
import { tmdbImg } from "@/api/tmdb";
import type { TVCastMember } from "./types";

defineProps<{
  creditsLoading: boolean;
  creditsLoaded: boolean;
  creditsError?: string;
  castMembers: TVCastMember[];
  personLink: (personId: number) => RouteLocationRaw;
  onRefresh: () => void;
}>();
</script>

<template>
  <div class="content-auto mt-6">
    <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
      <h3 class="text-sm font-semibold">主要演员</h3>
      <button class="btn-soft-xs px-3 py-1 disabled:opacity-60" :disabled="creditsLoading" @click="onRefresh">
        {{ creditsLoading ? "加载中..." : creditsLoaded ? "刷新演员" : "加载演员" }}
      </button>
    </div>
    <p v-if="creditsLoading" class="text-xs text-muted">正在加载演员信息...</p>
    <div v-else-if="creditsError" class="logs-refresh-error" role="status" aria-live="polite">
      <span>{{ creditsError }}</span>
      <button type="button" class="btn-soft-xs" :disabled="creditsLoading" @click="onRefresh">重试</button>
    </div>
    <p v-else-if="creditsLoaded && !castMembers.length" class="text-xs text-muted">暂无演员数据</p>
    <div v-else-if="castMembers.length" class="cast-grid">
      <div v-for="c in castMembers" :key="c.id" class="cast-card">
        <RouterLink
          :to="personLink(c.id)"
          @pointerenter="schedulePrefetch('person', c.id)"
          @pointerleave="cancelPrefetch('person', c.id)"
          @focus="schedulePrefetch('person', c.id)"
          @blur="cancelPrefetch('person', c.id)"
          @touchstart.passive="prefetchMediaDetail('person', c.id)"
        >
          <img :src="tmdbImg(c.profile_path, 'w185')" :alt="c.name" class="cast-img" loading="lazy" />
        </RouterLink>
        <p class="mt-1 truncate text-xs font-medium">{{ c.name }}</p>
        <p class="truncate text-xs text-muted">{{ c.character }}</p>
      </div>
    </div>
  </div>
</template>
