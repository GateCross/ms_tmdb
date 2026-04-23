<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import GlassSelect from "@/components/GlassSelect.vue";
import { prefetchMediaDetail } from "@/api/prefetch";
import {
  createMovie,
  createTV,
  deleteMovie,
  deleteTV,
  listMovies,
  listTV,
  uploadAdminImage,
  type AdminListResp,
  type AdminMovieListItem,
  type AdminTVListItem,
  type AdminCreateMoviePayload,
  type AdminCreateTVPayload,
} from "@/api/admin";
import { tmdbImg } from "@/api/tmdb";
import { getMovieGenreList } from "@/api/movie";
import { getTVGenreList } from "@/api/tv";
import { movieStatusOptions, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";
import type { ApiErrorLike } from "@/types/media";

type MediaTab = "movie" | "tv";
type ViewMode = "grid" | "table";
type SearchMode = "contains" | "prefix";
type UploadingKey = "" | "movie_poster_path" | "movie_backdrop_path" | "tv_poster_path" | "tv_backdrop_path";
type GenreOption = {
  id: number;
  name: string;
};

type LibraryListItem = AdminMovieListItem | AdminTVListItem;

type LocalMovieCreateForm = {
  title: string;
  original_title: string;
  genre_names: string[];
  release_date: string;
  status: string;
  runtime: string;
  original_language: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

type LocalTVCreateForm = {
  name: string;
  original_name: string;
  genre_names: string[];
  first_air_date: string;
  status: string;
  type: string;
  number_of_seasons: string;
  number_of_episodes: string;
  original_language: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

const route = useRoute();
const router = useRouter();
const activeTab = ref<MediaTab>(route.query.tab === "tv" ? "tv" : "movie");
const viewMode = ref<ViewMode>("grid");
const searchMode = ref<SearchMode>("contains");
const keywordInput = ref("");
const keyword = ref("");
const loading = ref(false);
const error = ref("");
const items = ref<LibraryListItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;

const createPanelVisible = ref(false);
const creating = ref(false);
const createError = ref("");
const uploadingKey = ref<UploadingKey>("");
const deletingId = ref<number | null>(null);
const deleteModalVisible = ref(false);
const pendingDeleteItem = ref<LibraryListItem | null>(null);
const movieCreateForm = ref<LocalMovieCreateForm>(emptyMovieForm());
const tvCreateForm = ref<LocalTVCreateForm>(emptyTVForm());
const movieGenreOptions = ref<GenreOption[]>([]);
const tvGenreOptions = ref<GenreOption[]>([]);

const languageOptions = [
  { label: "中文 (zh-CN)", value: "zh-CN" },
  { label: "英语 (en-US)", value: "en-US" },
  { label: "日语 (ja-JP)", value: "ja-JP" },
  { label: "韩语 (ko-KR)", value: "ko-KR" },
] as const;

const searchModeOptions = [
  { label: "模糊包含", value: "contains" },
  { label: "前缀匹配", value: "prefix" },
] as const;

const createTitle = computed(() => (activeTab.value === "movie" ? "新建本地电影" : "新建本地剧集"));
let previousBodyOverflow = "";

function resolveErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "message" in err) {
    const message = (err as ApiErrorLike).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

function normalizeListItem(item: unknown): LibraryListItem | null {
  if (!item || typeof item !== "object") return null;
  const raw = item as Record<string, unknown>;
  const tmdbId = Number(raw.tmdb_id);
  if (!Number.isFinite(tmdbId)) return null;

  const genres = Array.isArray(raw.genre_names)
    ? raw.genre_names.map((v) => String(v ?? "").trim()).filter(Boolean)
    : [];

  if (typeof raw.name === "string") {
    return {
      tmdb_id: tmdbId,
      name: String(raw.name ?? ""),
      original_name: String(raw.original_name ?? ""),
      poster_path: String(raw.poster_path ?? ""),
      vote_average: Number(raw.vote_average ?? 0),
      first_air_date: String(raw.first_air_date ?? ""),
      number_of_seasons: Number(raw.number_of_seasons ?? 0),
      number_of_episodes: Number(raw.number_of_episodes ?? 0),
      popularity: Number(raw.popularity ?? 0),
      is_modified: Boolean(raw.is_modified),
      genre_names: genres,
      id: Number.isFinite(Number(raw.id)) ? Number(raw.id) : tmdbId,
      media_type: "tv",
    };
  }

  return {
    tmdb_id: tmdbId,
    title: String(raw.title ?? ""),
    original_title: String(raw.original_title ?? ""),
    poster_path: String(raw.poster_path ?? ""),
    vote_average: Number(raw.vote_average ?? 0),
    release_date: String(raw.release_date ?? ""),
    popularity: Number(raw.popularity ?? 0),
    is_modified: Boolean(raw.is_modified),
    genre_names: genres,
    id: Number.isFinite(Number(raw.id)) ? Number(raw.id) : tmdbId,
    media_type: "movie",
  };
}

function normalizeListResults(raw: unknown): LibraryListItem[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item) => normalizeListItem(item))
    .filter((item): item is LibraryListItem => item !== null);
}

function normalizeListResponse(raw: unknown): AdminListResp<LibraryListItem> {
  const payload = raw && typeof raw === "object" ? (raw as Record<string, unknown>) : {};
  return {
    total: Number(payload.total ?? 0),
    page: Number(payload.page ?? 1),
    page_size: Number(payload.page_size ?? pageSize),
    results: normalizeListResults(payload.results),
  };
}

function emptyMovieForm(): LocalMovieCreateForm {
  return {
    title: "",
    original_title: "",
    genre_names: [],
    release_date: "",
    status: "Released",
    runtime: "",
    original_language: "zh-CN",
    poster_path: "",
    backdrop_path: "",
    vote_average: "",
    popularity: "",
    overview: "",
  };
}

function emptyTVForm(): LocalTVCreateForm {
  return {
    name: "",
    original_name: "",
    genre_names: [],
    first_air_date: "",
    status: "Returning Series",
    type: "Scripted",
    number_of_seasons: "",
    number_of_episodes: "",
    original_language: "zh-CN",
    poster_path: "",
    backdrop_path: "",
    vote_average: "",
    popularity: "",
    overview: "",
  };
}

async function loadData() {
  loading.value = true;
  error.value = "";
  try {
    const resp = activeTab.value === "movie"
      ? await listMovies(page.value, pageSize, keyword.value, searchMode.value)
      : await listTV(page.value, pageSize, keyword.value, searchMode.value);
    const normalized = normalizeListResponse(resp.data);
    items.value = normalized.results;
    total.value = normalized.total;
  } catch (err: unknown) {
    error.value = resolveErrorMessage(err, "加载失败");
  } finally {
    loading.value = false;
  }
}

function switchTab(tab: MediaTab) {
  activeTab.value = tab;
  page.value = 1;
}

function applySearch() {
  keyword.value = keywordInput.value.trim();
  page.value = 1;
}

function resetSearch() {
  keywordInput.value = "";
  keyword.value = "";
  page.value = 1;
}

function resetCreateForm() {
  movieCreateForm.value = emptyMovieForm();
  tvCreateForm.value = emptyTVForm();
  createError.value = "";
  uploadingKey.value = "";
}

function normalizeGenreOptions(raw: unknown): GenreOption[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item: unknown, idx: number) => {
      const value = item && typeof item === "object" ? (item as Record<string, unknown>) : {};
      return {
        id: Number(value.id) || idx + 1,
        name: String(value.name ?? "").trim(),
      };
    })
    .filter((item: GenreOption) => !!item.name);
}

async function loadMovieGenreOptions() {
  try {
    const resp = await getMovieGenreList();
    movieGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
  } catch {
    movieGenreOptions.value = [];
  }
}

async function loadTVGenreOptions() {
  try {
    const resp = await getTVGenreList();
    tvGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
  } catch {
    tvGenreOptions.value = [];
  }
}

function openCreatePanel() {
  createPanelVisible.value = true;
  resetCreateForm();
  if (activeTab.value === "movie") {
    void loadMovieGenreOptions();
  } else {
    void loadTVGenreOptions();
  }
}

function closeCreatePanel() {
  createPanelVisible.value = false;
  resetCreateForm();
}

function handleModalKeydown(event: KeyboardEvent) {
  if (event.key !== "Escape") return;
  if (deleteModalVisible.value) {
    closeDeleteModal();
    return;
  }
  if (createPanelVisible.value) {
    closeCreatePanel();
  }
}

function totalPages() {
  return Math.ceil(total.value / pageSize) || 1;
}

function gotoPage(p: number) {
  if (p < 1 || p > totalPages()) return;
  page.value = p;
}

function routeByItem(item: LibraryListItem) {
  if (activeTab.value === "movie") return `/movie/${item.tmdb_id}`;
  return `/tv/${item.tmdb_id}`;
}

function prefetchItemDetail(item: LibraryListItem) {
  prefetchMediaDetail(activeTab.value, Number(item.tmdb_id));
}

function openItemDetail(item: LibraryListItem) {
  void router.push(routeByItem(item));
}

function canDeleteItem(item: LibraryListItem): boolean {
  const id = Number(item?.tmdb_id);
  return Number.isInteger(id) && id !== 0;
}

function parseOptionalInt(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text);
  if (!Number.isFinite(value)) return undefined;
  return Math.trunc(value);
}

function parseOptionalFloat(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text);
  if (!Number.isFinite(value)) return undefined;
  return value;
}

async function uploadCreateImage(mediaType: MediaTab, field: "poster_path" | "backdrop_path", event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  const key = `${mediaType}_${field}` as UploadingKey;
  uploadingKey.value = key;
  createError.value = "";
  try {
    const resp = await uploadAdminImage(file);
    const path = String(resp.data?.path ?? "").trim();
    if (!path) {
      throw new Error("上传成功但未返回图片路径");
    }
    if (mediaType === "movie") {
      movieCreateForm.value[field] = path;
    } else {
      tvCreateForm.value[field] = path;
    }
  } catch (err: unknown) {
    createError.value = resolveErrorMessage(err, "图片上传失败");
  } finally {
    uploadingKey.value = "";
    input.value = "";
  }
}

async function submitCreate() {
  createError.value = "";

  if (activeTab.value === "movie") {
    const title = movieCreateForm.value.title.trim();
    if (!title) {
      createError.value = "电影标题不能为空";
      return;
    }

    const runtime = parseOptionalInt(movieCreateForm.value.runtime);
    if (movieCreateForm.value.runtime.trim() && runtime === undefined) {
      createError.value = "时长必须是数字";
      return;
    }
    const voteAverage = parseOptionalFloat(movieCreateForm.value.vote_average);
    if (movieCreateForm.value.vote_average.trim() && voteAverage === undefined) {
      createError.value = "评分必须是数字";
      return;
    }
    const popularity = parseOptionalFloat(movieCreateForm.value.popularity);
    if (movieCreateForm.value.popularity.trim() && popularity === undefined) {
      createError.value = "热度必须是数字";
      return;
    }

    const payload: AdminCreateMoviePayload = {
      title,
      original_title: movieCreateForm.value.original_title.trim(),
      release_date: movieCreateForm.value.release_date.trim(),
      status: movieCreateForm.value.status.trim(),
      original_language: movieCreateForm.value.original_language.trim(),
      poster_path: movieCreateForm.value.poster_path.trim(),
      backdrop_path: movieCreateForm.value.backdrop_path.trim(),
      overview: movieCreateForm.value.overview.trim(),
      genre_names: movieCreateForm.value.genre_names,
    };
    if (runtime !== undefined) payload.runtime = runtime;
    if (voteAverage !== undefined) payload.vote_average = voteAverage;
    if (popularity !== undefined) payload.popularity = popularity;

    creating.value = true;
    try {
      const resp = await createMovie(payload);
      const createdID = Number(resp.data?.tmdb_id);
      if (!Number.isInteger(createdID)) {
        throw new Error("创建成功但未返回有效 ID");
      }
      closeCreatePanel();
      await loadData();
      await router.push(`/movie/${createdID}`);
    } catch (err: unknown) {
      createError.value = resolveErrorMessage(err, "创建失败");
    } finally {
      creating.value = false;
    }
    return;
  }

  const name = tvCreateForm.value.name.trim();
  if (!name) {
    createError.value = "剧集名称不能为空";
    return;
  }

  const seasons = parseOptionalInt(tvCreateForm.value.number_of_seasons);
  if (tvCreateForm.value.number_of_seasons.trim() && seasons === undefined) {
    createError.value = "季数必须是数字";
    return;
  }
  const episodes = parseOptionalInt(tvCreateForm.value.number_of_episodes);
  if (tvCreateForm.value.number_of_episodes.trim() && episodes === undefined) {
    createError.value = "集数必须是数字";
    return;
  }
  const voteAverage = parseOptionalFloat(tvCreateForm.value.vote_average);
  if (tvCreateForm.value.vote_average.trim() && voteAverage === undefined) {
    createError.value = "评分必须是数字";
    return;
  }
  const popularity = parseOptionalFloat(tvCreateForm.value.popularity);
  if (tvCreateForm.value.popularity.trim() && popularity === undefined) {
    createError.value = "热度必须是数字";
    return;
  }

  const payload: AdminCreateTVPayload = {
    name,
    original_name: tvCreateForm.value.original_name.trim(),
    first_air_date: tvCreateForm.value.first_air_date.trim(),
    status: tvCreateForm.value.status.trim(),
    type: tvCreateForm.value.type.trim(),
    original_language: tvCreateForm.value.original_language.trim(),
    poster_path: tvCreateForm.value.poster_path.trim(),
    backdrop_path: tvCreateForm.value.backdrop_path.trim(),
    overview: tvCreateForm.value.overview.trim(),
    genre_names: tvCreateForm.value.genre_names,
  };
  if (seasons !== undefined) payload.number_of_seasons = seasons;
  if (episodes !== undefined) payload.number_of_episodes = episodes;
  if (voteAverage !== undefined) payload.vote_average = voteAverage;
  if (popularity !== undefined) payload.popularity = popularity;

  creating.value = true;
  try {
    const resp = await createTV(payload);
    const createdID = Number(resp.data?.tmdb_id);
    if (!Number.isInteger(createdID)) {
      throw new Error("创建成功但未返回有效 ID");
    }
    closeCreatePanel();
    await loadData();
    await router.push(`/tv/${createdID}`);
  } catch (err: unknown) {
    createError.value = resolveErrorMessage(err, "创建失败");
  } finally {
    creating.value = false;
  }
}

function requestDeleteItem(item: LibraryListItem) {
  const id = Number(item?.tmdb_id);
  if (!Number.isInteger(id) || id === 0) {
    return;
  }
  pendingDeleteItem.value = item;
  deleteModalVisible.value = true;
}

function closeDeleteModal() {
  deleteModalVisible.value = false;
  pendingDeleteItem.value = null;
}

async function confirmDeleteItem() {
  const item = pendingDeleteItem.value;
  if (!item) return;
  const id = Number(item.tmdb_id);
  if (!Number.isInteger(id) || id === 0) return;

  deletingId.value = id;
  error.value = "";
  try {
    if (activeTab.value === "movie") {
      await deleteMovie(id);
    } else {
      await deleteTV(id);
    }

    if (items.value.length <= 1 && page.value > 1) {
      page.value = page.value - 1;
      closeDeleteModal();
      return;
    }
    await loadData();
    closeDeleteModal();
  } catch (err: unknown) {
    error.value = resolveErrorMessage(err, "删除失败");
  } finally {
    deletingId.value = null;
  }
}

watch(
  () => route.query.tab,
  (tab) => {
    const nextTab: MediaTab = tab === "tv" ? "tv" : "movie";
    if (nextTab !== activeTab.value) {
      activeTab.value = nextTab;
      page.value = 1;
      createPanelVisible.value = false;
      createError.value = "";
      if (nextTab === "movie") {
        void loadMovieGenreOptions();
      } else {
        void loadTVGenreOptions();
      }
    }
  },
);

watch(() => createPanelVisible.value || deleteModalVisible.value, (visible) => {
  if (visible) {
    previousBodyOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleModalKeydown);
    return;
  }

  document.body.style.overflow = previousBodyOverflow;
  window.removeEventListener("keydown", handleModalKeydown);
});

onBeforeUnmount(() => {
  document.body.style.overflow = previousBodyOverflow;
  window.removeEventListener("keydown", handleModalKeydown);
});

watch([activeTab, page, keyword, searchMode], loadData);
onMounted(loadData);
</script>

<template>
  <section class="flex flex-wrap items-center gap-2">
    <div class="glass-pill gap-2">
      <button
        v-for="tab in ([
          { key: 'movie', label: '🎬 电影' },
          { key: 'tv', label: '📺 剧集' },
        ] as const)"
        :key="tab.key"
        class="glass-pill-btn px-5"
        :class="activeTab === tab.key ? 'glass-pill-btn-active' : ''"
        @click="switchTab(tab.key as MediaTab)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="glass-pill">
      <button
        class="glass-pill-btn px-4 py-1.5 text-xs"
        :class="viewMode === 'grid' ? 'glass-pill-btn-active' : ''"
        @click="viewMode = 'grid'"
      >
        卡片
      </button>
      <button
        class="glass-pill-btn px-4 py-1.5 text-xs"
        :class="viewMode === 'table' ? 'glass-pill-btn-active' : ''"
        @click="viewMode = 'table'"
      >
        表格
      </button>
    </div>

    <button class="btn-primary" @click="openCreatePanel">
      {{ createTitle }}
    </button>
  </section>

  <section class="card mt-4">
    <div class="grid gap-3 md:grid-cols-[1fr_auto_auto_auto] md:items-center">
      <input
        v-model="keywordInput"
        class="w-full field-control text-sm"
        placeholder="输入 TMDB ID 或片名/剧名关键词"
        @keyup.enter="applySearch"
      />
      <GlassSelect v-model="searchMode" :options="searchModeOptions" />
      <button class="btn-primary" @click="applySearch">
        搜索
      </button>
      <button class="btn-soft" @click="resetSearch">
        重置
      </button>
    </div>
  </section>

  <div v-if="createPanelVisible" class="fixed inset-0 z-[1000] flex items-center justify-center p-3 sm:p-6">
    <div class="absolute inset-0 bg-black/60 backdrop-blur-[2px]" @click="closeCreatePanel" />
    <section class="panel-glass relative z-10 w-full max-w-5xl overflow-hidden rounded-2xl">
      <div class="sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/60 bg-white/70 px-4 py-3 backdrop-blur sm:px-6">
        <h3 class="text-sm font-semibold">{{ createTitle }}</h3>
        <button class="btn-soft px-3 py-1.5 text-xs" @click="closeCreatePanel">
          关闭
        </button>
      </div>

      <div class="max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6">
    <div v-if="activeTab === 'movie'" class="grid gap-3 md:grid-cols-2">
      <label class="text-xs text-black/60">
        标题（必填）
        <input v-model="movieCreateForm.title" class="field-control mt-1 w-full text-sm" placeholder="电影标题" />
      </label>
      <label class="text-xs text-black/60">
        原始标题
        <input v-model="movieCreateForm.original_title" class="field-control mt-1 w-full text-sm" placeholder="Original Title" />
      </label>
      <label class="text-xs text-black/60">
        上映日期
        <input v-model="movieCreateForm.release_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
      </label>
      <label class="text-xs text-black/60">
        状态
        <GlassSelect v-model="movieCreateForm.status" :options="movieStatusOptions" class="mt-1 w-full" />
      </label>
      <label class="text-xs text-black/60">
        原始语言
        <GlassSelect v-model="movieCreateForm.original_language" :options="languageOptions" class="mt-1 w-full" />
      </label>
      <label class="text-xs text-black/60">
        时长（分钟）
        <input v-model="movieCreateForm.runtime" class="field-control mt-1 w-full text-sm" placeholder="120" />
      </label>
      <label class="text-xs text-black/60 md:col-span-2">
        类型（多选）
        <div class="mt-1 max-h-32 overflow-y-auto rounded-lg border border-white/70 bg-white/55 p-2 backdrop-blur">
          <label v-for="genre in movieGenreOptions" :key="genre.id" class="mr-3 inline-flex items-center gap-1.5 py-1 text-xs">
            <input v-model="movieCreateForm.genre_names" type="checkbox" class="check-control" :value="genre.name" />
            <span>{{ genre.name }}</span>
          </label>
          <span v-if="!movieGenreOptions.length" class="text-xs text-black/50">暂无可选类型</span>
        </div>
      </label>
      <label class="text-xs text-black/60">
        海报路径
        <input v-model="movieCreateForm.poster_path" readonly class="field-control mt-1 w-full text-sm opacity-80" placeholder="上传后自动填充" />
        <input class="mt-2 block w-full text-xs" type="file" accept="image/*" @change="(e) => uploadCreateImage('movie', 'poster_path', e)" />
        <span v-if="uploadingKey === 'movie_poster_path'" class="mt-1 inline-block text-[11px] text-black/50">上传中...</span>
      </label>
      <label class="text-xs text-black/60">
        背景图路径
        <input v-model="movieCreateForm.backdrop_path" readonly class="field-control mt-1 w-full text-sm opacity-80" placeholder="上传后自动填充" />
        <input class="mt-2 block w-full text-xs" type="file" accept="image/*" @change="(e) => uploadCreateImage('movie', 'backdrop_path', e)" />
        <span v-if="uploadingKey === 'movie_backdrop_path'" class="mt-1 inline-block text-[11px] text-black/50">上传中...</span>
      </label>
      <label class="text-xs text-black/60">
        评分
        <input v-model="movieCreateForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="7.8" />
      </label>
      <label class="text-xs text-black/60">
        热度
        <input v-model="movieCreateForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="123.4" />
      </label>
      <label class="text-xs text-black/60 md:col-span-2">
        简介
        <textarea v-model="movieCreateForm.overview" rows="3" class="field-control mt-1 w-full text-sm" placeholder="简介" />
      </label>
    </div>

    <div v-else class="grid gap-3 md:grid-cols-2">
      <label class="text-xs text-black/60">
        剧名（必填）
        <input v-model="tvCreateForm.name" class="field-control mt-1 w-full text-sm" placeholder="剧集名称" />
      </label>
      <label class="text-xs text-black/60">
        原始剧名
        <input v-model="tvCreateForm.original_name" class="field-control mt-1 w-full text-sm" placeholder="Original Name" />
      </label>
      <label class="text-xs text-black/60">
        首播日期
        <input v-model="tvCreateForm.first_air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
      </label>
      <label class="text-xs text-black/60">
        状态
        <GlassSelect v-model="tvCreateForm.status" :options="tvStatusOptions" class="mt-1 w-full" />
      </label>
      <label class="text-xs text-black/60">
        原始语言
        <GlassSelect v-model="tvCreateForm.original_language" :options="languageOptions" class="mt-1 w-full" />
      </label>
      <label class="text-xs text-black/60">
        剧集类型
        <GlassSelect v-model="tvCreateForm.type" :options="tvTypeOptions" class="mt-1 w-full" />
      </label>
      <label class="text-xs text-black/60">
        季数
        <input v-model="tvCreateForm.number_of_seasons" class="field-control mt-1 w-full text-sm" placeholder="3" />
      </label>
      <label class="text-xs text-black/60">
        集数
        <input v-model="tvCreateForm.number_of_episodes" class="field-control mt-1 w-full text-sm" placeholder="24" />
      </label>
      <label class="text-xs text-black/60 md:col-span-2">
        类型（多选）
        <div class="mt-1 max-h-32 overflow-y-auto rounded-lg border border-white/70 bg-white/55 p-2 backdrop-blur">
          <label v-for="genre in tvGenreOptions" :key="genre.id" class="mr-3 inline-flex items-center gap-1.5 py-1 text-xs">
            <input v-model="tvCreateForm.genre_names" type="checkbox" class="check-control" :value="genre.name" />
            <span>{{ genre.name }}</span>
          </label>
          <span v-if="!tvGenreOptions.length" class="text-xs text-black/50">暂无可选类型</span>
        </div>
      </label>
      <label class="text-xs text-black/60">
        海报路径
        <input v-model="tvCreateForm.poster_path" readonly class="field-control mt-1 w-full text-sm opacity-80" placeholder="上传后自动填充" />
        <input class="mt-2 block w-full text-xs" type="file" accept="image/*" @change="(e) => uploadCreateImage('tv', 'poster_path', e)" />
        <span v-if="uploadingKey === 'tv_poster_path'" class="mt-1 inline-block text-[11px] text-black/50">上传中...</span>
      </label>
      <label class="text-xs text-black/60">
        背景图路径
        <input v-model="tvCreateForm.backdrop_path" readonly class="field-control mt-1 w-full text-sm opacity-80" placeholder="上传后自动填充" />
        <input class="mt-2 block w-full text-xs" type="file" accept="image/*" @change="(e) => uploadCreateImage('tv', 'backdrop_path', e)" />
        <span v-if="uploadingKey === 'tv_backdrop_path'" class="mt-1 inline-block text-[11px] text-black/50">上传中...</span>
      </label>
      <label class="text-xs text-black/60">
        评分
        <input v-model="tvCreateForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="8.1" />
      </label>
      <label class="text-xs text-black/60">
        热度
        <input v-model="tvCreateForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="220.5" />
      </label>
      <label class="text-xs text-black/60 md:col-span-2">
        简介
        <textarea v-model="tvCreateForm.overview" rows="3" class="field-control mt-1 w-full text-sm" placeholder="简介" />
      </label>
    </div>

      <div class="mt-4 flex items-center gap-3">
        <button class="btn-primary disabled:opacity-60" :disabled="creating || uploadingKey !== ''" @click="submitCreate">
          {{ creating ? "创建中..." : "创建并进入详情" }}
        </button>
        <span v-if="createError" class="text-xs text-red-600">{{ createError }}</span>
      </div>
      </div>
    </section>
  </div>

  <div v-if="deleteModalVisible" class="fixed inset-0 z-[1100] flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/65 backdrop-blur-[2px]" @click="closeDeleteModal" />
    <section class="panel-glass relative z-10 w-full max-w-md rounded-2xl p-5">
      <h3 class="text-base font-semibold text-ink">确认删除</h3>
      <p class="mt-2 text-sm text-black/70">
        将删除本地数据：
        <span class="font-medium text-black">{{ pendingDeleteItem?.title || pendingDeleteItem?.name || `ID ${pendingDeleteItem?.tmdb_id ?? ""}` }}</span>
      </p>
      <p class="mt-1 text-xs text-black/55">删除后不可恢复。</p>
      <div class="mt-5 flex justify-end gap-2">
        <button
          class="btn-soft"
          :disabled="deletingId !== null"
          @click="closeDeleteModal"
        >
          取消
        </button>
        <button
          class="btn-danger-soft disabled:opacity-60"
          :disabled="deletingId !== null"
          @click="confirmDeleteItem"
        >
          {{ deletingId !== null ? "删除中..." : "确认删除" }}
        </button>
      </div>
    </section>
  </div>

  <p v-if="loading" class="card mt-4 text-sm text-black/60">加载中...</p>
  <p v-else-if="error" class="card mt-4 text-sm text-red-600">{{ error }}</p>

  <template v-else>
    <section class="mt-4 flex items-center justify-between">
      <p class="text-sm text-black/60">
        共 <strong>{{ total }}</strong> 条记录 · 第 {{ page }}/{{ totalPages() }} 页
      </p>
    </section>

    <section v-if="viewMode === 'grid'" class="mt-4 poster-grid">
      <div
        v-for="item in items"
        :key="item.tmdb_id"
        class="poster-card group relative"
      >
        <button
          v-if="canDeleteItem(item)"
          type="button"
          class="absolute right-2 top-2 z-20 flex h-8 w-8 items-center justify-center rounded-full border border-red-300 bg-white/95 text-red-600 shadow-sm opacity-0 transition-all duration-200 pointer-events-none group-hover:opacity-100 group-hover:pointer-events-auto group-focus-within:opacity-100 group-focus-within:pointer-events-auto hover:bg-red-50 disabled:cursor-not-allowed"
          :class="deletingId === item.tmdb_id ? 'opacity-100 pointer-events-none' : ''"
          :disabled="deletingId === item.tmdb_id"
          :title="deletingId === item.tmdb_id ? '删除中' : '删除本地数据'"
          @click.stop="requestDeleteItem(item)"
        >
          <span v-if="deletingId === item.tmdb_id" class="text-[11px]">...</span>
          <svg v-else viewBox="0 0 24 24" class="h-4 w-4 fill-none stroke-current" stroke-width="1.8" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
          </svg>
        </button>
        <RouterLink
          :to="routeByItem(item)"
          class="block"
          @mouseenter="prefetchItemDetail(item)"
          @focus="prefetchItemDetail(item)"
          @touchstart.passive="prefetchItemDetail(item)"
        >
          <img
            :src="tmdbImg(item.poster_path, 'w185')"
            :alt="item.title || item.name"
            class="poster-img"
            loading="lazy"
          />
          <div class="poster-info">
            <p class="truncate text-sm font-medium">{{ item.title || item.name }}</p>
            <p class="text-xs text-black/55">
              ⭐ {{ (item.vote_average ?? 0).toFixed(1) }}
              <span class="ml-1">{{ (item.release_date || item.first_air_date || "").slice(0, 4) }}</span>
            </p>
            <span v-if="item.tmdb_id < 0" class="chip-local-new mt-1 text-[10px]">
              本地新建
            </span>
            <span v-else-if="item.is_modified" class="chip-modified mt-1 text-[10px]">
              已修改
            </span>
          </div>
        </RouterLink>
      </div>
    </section>

    <section v-else class="table-shell">
      <table class="min-w-full text-left text-sm">
        <thead class="table-head text-xs uppercase tracking-wide text-black/60">
          <tr>
            <th class="px-4 py-3">TMDB ID</th>
            <th class="px-4 py-3">名称</th>
            <th class="px-4 py-3">评分</th>
            <th class="px-4 py-3">日期</th>
            <th class="px-4 py-3">类型</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="item.tmdb_id"
            class="border-t border-black/5 hover:bg-black/[0.02]"
          >
            <td class="px-4 py-3">{{ item.tmdb_id }}</td>
            <td class="px-4 py-3">
              <p class="font-medium">{{ item.title || item.name }}</p>
              <p class="text-xs text-black/50">{{ item.original_title || item.original_name }}</p>
            </td>
            <td class="px-4 py-3">⭐ {{ (item.vote_average ?? 0).toFixed(1) }}</td>
            <td class="px-4 py-3">{{ item.release_date || item.first_air_date || "-" }}</td>
            <td class="px-4 py-3">
              <span class="text-xs text-black/70">
                {{ Array.isArray(item.genre_names) && item.genre_names.length ? item.genre_names.join(" / ") : "-" }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span
                v-if="item.tmdb_id < 0"
                class="chip-local-new"
              >
                本地新建
              </span>
              <span
                v-else-if="item.is_modified"
                class="chip-modified"
              >
                已修改
              </span>
              <span v-else class="text-xs text-black/45">未修改</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <button
                  class="btn-soft px-3 py-1 text-xs"
                  @mouseenter="prefetchItemDetail(item)"
                  @focus="prefetchItemDetail(item)"
                  @click="openItemDetail(item)"
                >
                  查看详情
                </button>
                <button
                  v-if="canDeleteItem(item)"
                  class="rounded-lg border border-red-300 bg-red-50 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-100 disabled:opacity-60"
                  :disabled="deletingId === item.tmdb_id"
                  @click="requestDeleteItem(item)"
                >
                  {{ deletingId === item.tmdb_id ? "删除中..." : "删除" }}
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="items.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-black/50">无数据</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="mt-6 flex items-center justify-center gap-2">
      <button
        class="btn-soft px-3 py-1.5 disabled:opacity-40"
        :disabled="page <= 1"
        @click="gotoPage(page - 1)"
      >
        上一页
      </button>
      <span class="px-3 text-sm text-black/60">{{ page }} / {{ totalPages() }}</span>
      <button
        class="btn-soft px-3 py-1.5 disabled:opacity-40"
        :disabled="page >= totalPages()"
        @click="gotoPage(page + 1)"
      >
        下一页
      </button>
    </section>
  </template>
</template>
