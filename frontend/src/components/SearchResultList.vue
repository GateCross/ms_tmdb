<script setup lang="ts">
import { computed } from "vue";
import type { SearchType } from "@/api/search";
import type { SearchResultItem } from "@/types/media";
import {
  getSearchResultKey,
  getSearchResultRoute,
  getSearchResultSubtitle,
  getSearchResultThumb,
  getSearchResultTitle,
} from "@/utils/searchResult";
import { ratingTierClass } from "@/utils/ratingTier";

const props = withDefaults(
  defineProps<{
    items: SearchResultItem[];
    fallbackType: SearchType;
    limit?: number;
    emptyText?: string;
  }>(),
  {
    limit: undefined,
    emptyText: "",
  },
);

const visibleItems = computed(() => {
  if (typeof props.limit !== "number") return props.items;
  return props.items.slice(0, props.limit);
});
</script>

<template>
  <ul v-if="visibleItems.length" class="grid gap-2 md:grid-cols-2">
    <li v-for="item in visibleItems" :key="getSearchResultKey(item, fallbackType)" class="search-item">
      <RouterLink :to="getSearchResultRoute(item, fallbackType)" class="flex h-full items-center gap-3">
        <img
          :src="getSearchResultThumb(item, fallbackType)"
          :alt="getSearchResultTitle(item)"
          class="search-thumb"
          loading="lazy"
        />
        <div class="min-w-0 flex-1">
          <p class="truncate font-medium text-slate-800">{{ getSearchResultTitle(item) }}</p>
          <p class="text-xs text-black/55">{{ getSearchResultSubtitle(item, fallbackType) }}</p>
          <p v-if="item.overview" class="mt-0.5 text-xs text-black/50 line-clamp-1">
            {{ item.overview }}
          </p>
        </div>
        <span
          v-if="typeof item.vote_average === 'number'"
          class="rating-badge"
          :class="ratingTierClass(item.vote_average)"
          >{{ item.vote_average.toFixed(1) }} 分</span
        >
      </RouterLink>
    </li>
  </ul>
  <p v-else-if="emptyText" class="empty-state">{{ emptyText }}</p>
</template>
