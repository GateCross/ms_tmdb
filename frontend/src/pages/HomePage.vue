<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { getPopularMovies } from "@/api/movie";
import { searchByType, type SearchType } from "@/api/search";
import { getPopularTV } from "@/api/tv";
import { profileImg, tmdbImg } from "@/api/tmdb";
import type { ApiErrorLike, MediaSummary, SearchResultItem } from "@/types/media";

const loading = ref(false);
const error = ref("");
const movies = ref<MediaSummary[]>([]);
const tvSeries = ref<MediaSummary[]>([]);
const searchQuery = ref("");
const searchType = ref<SearchType>("multi");
const searching = ref(false);
const searchError = ref("");
const searchResults = ref<SearchResultItem[]>([]);
const hasSearched = ref(false);
const heroStyle = computed(() => {
  const backdropPath = movies.value[0]?.backdrop_path;
  if (!backdropPath) return undefined;
  return { backgroundImage: `url(${tmdbImg(backdropPath, "w780")})` };
});

const searchTypeOptions: ReadonlyArray<{ label: string; value: SearchType }> = [
  { label: "综合", value: "multi" },
  { label: "电影", value: "movie" },
  { label: "剧集", value: "tv" },
  { label: "人物", value: "person" },
];

function resolveErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "message" in err) {
    const message = (err as ApiErrorLike).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

async function loadData() {
  loading.value = true;
  error.value = "";
  try {
    const [movieResp, tvResp] = await Promise.all([
      getPopularMovies(1),
      getPopularTV(1),
    ]);
    movies.value = movieResp.data?.results ?? [];
    tvSeries.value = tvResp.data?.results ?? [];
  } catch (err: unknown) {
    error.value = resolveErrorMessage(err, "加载失败");
  } finally {
    loading.value = false;
  }
}

async function handleHomeSearch() {
  const trimmedQuery = searchQuery.value.trim();
  if (!trimmedQuery) {
    searchError.value = "请输入关键词";
    hasSearched.value = false;
    searchResults.value = [];
    return;
  }

  searching.value = true;
  searchError.value = "";

  try {
    const resp = await searchByType(searchType.value, trimmedQuery, 1);
    searchResults.value = (resp.data?.results ?? []).slice(0, 12);
    hasSearched.value = true;
  } catch (err: unknown) {
    searchError.value = resolveErrorMessage(err, "搜索失败");
    hasSearched.value = false;
  } finally {
    searching.value = false;
  }
}

function routeByItem(item: SearchResultItem) {
  const mediaType = item.media_type ?? searchType.value;
  if (mediaType === "movie") return `/movie/${item.id}`;
  if (mediaType === "tv") return `/tv/${item.id}`;
  if (mediaType === "person") return `/person/${item.id}`;
  return "/search";
}

function thumbByItem(item: SearchResultItem) {
  const mediaType = item.media_type ?? searchType.value;
  if (mediaType === "person") return profileImg(item.profile_path, "w92");
  return tmdbImg(item.poster_path, "w92");
}

function titleByItem(item: SearchResultItem) {
  return item.title || item.name || item.original_title || `ID ${item.id}`;
}

function subtitleByItem(item: SearchResultItem) {
  const mediaType = item.media_type ?? searchType.value;
  const labels: Record<string, string> = { movie: "电影", tv: "剧集", person: "人物" };
  const tag = labels[mediaType] ?? mediaType;
  const date = item.release_date || item.first_air_date || "";
  return date ? `${tag} · ${date}` : tag;
}

onMounted(loadData);
</script>

<template>
  <section
    class="home-hero hero-banner"
    :style="heroStyle"
  >
    <div class="home-hero-overlay">
      <p class="home-hero-tag">欢迎。</p>
      <h2 class="home-hero-title">电影、剧集、人物一站式搜索</h2>
      <p class="home-hero-desc">输入关键词即可查看结果，并快速进入对应详情页。</p>

      <div class="home-search-row">
        <input
          v-model="searchQuery"
          type="text"
          class="home-search-input"
          placeholder="搜索电影、剧集、人物..."
          @keyup.enter="handleHomeSearch"
        />
        <button class="home-search-btn" :disabled="searching" @click="handleHomeSearch">
          {{ searching ? "探索中..." : "探索" }}
        </button>
      </div>

      <div class="home-type-tabs">
        <button
          v-for="option in searchTypeOptions"
          :key="option.value"
          class="home-type-btn"
          :class="{ 'home-type-btn-active': searchType === option.value }"
          @click="searchType = option.value"
        >
          {{ option.label }}
        </button>
      </div>

      <p v-if="searchError" class="mt-3 text-sm text-rose-200">{{ searchError }}</p>
    </div>
  </section>

  <section v-if="hasSearched" class="card mt-4">
    <div class="mb-3 flex items-center justify-between gap-2">
      <h3 class="section-title !mb-0">搜索结果</h3>
      <span class="text-xs text-black/55">
        {{ searchResults.length ? `展示前 ${searchResults.length} 条` : "没有匹配结果" }}
      </span>
    </div>
    <ul v-if="searchResults.length" class="space-y-2">
      <li v-for="item in searchResults" :key="`${item.media_type ?? searchType}-${item.id}`" class="search-item">
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
          <span class="text-xs text-black/40">
            {{ typeof item.vote_average === "number" ? `⭐ ${item.vote_average.toFixed(1)}` : "" }}
          </span>
        </RouterLink>
      </li>
    </ul>
    <p v-else class="text-sm text-black/55">未找到结果，请尝试更换关键词。</p>
  </section>

  <section class="mt-4 flex items-center justify-between">
    <p class="section-label">今日看点</p>
    <button
      class="btn-primary"
      :disabled="loading"
      @click="loadData"
    >
      {{ loading ? "刷新中..." : "刷新数据" }}
    </button>
  </section>
  <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>

  <!-- 热门电影 -->
  <section class="mt-6">
    <h3 class="section-title">热门电影</h3>
    <div class="poster-grid">
      <RouterLink
        v-for="item in movies.slice(0, 10)"
        :key="item.id"
        :to="`/movie/${item.id}`"
        class="poster-card"
      >
        <img
          :src="tmdbImg(item.poster_path, 'w185')"
          :alt="item.title"
          class="poster-img"
          loading="lazy"
        />
        <div class="poster-info">
          <p class="truncate text-sm font-medium">{{ item.title || item.original_title }}</p>
          <p class="text-xs text-black/55">
            ⭐ {{ item.vote_average?.toFixed(1) ?? "-" }}
            <span class="ml-1">{{ item.release_date?.slice(0, 4) ?? "" }}</span>
          </p>
        </div>
      </RouterLink>
    </div>
  </section>

  <!-- 热门剧集 -->
  <section class="mt-8">
    <h3 class="section-title">热门剧集</h3>
    <div class="poster-grid">
      <RouterLink
        v-for="item in tvSeries.slice(0, 10)"
        :key="item.id"
        :to="`/tv/${item.id}`"
        class="poster-card"
      >
        <img
          :src="tmdbImg(item.poster_path, 'w185')"
          :alt="item.name"
          class="poster-img"
          loading="lazy"
        />
        <div class="poster-info">
          <p class="truncate text-sm font-medium">{{ item.name || item.original_name }}</p>
          <p class="text-xs text-black/55">
            ⭐ {{ item.vote_average?.toFixed(1) ?? "-" }}
            <span class="ml-1">{{ item.first_air_date?.slice(0, 4) ?? "" }}</span>
          </p>
        </div>
      </RouterLink>
    </div>
  </section>
</template>
