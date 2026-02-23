<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { compareMovieRemote, syncMovie, updateMovie } from "@/api/admin";
import type { AdminSyncMode } from "@/api/admin";
import DetailSyncPanel from "@/components/DetailSyncPanel.vue";
import { getMovieDetail, getMovieGenreList } from "@/api/movie";
import { tmdbImg } from "@/api/tmdb";

type GenreOption = {
  id: number;
  name: string;
};

type MovieEditForm = {
  title: string;
  original_title: string;
  genre_names: string[];
  tagline: string;
  release_date: string;
  status: string;
  runtime: string;
  original_language: string;
  homepage: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

type RemoteDiffNotice = {
  summary: string;
  fields: string[];
};

type RemoteDiffDecision = "unknown" | "has_diff_pending" | "keep_local" | "overwritten" | "no_diff";

const route = useRoute();
const loading = ref(false);
const error = ref("");
const detail = ref<any>(null);
const isEditing = ref(false);
const saving = ref(false);
const saveError = ref("");
const saveMessage = ref("");
const comparedRemoteId = ref<number | null>(null);
const checkingRemoteDiff = ref(false);
const remoteDiffNotice = ref<RemoteDiffNotice | null>(null);
const resolvingRemoteDiff = ref(false);
const remoteDiffMessage = ref("");
const remoteDiffError = ref("");
const remoteDiffDecision = ref<RemoteDiffDecision>("unknown");
const genreOptions = ref<GenreOption[]>([]);
const genreKeyword = ref("");
const filteredGenreOptions = computed(() => {
  const keyword = genreKeyword.value.trim().toLowerCase();
  if (!keyword) {
    return genreOptions.value;
  }
  return genreOptions.value.filter((genre) => genre.name.toLowerCase().includes(keyword));
});
const editForm = ref<MovieEditForm>({
  title: "",
  original_title: "",
  genre_names: [],
  tagline: "",
  release_date: "",
  status: "",
  runtime: "",
  original_language: "",
  homepage: "",
  poster_path: "",
  backdrop_path: "",
  vote_average: "",
  popularity: "",
  overview: "",
});

const movieId = computed(() => Number(route.params.id));
const shouldShowSyncPanel = computed(() => {
  return remoteDiffDecision.value === "has_diff_pending" || remoteDiffDecision.value === "keep_local";
});
const allowedSyncModes = computed<AdminSyncMode[]>(() => {
  if (remoteDiffDecision.value === "no_diff") {
    return ["update_unmodified"];
  }
  if (remoteDiffDecision.value === "has_diff_pending") {
    return ["overwrite_all", "selective"];
  }
  if (remoteDiffDecision.value === "keep_local") {
    return ["update_unmodified", "selective"];
  }
  return ["update_unmodified", "overwrite_all", "selective"];
});

function resetEditForm(data: any) {
  editForm.value = {
    title: data?.title ?? "",
    original_title: data?.original_title ?? "",
    genre_names: Array.isArray(data?.genres) ? data.genres.map((g: any) => String(g?.name ?? "").trim()).filter(Boolean) : [],
    tagline: data?.tagline ?? "",
    release_date: data?.release_date ?? "",
    status: data?.status ?? "",
    runtime: data?.runtime != null ? String(data.runtime) : "",
    original_language: data?.original_language ?? "",
    homepage: data?.homepage ?? "",
    poster_path: data?.poster_path ?? "",
    backdrop_path: data?.backdrop_path ?? "",
    vote_average: data?.vote_average != null ? String(data.vote_average) : "",
    popularity: data?.popularity != null ? String(data.popularity) : "",
    overview: data?.overview ?? "",
  };
}

function normalizeGenreOptions(raw: any): GenreOption[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item: any, idx: number) => ({
      id: Number(item?.id) || idx + 1,
      name: String(item?.name ?? "").trim(),
    }))
    .filter((item: GenreOption) => !!item.name);
}

async function loadGenreOptions() {
  try {
    const resp = await getMovieGenreList();
    const options = normalizeGenreOptions(resp.data?.genres);
    if (options.length > 0) {
      genreOptions.value = options;
      return;
    }
  } catch {
    // 忽略类型列表加载失败，降级使用详情已有类型
  }

  genreOptions.value = normalizeGenreOptions(detail.value?.genres);
}

function enterEditMode() {
  if (!detail.value) return;
  resetEditForm(detail.value);
  genreKeyword.value = "";
  saveError.value = "";
  saveMessage.value = "";
  isEditing.value = true;
}

function cancelEditMode() {
  if (detail.value) {
    resetEditForm(detail.value);
  }
  genreKeyword.value = "";
  saveError.value = "";
  isEditing.value = false;
}

async function checkRemoteDiffAndPrompt() {
  if (!movieId.value || checkingRemoteDiff.value || comparedRemoteId.value === movieId.value) {
    return;
  }
  comparedRemoteId.value = movieId.value;
  checkingRemoteDiff.value = true;
  try {
    const resp = await compareMovieRemote(movieId.value);
    const diffFields = Array.isArray(resp.data?.diff_fields) ? resp.data.diff_fields : [];
    if (!resp.data?.has_diff) {
      remoteDiffNotice.value = null;
      remoteDiffDecision.value = "no_diff";
      return;
    }

    const fieldPreview = diffFields.slice(0, 6).join("、");
    const summary = diffFields.length > 6 ? `${fieldPreview} 等 ${diffFields.length} 项` : `${fieldPreview}（共 ${diffFields.length} 项）`;
    remoteDiffNotice.value = {
      summary,
      fields: diffFields,
    };
    remoteDiffDecision.value = "has_diff_pending";
  } catch (err: any) {
    remoteDiffError.value = err.message ?? "远程差异检测失败";
  } finally {
    checkingRemoteDiff.value = false;
  }
}

async function overwriteWithRemoteData() {
  if (!movieId.value || resolvingRemoteDiff.value) {
    return;
  }

  resolvingRemoteDiff.value = true;
  remoteDiffError.value = "";
  remoteDiffMessage.value = "";
  try {
    await syncMovie(movieId.value, { mode: "overwrite_all" });
    remoteDiffNotice.value = null;
    remoteDiffDecision.value = "overwritten";
    remoteDiffMessage.value = "已覆盖并拉取最新远程数据";
    await loadData({ checkRemoteDiff: false });
  } catch (err: any) {
    remoteDiffError.value = err.message ?? "覆盖拉取失败";
  } finally {
    resolvingRemoteDiff.value = false;
  }
}

function keepLocalData() {
  remoteDiffNotice.value = null;
  remoteDiffDecision.value = "keep_local";
  remoteDiffError.value = "";
  remoteDiffMessage.value = "已保留本地数据，可在“数据库同步”区域手动更新";
}

async function loadData(options: { checkRemoteDiff?: boolean } = {}) {
  const { checkRemoteDiff = true } = options;
  if (!movieId.value) {
    error.value = "无效电影 ID";
    return;
  }
  loading.value = true;
  error.value = "";
  remoteDiffError.value = "";
  try {
    const resp = await getMovieDetail(movieId.value);
    detail.value = resp.data;
    resetEditForm(resp.data);
    await loadGenreOptions();
    genreKeyword.value = "";
    isEditing.value = false;
    if (checkRemoteDiff) {
      await checkRemoteDiffAndPrompt();
    }
  } catch (err: any) {
    error.value = err.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
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

async function saveMovieChanges() {
  if (!movieId.value) {
    saveError.value = "无效电影 ID";
    return;
  }
  const runtime = parseOptionalInt(editForm.value.runtime);
  if (editForm.value.runtime.trim() && runtime === undefined) {
    saveError.value = "时长必须是数字";
    return;
  }
  const voteAverage = parseOptionalFloat(editForm.value.vote_average);
  if (editForm.value.vote_average.trim() && voteAverage === undefined) {
    saveError.value = "评分必须是数字";
    return;
  }
  const popularity = parseOptionalFloat(editForm.value.popularity);
  if (editForm.value.popularity.trim() && popularity === undefined) {
    saveError.value = "热度必须是数字";
    return;
  }

  saving.value = true;
  saveError.value = "";
  saveMessage.value = "";
  try {
    const payload: Record<string, unknown> = {
      title: editForm.value.title.trim(),
      original_title: editForm.value.original_title.trim(),
      genre_names: editForm.value.genre_names,
      tagline: editForm.value.tagline.trim(),
      release_date: editForm.value.release_date.trim(),
      status: editForm.value.status.trim(),
      original_language: editForm.value.original_language.trim(),
      homepage: editForm.value.homepage.trim(),
      poster_path: editForm.value.poster_path.trim(),
      backdrop_path: editForm.value.backdrop_path.trim(),
      overview: editForm.value.overview.trim(),
    };
    if (runtime !== undefined) {
      payload.runtime = runtime;
    }
    if (voteAverage !== undefined) {
      payload.vote_average = voteAverage;
    }
    if (popularity !== undefined) {
      payload.popularity = popularity;
    }

    await updateMovie(movieId.value, payload);
    saveMessage.value = "已保存到本地数据库";
    isEditing.value = false;
    await loadData();
  } catch (err: any) {
    saveError.value = err.message ?? "保存失败";
  } finally {
    saving.value = false;
  }
}

onMounted(loadData);
watch(movieId, () => {
  void loadData();
});
</script>

<template>
  <p v-if="loading" class="card text-sm text-black/60">加载中...</p>
  <p v-else-if="error" class="card text-sm text-red-600">{{ error }}</p>

  <template v-else-if="detail">
    <!-- 背景横幅 -->
    <section
      class="hero-banner"
      :style="{ backgroundImage: `url(${tmdbImg(detail.backdrop_path, 'w780')})` }"
    >
      <div class="hero-overlay">
        <h1 class="text-2xl font-bold text-white md:text-3xl">{{ detail.title || detail.original_title }}</h1>
        <p class="mt-1 text-sm text-white/70">
          {{ detail.tagline }}
        </p>
      </div>
    </section>

    <!-- 主体内容 -->
    <section class="card mt-4">
      <div class="detail-layout">
        <!-- 海报 -->
        <div class="detail-poster">
          <img
            :src="tmdbImg(detail.poster_path, 'w342')"
            :alt="detail.title"
            class="w-full rounded-xl shadow-soft"
          />
        </div>

        <!-- 信息面板 -->
        <div class="detail-info">
          <h2 class="text-xl font-bold">{{ detail.title }}</h2>
          <p v-if="detail.original_title !== detail.title" class="text-sm text-black/55">
            {{ detail.original_title }}
          </p>

          <div class="mt-3 flex flex-wrap gap-2">
            <span class="badge">⭐ {{ detail.vote_average?.toFixed(1) ?? "-" }}</span>
            <span class="badge">📅 {{ detail.release_date ?? "-" }}</span>
            <span v-if="detail.runtime" class="badge">⏱ {{ detail.runtime }} 分钟</span>
            <span class="badge">{{ detail.status ?? "未知" }}</span>
          </div>

          <!-- 类型标签 -->
          <div v-if="detail.genres?.length" class="mt-3 flex flex-wrap gap-1.5">
            <span
              v-for="g in detail.genres"
              :key="g.id"
              class="rounded-full bg-sand/60 px-3 py-1 text-xs text-ink"
            >
              {{ g.name }}
            </span>
          </div>

          <p class="mt-4 text-sm leading-relaxed text-black/75">
            {{ detail.overview || "暂无简介" }}
          </p>

          <div
            v-if="checkingRemoteDiff || remoteDiffNotice || remoteDiffMessage || remoteDiffError"
            class="mt-4 rounded-xl border border-amber-200 bg-amber-50/80 p-4"
          >
            <p v-if="checkingRemoteDiff" class="text-xs text-amber-700">
              正在检测远程数据差异...
            </p>

            <template v-else-if="remoteDiffNotice">
              <p class="text-sm font-medium text-amber-800">
                检测到远程电影数据与本地不一致
              </p>
              <p class="mt-1 text-xs text-amber-700">
                变化字段：{{ remoteDiffNotice.summary }}
              </p>
              <div class="mt-3 flex flex-wrap items-center gap-2">
                <button
                  class="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700 disabled:opacity-60"
                  :disabled="resolvingRemoteDiff"
                  @click="overwriteWithRemoteData"
                >
                  {{ resolvingRemoteDiff ? "更新中..." : "覆盖拉取最新数据" }}
                </button>
                <button
                  class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100 disabled:opacity-60"
                  :disabled="resolvingRemoteDiff"
                  @click="keepLocalData"
                >
                  保留本地数据
                </button>
              </div>
            </template>

            <p v-else-if="remoteDiffMessage" class="text-xs text-green-700">
              {{ remoteDiffMessage }}
            </p>
            <p v-if="remoteDiffError" class="mt-1 text-xs text-red-600">
              {{ remoteDiffError }}
            </p>
          </div>

          <DetailSyncPanel
            v-if="shouldShowSyncPanel"
            media-type="movie"
            :target-id="movieId"
            :allowed-modes="allowedSyncModes"
            @synced="loadData"
          />

          <div class="mt-6 rounded-xl border border-black/10 bg-white/70 p-4">
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold">本地信息编辑</h3>
              <button
                v-if="!isEditing"
                class="rounded-lg border border-black/10 bg-white px-3 py-1.5 text-xs hover:bg-sand/50"
                @click="enterEditMode"
              >
                编辑
              </button>
            </div>

            <p v-if="!isEditing" class="mt-2 text-xs text-black/60">
              当前为查看模式，点击“编辑”后可修改并保存到本地数据库。
            </p>

            <div v-else class="mt-3">
              <div class="grid gap-3 md:grid-cols-2">
                <label class="text-xs text-black/60">
                  片名
                  <input
                    v-model="editForm.title"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="电影标题"
                  />
                </label>
                <label class="text-xs text-black/60">
                  原始片名
                  <input
                    v-model="editForm.original_title"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="Original Title"
                  />
                </label>
                <label class="text-xs text-black/60 md:col-span-2">
                  类型（多选）
                  <div class="mt-1 flex flex-wrap gap-2 rounded-lg border border-black/10 bg-white p-2">
                    <input
                      v-model="genreKeyword"
                      class="w-full rounded-md border border-black/10 bg-sand/20 px-2.5 py-1.5 text-xs"
                      placeholder="筛选类型"
                    />
                    <label
                      v-for="genre in filteredGenreOptions"
                      :key="genre.id"
                      class="inline-flex items-center gap-1.5 rounded-md border border-black/10 px-2 py-1 text-xs"
                    >
                      <input
                        v-model="editForm.genre_names"
                        type="checkbox"
                        :value="genre.name"
                      />
                      <span>{{ genre.name }}</span>
                    </label>
                    <span v-if="!genreOptions.length" class="px-1 py-1 text-xs text-black/50">
                      暂无可选类型
                    </span>
                    <span v-else-if="!filteredGenreOptions.length" class="px-1 py-1 text-xs text-black/50">
                      无匹配类型
                    </span>
                  </div>
                </label>
                <label class="text-xs text-black/60">
                  上映日期
                  <input
                    v-model="editForm.release_date"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="YYYY-MM-DD"
                  />
                </label>
                <label class="text-xs text-black/60">
                  状态
                  <input
                    v-model="editForm.status"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="Released"
                  />
                </label>
                <label class="text-xs text-black/60">
                  标语
                  <input
                    v-model="editForm.tagline"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="Tagline"
                  />
                </label>
                <label class="text-xs text-black/60">
                  时长(分钟)
                  <input
                    v-model="editForm.runtime"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="Runtime"
                  />
                </label>
                <label class="text-xs text-black/60">
                  原始语言
                  <input
                    v-model="editForm.original_language"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="zh / en"
                  />
                </label>
                <label class="text-xs text-black/60">
                  主页链接
                  <input
                    v-model="editForm.homepage"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="https://..."
                  />
                </label>
                <label class="text-xs text-black/60">
                  海报路径
                  <input
                    v-model="editForm.poster_path"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="/poster.jpg"
                  />
                </label>
                <label class="text-xs text-black/60">
                  背景图路径
                  <input
                    v-model="editForm.backdrop_path"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="/backdrop.jpg"
                  />
                </label>
                <label class="text-xs text-black/60">
                  评分
                  <input
                    v-model="editForm.vote_average"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="7.8"
                  />
                </label>
                <label class="text-xs text-black/60">
                  热度
                  <input
                    v-model="editForm.popularity"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="123.45"
                  />
                </label>
                <label class="text-xs text-black/60 md:col-span-2">
                  简介
                  <textarea
                    v-model="editForm.overview"
                    rows="4"
                    class="mt-1 w-full rounded-lg border border-black/10 bg-white px-3 py-2 text-sm"
                    placeholder="简介"
                  />
                </label>
              </div>

              <div class="mt-3 flex items-center gap-3">
                <button
                  class="rounded-lg bg-coral px-4 py-2 text-sm font-medium text-white hover:bg-coral/90 disabled:opacity-60"
                  :disabled="saving"
                  @click="saveMovieChanges"
                >
                  {{ saving ? "保存中..." : "保存到本地数据库" }}
                </button>
                <button
                  class="rounded-lg border border-black/10 bg-white px-4 py-2 text-sm hover:bg-sand/50 disabled:opacity-60"
                  :disabled="saving"
                  @click="cancelEditMode"
                >
                  取消
                </button>
              </div>
            </div>

            <div class="mt-2">
              <span v-if="saveMessage" class="text-xs text-green-700">{{ saveMessage }}</span>
              <span v-if="saveError" class="text-xs text-red-600">{{ saveError }}</span>
            </div>
          </div>

          <!-- 演员 -->
          <div v-if="detail.credits?.cast?.length" class="mt-6">
            <h3 class="mb-2 text-sm font-semibold">主要演员</h3>
            <div class="cast-grid">
              <div v-for="c in detail.credits.cast.slice(0, 8)" :key="c.id" class="cast-card">
                <RouterLink :to="`/person/${c.id}`">
                  <img
                    :src="tmdbImg(c.profile_path, 'w185')"
                    :alt="c.name"
                    class="cast-img"
                    loading="lazy"
                  />
                </RouterLink>
                <p class="mt-1 truncate text-xs font-medium">{{ c.name }}</p>
                <p class="truncate text-xs text-black/50">{{ c.character }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </template>
</template>
