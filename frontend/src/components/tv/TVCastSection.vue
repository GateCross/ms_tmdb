<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";
import { tmdbImg } from "@/api/tmdb";
import type { TVCastMember } from "./types";

defineProps<{
  creditsLoading: boolean;
  creditsLoaded: boolean;
  creditsError: string;
  castMembers: TVCastMember[];
  personLink: (personId: number) => RouteLocationRaw;
  onRefresh: () => void;
  onPrefetchPerson: (personId: number) => void;
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
    <p v-if="creditsLoading" class="text-xs text-black/55">正在加载演员信息...</p>
    <p v-else-if="creditsError" class="text-xs text-red-600">{{ creditsError }}</p>
    <p v-else-if="creditsLoaded && !castMembers.length" class="text-xs text-black/55">暂无演员数据</p>
    <div v-else-if="castMembers.length" class="cast-grid">
      <div v-for="c in castMembers" :key="c.id" class="cast-card">
        <RouterLink
          :to="personLink(c.id)"
          @mouseenter="onPrefetchPerson(c.id)"
          @focus="onPrefetchPerson(c.id)"
          @touchstart.passive="onPrefetchPerson(c.id)"
        >
          <img :src="tmdbImg(c.profile_path, 'w185')" :alt="c.name" class="cast-img" loading="lazy" />
        </RouterLink>
        <p class="mt-1 truncate text-xs font-medium">{{ c.name }}</p>
        <p class="truncate text-xs text-black/50">{{ c.character }}</p>
      </div>
    </div>
  </div>
</template>
