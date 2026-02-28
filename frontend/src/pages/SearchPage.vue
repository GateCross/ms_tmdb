<script setup lang="ts">
import { ref } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
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
</script>

<template>
  <section class="card">
    <h2 class="mb-4 text-lg font-semibold text-slate-800">全站搜索</h2>
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
    <h3 class="section-title">结果（{{ results.length }}）</h3>
    <ul class="space-y-2">
      <li v-for="item in results.slice(0, 20)" :key="item.id" class="search-item">
        <RouterLink :to="routeByItem(item)" class="flex items-center gap-3">
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
          <span class="text-xs text-black/40">⭐ {{ item.vote_average?.toFixed(1) ?? "" }}</span>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>
