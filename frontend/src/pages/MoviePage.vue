<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { compareMovieRemote, deleteMovie, updateMovie } from "@/api/admin";
import { prefetchMediaDetail } from "@/api/prefetch";
import type { AdminCompareFieldDetail, AdminSyncMode } from "@/api/admin";
import { getMovieCredits, getMovieDetail, getMovieGenreList } from "@/api/movie";
import { tmdbImg } from "@/api/tmdb";
import { formatStatusLabel, movieStatusOptions } from "@/constants/mediaStatus";
import { scheduleAfterPaint } from "@/utils/schedule";

const DetailSyncPanel = defineAsyncComponent(() => import("@/components/DetailSyncPanel.vue"));
const GlassSelect = defineAsyncComponent(() => import("@/components/GlassSelect.vue"));

type GenreOption = {
  id: number;
  name: string;
};

type MovieCastMember = {
  id: number;
  name: string;
  character: string;
  profile_path: string;
};

type MovieEditForm = {
  tmdb_id: string;
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
  remoteSummary: string;
  localOverrideSummary: string;
  remoteFields: string[];
  localOverrideFields: string[];
  remoteDetails: AdminCompareFieldDetail[];
  localOverrideDetails: AdminCompareFieldDetail[];
};

type RemoteDiffDecision = "unknown" | "has_diff_pending" | "keep_local" | "overwritten" | "no_diff";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const error = ref("");
const detail = ref<any>(null);
const castMembers = ref<MovieCastMember[]>([]);
const creditsLoading = ref(false);
const creditsLoaded = ref(false);
const creditsError = ref("");
const isEditing = ref(false);
const saving = ref(false);
const deleting = ref(false);
const saveError = ref("");
const saveMessage = ref("");
const deleteError = ref("");
const comparedRemoteId = ref<number | null>(null);
const checkingRemoteDiff = ref(false);
const remoteDiffNotice = ref<RemoteDiffNotice | null>(null);
const remoteDiffMessage = ref("");
const remoteDiffError = ref("");
const remoteDiffDecision = ref<RemoteDiffDecision>("unknown");
const showRemoteDiffDetails = ref(false);
const showLocalOverrideDiffDetails = ref(false);
const tmdbRiskModalVisible = ref(false);
const tmdbRiskCurrentId = ref<number | null>(null);
const tmdbRiskNextId = ref<number | null>(null);
let tmdbRiskConfirmResolver: ((confirmed: boolean) => void) | null = null;
const deleteConfirmModalVisible = ref(false);
const genreOptions = ref<GenreOption[]>([]);
const genreOptionsLoaded = ref(false);
const genreKeyword = ref("");
let loadReqSeq = 0;
let creditsReqSeq = 0;
let cancelDeferredLoads: (() => void) | null = null;
const filteredGenreOptions = computed(() => {
  const keyword = genreKeyword.value.trim().toLowerCase();
  if (!keyword) {
    return genreOptions.value;
  }
  return genreOptions.value.filter((genre) => genre.name.toLowerCase().includes(keyword));
});
const editForm = ref<MovieEditForm>({
  tmdb_id: "",
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
const currentTmdbId = computed(() => Number(detail.value?.id ?? movieId.value ?? 0));
const originalTmdbId = computed(() => Number(detail.value?.sync_tmdb_id ?? detail.value?.id ?? movieId.value ?? 0));
const hasRewrittenTmdbId = computed(() => {
  return originalTmdbId.value > 0 && currentTmdbId.value > 0 && originalTmdbId.value !== currentTmdbId.value;
});
const hasRemoteOnlyDiff = computed(() => (remoteDiffNotice.value?.remoteFields.length ?? 0) > 0);
const hasLocalOverrideDiff = computed(() => (remoteDiffNotice.value?.localOverrideFields.length ?? 0) > 0);
const shouldShowSyncPanel = computed(() => {
  return remoteDiffDecision.value === "has_diff_pending";
});
const allowedSyncModes = computed<AdminSyncMode[]>(() => {
  if (remoteDiffDecision.value === "no_diff") {
    return ["update_unmodified"];
  }
  if (remoteDiffDecision.value === "has_diff_pending") {
    if (hasRemoteOnlyDiff.value && hasLocalOverrideDiff.value) {
      return ["update_unmodified", "overwrite_all", "selective"];
    }
    if (hasRemoteOnlyDiff.value) {
      return ["update_unmodified", "overwrite_all"];
    }
    return ["overwrite_all", "selective"];
  }
  if (remoteDiffDecision.value === "keep_local") {
    if (hasRemoteOnlyDiff.value && hasLocalOverrideDiff.value) {
      return ["update_unmodified", "overwrite_all", "selective"];
    }
    if (hasRemoteOnlyDiff.value) {
      return ["update_unmodified", "overwrite_all"];
    }
    return ["overwrite_all", "selective"];
  }
  return ["update_unmodified", "overwrite_all", "selective"];
});

function goBack() {
  void router.push({
    path: "/library",
    query: { tab: "movie" },
  });
}

function personLink(personId: number) {
  return {
    path: `/person/${personId}`,
    query: {
      fromType: "movie",
      fromId: String(movieId.value),
    },
  };
}

function prefetchPerson(personId: number) {
  prefetchMediaDetail("person", personId);
}

function resetEditForm(data: any) {
  editForm.value = {
    tmdb_id: data?.id != null ? String(data.id) : String(movieId.value || ""),
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

function normalizeCastMembers(raw: unknown): MovieCastMember[] {
  const cast = Array.isArray((raw as Record<string, unknown> | null)?.cast)
    ? ((raw as Record<string, unknown>).cast as unknown[])
    : [];
  return cast
    .map((item: any) => ({
      id: Number(item?.id) || 0,
      name: String(item?.name ?? "").trim(),
      character: String(item?.character ?? "").trim(),
      profile_path: String(item?.profile_path ?? ""),
    }))
    .filter((item) => item.id > 0 && item.name)
    .slice(0, 8);
}

function resetRemoteDiffState() {
  remoteDiffNotice.value = null;
  remoteDiffMessage.value = "";
  remoteDiffError.value = "";
  remoteDiffDecision.value = "unknown";
  showRemoteDiffDetails.value = false;
  showLocalOverrideDiffDetails.value = false;
  checkingRemoteDiff.value = false;
  comparedRemoteId.value = null;
}

function resetCreditsState() {
  creditsReqSeq++;
  castMembers.value = [];
  creditsLoading.value = false;
  creditsLoaded.value = false;
  creditsError.value = "";
}

function stopDeferredLoads() {
  if (cancelDeferredLoads) {
    cancelDeferredLoads();
    cancelDeferredLoads = null;
  }
}

function scheduleDeferredLoadsForDetail() {
  stopDeferredLoads();
  cancelDeferredLoads = scheduleAfterPaint(() => {
    void loadMovieCredits();
  });
}

async function loadGenreOptions(force = false) {
  if (!force && genreOptionsLoaded.value) {
    return;
  }
  try {
    const resp = await getMovieGenreList();
    const options = normalizeGenreOptions(resp.data?.genres);
    if (options.length > 0) {
      genreOptions.value = options;
      genreOptionsLoaded.value = true;
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
  if (!genreOptionsLoaded.value) {
    void loadGenreOptions();
  }
}

function cancelEditMode() {
  if (detail.value) {
    resetEditForm(detail.value);
  }
  genreKeyword.value = "";
  saveError.value = "";
  isEditing.value = false;
}

function closeTmdbRiskModal(confirmed: boolean) {
  tmdbRiskModalVisible.value = false;
  const resolver = tmdbRiskConfirmResolver;
  tmdbRiskConfirmResolver = null;
  tmdbRiskCurrentId.value = null;
  tmdbRiskNextId.value = null;
  if (resolver) {
    resolver(confirmed);
  }
}

function askTmdbRiskConfirm(currentId: number, nextId: number): Promise<boolean> {
  tmdbRiskCurrentId.value = currentId;
  tmdbRiskNextId.value = nextId;
  tmdbRiskModalVisible.value = true;
  return new Promise((resolve) => {
    tmdbRiskConfirmResolver = resolve;
  });
}

async function deleteCurrentMovie() {
  if (!movieId.value) {
    deleteError.value = "无效电影 ID";
    return;
  }
  deleteConfirmModalVisible.value = true;
}

function closeDeleteConfirmModal() {
  deleteConfirmModalVisible.value = false;
}

async function confirmDeleteCurrentMovie() {
  if (!movieId.value) {
    deleteError.value = "无效电影 ID";
    deleteConfirmModalVisible.value = false;
    return;
  }

  deleting.value = true;
  deleteError.value = "";
  try {
    deleteConfirmModalVisible.value = false;
    await deleteMovie(movieId.value);
    await router.push({
      path: "/library",
      query: { tab: "movie" },
    });
  } catch (err: any) {
    deleteError.value = err.message ?? "删除失败";
  } finally {
    deleting.value = false;
  }
}

async function loadMovieCredits(force = false) {
  if (!movieId.value || creditsLoading.value || (creditsLoaded.value && !force)) {
    return;
  }

  const requestSeq = ++creditsReqSeq;
  const targetId = movieId.value;
  creditsLoading.value = true;
  creditsError.value = "";
  try {
    const resp = await getMovieCredits(targetId, "zh-CN", { force });
    if (requestSeq !== creditsReqSeq || targetId !== movieId.value) {
      return;
    }
    castMembers.value = normalizeCastMembers(resp.data);
    creditsLoaded.value = true;
  } catch (err: any) {
    if (requestSeq !== creditsReqSeq || targetId !== movieId.value) {
      return;
    }
    creditsError.value = err.message ?? "加载演员失败";
  } finally {
    if (requestSeq === creditsReqSeq) {
      creditsLoading.value = false;
    }
  }
}

async function checkRemoteDiffAndPrompt(force = false) {
  if (!movieId.value || checkingRemoteDiff.value || (!force && comparedRemoteId.value === movieId.value)) {
    return;
  }
  if (movieId.value < 0) {
    remoteDiffNotice.value = null;
    showRemoteDiffDetails.value = false;
    showLocalOverrideDiffDetails.value = false;
    remoteDiffDecision.value = "keep_local";
    remoteDiffError.value = "";
    remoteDiffMessage.value = "本地新建条目不参与 TMDB 远程差异检测";
    comparedRemoteId.value = movieId.value;
    return;
  }
  checkingRemoteDiff.value = true;
  remoteDiffError.value = "";
  try {
    const resp = await compareMovieRemote(movieId.value);
    const remoteFields = Array.isArray(resp.data?.diff_fields) ? resp.data.diff_fields : [];
    const localOverrideFields = Array.isArray(resp.data?.local_override_diff_fields) ? resp.data.local_override_diff_fields : [];
    const hasDiff = Boolean(resp.data?.has_diff) && (remoteFields.length > 0 || localOverrideFields.length > 0);
    if (!hasDiff) {
      remoteDiffNotice.value = null;
      showRemoteDiffDetails.value = false;
      showLocalOverrideDiffDetails.value = false;
      remoteDiffDecision.value = "no_diff";
      remoteDiffMessage.value = "";
      comparedRemoteId.value = movieId.value;
      return;
    }

    const remoteFieldPreview = remoteFields.slice(0, 6).join("、");
    const remoteSummary = remoteFields.length === 0
      ? "无"
      : remoteFields.length > 6
        ? `${remoteFieldPreview} 等 ${remoteFields.length} 项`
        : `${remoteFieldPreview}（共 ${remoteFields.length} 项）`;
    const localOverridePreview = localOverrideFields.slice(0, 6).join("、");
    const localOverrideSummary = localOverrideFields.length === 0
      ? "无"
      : localOverrideFields.length > 6
        ? `${localOverridePreview} 等 ${localOverrideFields.length} 项`
        : `${localOverridePreview}（共 ${localOverrideFields.length} 项）`;
    const detailItems = normalizeDiffDetails(resp.data?.diff_details);
    const remoteDetails = buildDiffDetailsByFields(remoteFields, detailItems, "remote");
    const localOverrideDetails = buildDiffDetailsByFields(localOverrideFields, detailItems, "local_override");
    remoteDiffNotice.value = {
      remoteSummary,
      localOverrideSummary,
      remoteFields,
      localOverrideFields,
      remoteDetails,
      localOverrideDetails,
    };
    showRemoteDiffDetails.value = false;
    showLocalOverrideDiffDetails.value = false;
    remoteDiffMessage.value = "";
    remoteDiffDecision.value = "has_diff_pending";
    comparedRemoteId.value = movieId.value;
  } catch (err: any) {
    remoteDiffError.value = err.message ?? "远程差异检测失败";
  } finally {
    checkingRemoteDiff.value = false;
  }
}

function keepLocalData() {
  remoteDiffNotice.value = null;
  showRemoteDiffDetails.value = false;
  showLocalOverrideDiffDetails.value = false;
  remoteDiffDecision.value = "keep_local";
  remoteDiffError.value = "";
  remoteDiffMessage.value = "已保留本地数据，已跳过本次远程差异处理";
}

function handleSynced() {
  comparedRemoteId.value = null;
  void loadData({ force: true });
}

async function loadData(options: { force?: boolean; checkRemoteDiff?: boolean } = {}) {
  const { force = false, checkRemoteDiff = true } = options;
  if (!movieId.value) {
    error.value = "无效电影 ID";
    return;
  }
  const requestSeq = ++loadReqSeq;
  stopDeferredLoads();
  loading.value = true;
  error.value = "";
  resetRemoteDiffState();
  resetCreditsState();
  try {
    const resp = await getMovieDetail(movieId.value, "zh-CN", "", { force });
    if (requestSeq !== loadReqSeq) {
      return;
    }
    detail.value = resp.data;
    resetEditForm(resp.data);
    genreOptions.value = normalizeGenreOptions(resp.data?.genres);
    genreOptionsLoaded.value = false;
    genreKeyword.value = "";
    isEditing.value = false;
    if (checkRemoteDiff) {
      await checkRemoteDiffAndPrompt();
    }
    scheduleDeferredLoadsForDetail();
  } catch (err: any) {
    if (requestSeq === loadReqSeq) {
      error.value = err.message ?? "加载失败";
    }
  } finally {
    if (requestSeq === loadReqSeq) {
      loading.value = false;
    }
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

function normalizeDiffDetails(raw: unknown): AdminCompareFieldDetail[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item: any) => ({
      field: String(item?.field ?? "").trim(),
      diff_type: String(item?.diff_type ?? "remote").trim() || "remote",
      local: String(item?.local ?? "-"),
      remote: String(item?.remote ?? "-"),
    }))
    .filter((item) => item.field.length > 0);
}

function buildDiffDetailsByFields(
  fields: string[],
  details: AdminCompareFieldDetail[],
  diffType: "remote" | "local_override",
): AdminCompareFieldDetail[] {
  const detailMap = new Map(
    details
      .filter((item) => item.diff_type === diffType)
      .map((item) => [item.field, item]),
  );
  return fields.map((field) => detailMap.get(field) ?? {
    field,
    diff_type: diffType,
    local: "-",
    remote: "-",
  });
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

  const rawTmdbID = editForm.value.tmdb_id.trim();
  const nextTmdbID = parseOptionalInt(rawTmdbID);
  const tmdbChanged = nextTmdbID !== undefined && nextTmdbID !== movieId.value;
  if (tmdbChanged) {
    if (nextTmdbID === undefined || nextTmdbID <= 0) {
      saveError.value = "TMDB ID 必须是大于 0 的整数";
      return;
    }
    const riskConfirm = await askTmdbRiskConfirm(movieId.value, nextTmdbID);
    if (!riskConfirm) {
      return;
    }
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
    if (tmdbChanged && nextTmdbID !== undefined) {
      payload.tmdb_id = nextTmdbID;
    }

    await updateMovie(movieId.value, payload);
    saveMessage.value = "已保存到本地数据库";
    isEditing.value = false;
    comparedRemoteId.value = null;
    if (tmdbChanged && nextTmdbID !== undefined) {
      await router.replace(`/movie/${nextTmdbID}`);
      return;
    }
    await loadData({ force: true });
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

onBeforeUnmount(() => {
  loadReqSeq++;
  creditsReqSeq++;
  stopDeferredLoads();
});
</script>

<template>
  <p v-if="loading" class="card text-sm text-black/60">加载中...</p>
  <p v-else-if="error" class="card text-sm text-red-600">{{ error }}</p>

  <template v-else-if="detail">
    <!-- 背景横幅 -->
    <section class="hero-banner hero-banner-detail">
      <img
        :src="tmdbImg(detail.backdrop_path, 'w780')"
        :alt="detail.title || detail.original_title"
        class="hero-banner-media"
      />
      <div class="absolute left-4 top-4 z-10">
        <button
          class="rounded-lg border border-white/40 bg-black/40 px-3 py-1.5 text-xs text-white hover:bg-black/55"
          @click="goBack"
        >
          返回上一页
        </button>
      </div>
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
          <div class="mt-2 grid gap-1 text-xs text-black/60 sm:grid-cols-2">
            <template v-if="hasRewrittenTmdbId">
              <p>
                修改后 TMDB ID：
                <span class="font-medium text-black">{{ currentTmdbId }}</span>
              </p>
              <p>
                原始 TMDB ID：
                <span class="font-medium text-black">{{ originalTmdbId }}</span>
              </p>
            </template>
            <p v-else>
              TMDB ID：
              <span class="font-medium text-black">{{ currentTmdbId }}</span>
            </p>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <span class="badge">⭐ {{ detail.vote_average?.toFixed(1) ?? "-" }}</span>
            <span class="badge">📅 {{ detail.release_date ?? "-" }}</span>
            <span v-if="detail.runtime" class="badge">⏱ {{ detail.runtime }} 分钟</span>
            <span class="badge">{{ formatStatusLabel(detail.status) }}</span>
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
            v-if="checkingRemoteDiff || remoteDiffNotice || remoteDiffMessage || remoteDiffError || remoteDiffDecision === 'no_diff'"
            class="mt-4 rounded-xl border border-amber-200 bg-amber-50/80 p-4"
          >
            <p v-if="checkingRemoteDiff" class="text-xs text-amber-700">
              正在检测远程数据差异...
            </p>

            <template v-else-if="remoteDiffNotice">
              <p class="mt-3 text-sm font-medium text-amber-800">
                检测到远程电影数据与本地不一致
              </p>
              <p class="mt-1 text-xs text-amber-700">
                远程变化字段：{{ remoteDiffNotice.remoteSummary }}
              </p>
              <p class="mt-1 text-xs text-amber-700">
                本地修改字段：{{ remoteDiffNotice.localOverrideSummary }}
              </p>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <button
                  v-if="remoteDiffNotice.remoteDetails.length"
                  class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100"
                  @click="showRemoteDiffDetails = !showRemoteDiffDetails"
                >
                  {{ showRemoteDiffDetails ? "收起远程变化明细" : "查看远程变化明细" }}
                </button>
                <button
                  v-if="remoteDiffNotice.localOverrideDetails.length"
                  class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100"
                  @click="showLocalOverrideDiffDetails = !showLocalOverrideDiffDetails"
                >
                  {{ showLocalOverrideDiffDetails ? "收起本地修改明细" : "查看本地修改明细" }}
                </button>
                <button
                  class="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs text-amber-700 hover:bg-amber-100 disabled:opacity-60"
                  @click="keepLocalData"
                >
                  暂不处理，保留本地
                </button>
              </div>

              <div
                v-if="showRemoteDiffDetails && remoteDiffNotice.remoteDetails.length"
                class="mt-2 space-y-2 rounded-lg border border-amber-200 bg-white/70 p-2"
              >
                <div
                  v-for="item in remoteDiffNotice.remoteDetails"
                  :key="`remote-${item.field}`"
                  class="rounded-md bg-amber-50/70 p-2"
                >
                  <p class="text-xs font-semibold text-amber-900">{{ item.field }}</p>
                  <p class="mt-1 text-xs text-amber-800">本地：{{ item.local }}</p>
                  <p class="mt-1 text-xs text-amber-800">远程：{{ item.remote }}</p>
                </div>
              </div>

              <div
                v-if="showLocalOverrideDiffDetails && remoteDiffNotice.localOverrideDetails.length"
                class="mt-2 space-y-2 rounded-lg border border-amber-200 bg-white/70 p-2"
              >
                <div
                  v-for="item in remoteDiffNotice.localOverrideDetails"
                  :key="`local-${item.field}`"
                  class="rounded-md bg-amber-50/70 p-2"
                >
                  <p class="text-xs font-semibold text-amber-900">{{ item.field }}</p>
                  <p class="mt-1 text-xs text-amber-800">本地：{{ item.local }}</p>
                  <p class="mt-1 text-xs text-amber-800">远程：{{ item.remote }}</p>
                </div>
              </div>
            </template>

            <DetailSyncPanel
              v-if="shouldShowSyncPanel"
              media-type="movie"
              :target-id="movieId"
              :allowed-modes="allowedSyncModes"
              :preset-changed-fields="remoteDiffNotice?.localOverrideFields ?? []"
              :embedded="true"
              @synced="handleSynced"
            />

            <p v-if="!checkingRemoteDiff && !remoteDiffNotice && remoteDiffDecision === 'no_diff'" class="mt-3 text-xs text-green-700">
              已完成检查，当前未发现远程差异。
            </p>
            <p v-if="!checkingRemoteDiff && !remoteDiffNotice && remoteDiffMessage" class="mt-3 text-xs text-green-700">
              {{ remoteDiffMessage }}
            </p>
            <p v-if="remoteDiffError" class="mt-1 text-xs text-red-600">
              {{ remoteDiffError }}
            </p>
          </div>

          <div class="panel-glass content-auto mt-6 rounded-xl p-4">
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold">本地信息编辑</h3>
              <div class="flex items-center gap-2">
                <button
                  class="btn-danger-soft-xs disabled:opacity-60"
                  :disabled="deleting || saving"
                  @click="deleteCurrentMovie"
                >
                  {{ deleting ? "删除中..." : "删除本地数据" }}
                </button>
                <button
                  v-if="!isEditing"
                  class="btn-soft-xs"
                  @click="enterEditMode"
                >
                  编辑
                </button>
              </div>
            </div>

            <p v-if="!isEditing" class="mt-2 text-xs text-black/60">
              当前为查看模式，点击“编辑”后可修改并保存到本地数据库。
            </p>

            <div v-else class="mt-3">
              <div class="grid gap-3 md:grid-cols-2">
                <label class="text-xs text-black/60">
                  TMDB ID
                  <input
                    v-model="editForm.tmdb_id"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="例如：550"
                  />
                  <p class="mt-1 text-[11px] text-amber-700">
                    高风险：改动后，后续同步仍使用旧 TMDB ID 拉取；对外返回与访问使用新 TMDB ID。
                  </p>
                </label>
                <label class="text-xs text-black/60">
                  片名
                  <input
                    v-model="editForm.title"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="电影标题"
                  />
                </label>
                <label class="text-xs text-black/60">
                  原始片名
                  <input
                    v-model="editForm.original_title"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Original Title"
                  />
                </label>
                <label class="text-xs text-black/60 md:col-span-2">
                  类型（多选）
                  <div class="mt-1 flex flex-wrap gap-2 rounded-lg border border-white/70 bg-white/55 p-2">
                    <input
                      v-model="genreKeyword"
                      class="field-control-xs w-full"
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
                        class="check-control"
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
                    class="field-control mt-1 w-full text-sm"
                    placeholder="YYYY-MM-DD"
                  />
                </label>
                <label class="text-xs text-black/60">
                  状态
                  <GlassSelect v-model="editForm.status" :options="movieStatusOptions" class="mt-1 w-full" />
                </label>
                <label class="text-xs text-black/60">
                  标语
                  <input
                    v-model="editForm.tagline"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Tagline"
                  />
                </label>
                <label class="text-xs text-black/60">
                  时长(分钟)
                  <input
                    v-model="editForm.runtime"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Runtime"
                  />
                </label>
                <label class="text-xs text-black/60">
                  原始语言
                  <input
                    v-model="editForm.original_language"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="zh / en"
                  />
                </label>
                <label class="text-xs text-black/60">
                  主页链接
                  <input
                    v-model="editForm.homepage"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="https://..."
                  />
                </label>
                <label class="text-xs text-black/60">
                  海报路径
                  <input
                    v-model="editForm.poster_path"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="/poster.jpg"
                  />
                </label>
                <label class="text-xs text-black/60">
                  背景图路径
                  <input
                    v-model="editForm.backdrop_path"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="/backdrop.jpg"
                  />
                </label>
                <label class="text-xs text-black/60">
                  评分
                  <input
                    v-model="editForm.vote_average"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="7.8"
                  />
                </label>
                <label class="text-xs text-black/60">
                  热度
                  <input
                    v-model="editForm.popularity"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="123.45"
                  />
                </label>
                <label class="text-xs text-black/60 md:col-span-2">
                  简介
                  <textarea
                    v-model="editForm.overview"
                    rows="4"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="简介"
                  />
                </label>
              </div>

              <div class="mt-3 flex items-center gap-3">
                <button
                  class="btn-primary disabled:opacity-60"
                  :disabled="saving"
                  @click="saveMovieChanges"
                >
                  {{ saving ? "保存中..." : "保存到本地数据库" }}
                </button>
                <button
                  class="btn-soft disabled:opacity-60"
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
              <span v-if="deleteError" class="ml-2 text-xs text-red-600">{{ deleteError }}</span>
            </div>
          </div>

          <!-- 演员 -->
          <div class="content-auto mt-6">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-semibold">主要演员</h3>
              <button
                class="btn-soft-xs px-3 py-1 disabled:opacity-60"
                :disabled="creditsLoading"
                @click="loadMovieCredits(true)"
              >
                {{ creditsLoading ? "加载中..." : (creditsLoaded ? "刷新演员" : "加载演员") }}
              </button>
            </div>
            <p v-if="creditsLoading" class="text-xs text-black/55">正在加载演员信息...</p>
            <p v-else-if="creditsError" class="text-xs text-red-600">{{ creditsError }}</p>
            <p v-else-if="creditsLoaded && !castMembers.length" class="text-xs text-black/55">暂无演员数据</p>
            <div v-else-if="castMembers.length" class="cast-grid">
              <div v-for="c in castMembers" :key="c.id" class="cast-card">
                <RouterLink
                  :to="personLink(c.id)"
                  @mouseenter="prefetchPerson(c.id)"
                  @focus="prefetchPerson(c.id)"
                  @touchstart.passive="prefetchPerson(c.id)"
                >
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

  <div
    v-if="tmdbRiskModalVisible"
    class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/45 p-4"
    @click.self="closeTmdbRiskModal(false)"
  >
    <section class="panel-glass w-full max-w-md rounded-2xl p-5">
      <h3 class="text-base font-semibold text-amber-800">修改 TMDB ID 风险确认</h3>
      <p class="mt-2 text-sm text-black/75">
        你正在修改电影 TMDB ID：
        <span class="font-medium">{{ tmdbRiskCurrentId }}</span>
        ->
        <span class="font-medium">{{ tmdbRiskNextId }}</span>
      </p>
      <div class="mt-3 rounded-lg border border-amber-200 bg-amber-50/80 p-3 text-xs leading-relaxed text-amber-800">
        <p>1) 这是高风险操作，可能导致与第三方历史引用不一致；</p>
        <p>2) 之后自动/手动同步将继续使用旧 TMDB ID 向 TMDB 拉取；</p>
        <p>3) 对外返回与页面访问将使用新的 TMDB ID。</p>
      </div>

      <div class="mt-4 flex items-center justify-end gap-2">
        <button class="btn-soft" @click="closeTmdbRiskModal(false)">取消</button>
        <button class="btn-primary" @click="closeTmdbRiskModal(true)">确认继续</button>
      </div>
    </section>
  </div>

  <div
    v-if="deleteConfirmModalVisible"
    class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/45 p-4"
    @click.self="closeDeleteConfirmModal"
  >
    <section class="panel-glass w-full max-w-md rounded-2xl p-5">
      <h3 class="text-base font-semibold text-red-700">删除本地数据确认</h3>
      <p class="mt-2 text-sm text-black/75">
        确认删除电影
        <span class="font-medium">{{ detail?.title || detail?.original_title || `ID ${movieId}` }}</span>
        的本地数据吗？
      </p>
      <p class="mt-2 text-xs text-red-700">删除后不可恢复。</p>

      <div class="mt-4 flex items-center justify-end gap-2">
        <button class="btn-soft" :disabled="deleting" @click="closeDeleteConfirmModal">取消</button>
        <button class="btn-danger-soft" :disabled="deleting" @click="confirmDeleteCurrentMovie">
          {{ deleting ? "删除中..." : "确认删除" }}
        </button>
      </div>
    </section>
  </div>
</template>
