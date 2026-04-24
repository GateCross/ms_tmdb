<script setup lang="ts">
import { ref } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import { prefetchMediaDetail } from "@/api/prefetch";
import { searchByType, type SearchType } from "@/api/search";
import { tmdbImg, profileImg } from "@/api/tmdb";
import type { ApiErrorLike, SearchResultItem } from "@/types/media";

const query = ref("");
const type = ref<SearchType>("multi");
const loading = ref(false);
const error = ref("");
const results = ref<SearchResultItem[]>([]);

function resolveErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "message" in err) {
    const message = (err as ApiErrorLike).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

const typeOptions = [
  { label: "综合", value: "multi" },
  { label: "电影", value: "movie" },
  { label: "剧集", value: "tv" },
  { label: "人物", value: "person" },
] as const;

async function handleSearch() {
  if (!query.value.trim()) {
    error.value = "请输入关键词";
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const resp = await searchByType(type.value, query.value.trim(), 1);
    results.value = resp.data?.results ?? [];
  } catch (err: unknown) {
    error.value = resolveErrorMessage(err, "搜索失败");
  } finally {
    loading.value = false;
  }
}

function routeByItem(item: SearchResultItem) {
  const mt = item.media_type ?? type.value;
  if (mt === "movie") return `/movie/${item.id}`;
  if (mt === "tv") return `/tv/${item.id}`;
  if (mt === "person") return `/person/${item.id}`;
  return "";
}

function thumbByItem(item: SearchResultItem) {
  const mt = item.media_type ?? type.value;
  if (mt === "person") return profileImg(item.profile_path, "w92");
  return tmdbImg(item.poster_path, "w92");
}

function titleByItem(item: SearchResultItem) {
  return item.title || item.name || item.original_title || `ID ${item.id}`;
}

function subtitleByItem(item: SearchResultItem) {
  const mt = item.media_type ?? type.value;
  const labels: Record<string, string> = { movie: "电影", tv: "剧集", person: "人物" };
  const tag = labels[mt] ?? mt;
  const date = item.release_date || item.first_air_date || "";
  return date ? `${tag} · ${date}` : tag;
}

function prefetchSearchItem(item: SearchResultItem) {
  const mt = item.media_type ?? type.value;
  if (mt === "movie" || mt === "tv" || mt === "person") {
    prefetchMediaDetail(mt, Number(item.id));
  }
}
</script>

<template>
  <section class="card">
    <div class="mb-4">
      <p class="section-label">Search</p>
      <h2 class="search-page-title">全站搜索</h2>
      <p class="mt-1 text-sm text-black/55">跨电影、剧集和人物搜索 TMDB 数据，并快速进入详情页。</p>
    </div>
    <div class="grid gap-3 md:grid-cols-[140px_1fr_auto]">
      <GlassSelect v-model="type" :options="typeOptions" />
      <input
        v-model="query"
        type="text"
        class="field-control"
        placeholder="输入关键词，例如：Fight Club"
        @keyup.enter="handleSearch"
      />
      <button
        class="btn-primary"
        @click="handleSearch"
      >
        {{ loading ? "搜索中..." : "搜索" }}
      </button>
    </div>
    <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
  </section>

  <section v-if="results.length" class="card mt-4">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <h3 class="section-title !mb-0">结果</h3>
      <span class="badge">{{ results.length }} 条匹配</span>
    </div>
    <ul class="grid gap-2 md:grid-cols-2">
      <li v-for="item in results.slice(0, 20)" :key="item.id" class="search-item">
        <RouterLink
          :to="routeByItem(item)"
          class="flex h-full items-center gap-3"
          @mouseenter="prefetchSearchItem(item)"
          @focus="prefetchSearchItem(item)"
          @touchstart.passive="prefetchSearchItem(item)"
        >
          <img
            :src="thumbByItem(item)"
            :alt="titleByItem(item)"
            class="search-thumb"
            loading="lazy"
          />
          <div class="min-w-0 flex-1">
            <p class="truncate font-medium text-slate-800">{{ titleByItem(item) }}</p>
            <p class="text-xs text-black/55">{{ subtitleByItem(item) }}</p>
            <p v-if="item.overview" class="mt-0.5 text-xs text-black/50 line-clamp-1">
              {{ item.overview }}
            </p>
          </div>
          <span v-if="typeof item.vote_average === 'number'" class="search-score-badge">⭐ {{ item.vote_average.toFixed(1) }}</span>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>
