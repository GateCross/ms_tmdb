import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { compareMovieRemote, deleteMovie, updateMovie, uploadAdminImage } from "@/api/admin";

import type { AdminCompareFieldDetail, AdminSyncMode } from "@/api/admin";
import { clearMovieCache, getMovieCredits, getMovieDetail, getMovieGenreList } from "@/api/movie";
import type {
  GenreOption,
  MovieCastMember,
  MovieDetail,
  MovieEditForm,
  RemoteDiffDecision,
  RemoteDiffNotice,
} from "@/components/movie/types";
import { resolveErrorMessage } from "@/utils/errors";
import {
  normalizeCastMembers,
  normalizeGenreOptions,
  normalizeMovieEditForm,
} from "@/utils/mediaNormalizers";
import { scheduleAfterPaint } from "@/utils/schedule";

export function useMovieDetail() {
  const route = useRoute();
  const router = useRouter();

  const loading = ref(false);
  const loadError = ref("");
  const refreshError = ref("");
  const detail = ref<MovieDetail | null>(null);
  const castMembers = ref<MovieCastMember[]>([]);
  const creditsLoading = ref(false);
  const creditsLoaded = ref(false);
  const creditsError = ref("");
  const isEditing = ref(false);
  const saving = ref(false);
  const deleting = ref(false);
  const uploadingField = ref<"" | "poster_path" | "backdrop_path">("");
  const saveMessage = ref("");
  const comparedRemoteId = ref<number | null>(null);
  const checkingRemoteDiff = ref(false);
  const remoteDiffNotice = ref<RemoteDiffNotice | null>(null);
  const remoteDiffMessage = ref("");
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
    const historyState = window.history.state as { back?: string } | null;
    if (historyState?.back) {
      router.back();
      return;
    }
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

  function updateGenreKeyword(value: string) {
    genreKeyword.value = value;
  }

  function toggleRemoteDiffDetails() {
    showRemoteDiffDetails.value = !showRemoteDiffDetails.value;
  }

  function toggleLocalOverrideDiffDetails() {
    showLocalOverrideDiffDetails.value = !showLocalOverrideDiffDetails.value;
  }

  function resetEditForm(data: unknown) {
    editForm.value = normalizeMovieEditForm(data, movieId.value);
  }

  function resetRemoteDiffState() {
    remoteDiffNotice.value = null;
    remoteDiffMessage.value = "";
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
      // 类型列表为编辑辅助资源，失败静默并降级
      const resp = await getMovieGenreList("zh-CN", { showErrorToast: false });
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
    isEditing.value = false;
  }

  async function uploadEditImage(field: "poster_path" | "backdrop_path", event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    uploadingField.value = field;
    try {
      const resp = await uploadAdminImage(file);
      const path = String(resp.data?.path ?? "").trim();
      if (!path) {
        throw new Error("上传成功但未返回图片路径");
      }
      editForm.value[field] = path;
    } catch {
      /* handled by global toast */
    } finally {
      uploadingField.value = "";
      input.value = "";
    }
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
      return;
    }
    deleteConfirmModalVisible.value = true;
  }

  function closeDeleteConfirmModal() {
    deleteConfirmModalVisible.value = false;
  }

  async function confirmDeleteCurrentMovie() {
    if (!movieId.value) {
      deleteConfirmModalVisible.value = false;
      return;
    }

    deleting.value = true;
    try {
      deleteConfirmModalVisible.value = false;
      const deletedId = movieId.value;
      await deleteMovie(deletedId);
      clearMovieCache(deletedId);
      await router.push({
        path: "/library",
        query: { tab: "movie" },
      });
    } catch {
      /* handled by global toast */
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
      // 演员为辅助资源：静默失败，由区域状态处理
      const resp = await getMovieCredits(targetId, "zh-CN", { force, showErrorToast: false });
      if (requestSeq !== creditsReqSeq || targetId !== movieId.value) {
        return;
      }
      castMembers.value = normalizeCastMembers(resp.data);
      creditsLoaded.value = true;
      creditsError.value = "";
    } catch (error) {
      if (requestSeq !== creditsReqSeq || targetId !== movieId.value) {
        return;
      }
      creditsError.value = resolveErrorMessage(error, "演员加载失败，请重试");
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
      remoteDiffMessage.value = "本地新建条目不参与 TMDB 远程差异检测";
      comparedRemoteId.value = movieId.value;
      return;
    }
    checkingRemoteDiff.value = true;
    try {
      // 远程差异为辅助检查：静默失败，不打断详情浏览
      const resp = await compareMovieRemote(movieId.value, { showErrorToast: false });
      const remoteFields = Array.isArray(resp.data?.diff_fields) ? resp.data.diff_fields : [];
      const localOverrideFields = Array.isArray(resp.data?.local_override_diff_fields)
        ? resp.data.local_override_diff_fields
        : [];
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

      const detailItems = normalizeDiffDetails(resp.data?.diff_details);
      const remoteDetails = buildDiffDetailsByFields(remoteFields, detailItems, "remote");
      const localOverrideDetails = buildDiffDetailsByFields(localOverrideFields, detailItems, "local_override");
      remoteDiffNotice.value = {
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
    } catch {
      /* handled by global toast */
    } finally {
      checkingRemoteDiff.value = false;
    }
  }

  function keepLocalData() {
    remoteDiffNotice.value = null;
    showRemoteDiffDetails.value = false;
    showLocalOverrideDiffDetails.value = false;
    remoteDiffDecision.value = "keep_local";
    remoteDiffMessage.value = "已保留本地数据，已跳过本次远程差异处理";
  }

  function handleSynced() {
    comparedRemoteId.value = null;
    if (movieId.value) {
      clearMovieCache(movieId.value);
    }
    void loadData({ force: true });
  }

  async function loadData(options: { force?: boolean; checkRemoteDiff?: boolean } = {}) {
    const { force = false, checkRemoteDiff = true } = options;
    if (!movieId.value) {
      return;
    }
    const requestSeq = ++loadReqSeq;
    // 仅同 ID 详情视为“已有数据刷新”；切换路由 ID 按首载处理
    const sameDetail =
      !!detail.value && Number(detail.value.id ?? detail.value.sync_tmdb_id) === movieId.value;
    if (!sameDetail) {
      detail.value = null;
    }
    const hadDetail = sameDetail;
    stopDeferredLoads();
    loading.value = true;
    loadError.value = "";
    refreshError.value = "";
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
      loadError.value = "";
      refreshError.value = "";
      if (checkRemoteDiff) {
        await checkRemoteDiffAndPrompt();
      }
      scheduleDeferredLoadsForDetail();
    } catch (error) {
      if (requestSeq !== loadReqSeq) {
        return;
      }
      const message = resolveErrorMessage(error, "请求失败，请重试");
      if (hadDetail) {
        refreshError.value = message;
        loadError.value = "";
      } else {
        detail.value = null;
        loadError.value = message;
        refreshError.value = "";
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
      .map((item) => {
        const value = item && typeof item === "object" ? (item as Record<string, unknown>) : {};
        return {
          field: String(value.field ?? "").trim(),
          diff_type: String(value.diff_type ?? "remote").trim() || "remote",
          local: String(value.local ?? "-"),
          remote: String(value.remote ?? "-"),
        };
      })
      .filter((item) => item.field.length > 0);
  }

  function buildDiffDetailsByFields(
    fields: string[],
    details: AdminCompareFieldDetail[],
    diffType: "remote" | "local_override",
  ): AdminCompareFieldDetail[] {
    const detailMap = new Map(details.filter((item) => item.diff_type === diffType).map((item) => [item.field, item]));
    return fields.map(
      (field) =>
        detailMap.get(field) ?? {
          field,
          diff_type: diffType,
          local: "-",
          remote: "-",
        },
    );
  }

  async function saveMovieChanges() {
    if (!movieId.value) {
      return;
    }
    const runtime = parseOptionalInt(editForm.value.runtime);
    if (editForm.value.runtime.trim() && runtime === undefined) {
      return;
    }
    const voteAverage = parseOptionalFloat(editForm.value.vote_average);
    if (editForm.value.vote_average.trim() && voteAverage === undefined) {
      return;
    }
    const popularity = parseOptionalFloat(editForm.value.popularity);
    if (editForm.value.popularity.trim() && popularity === undefined) {
      return;
    }

    const rawTmdbID = editForm.value.tmdb_id.trim();
    const nextTmdbID = parseOptionalInt(rawTmdbID);
    const tmdbChanged = nextTmdbID !== undefined && nextTmdbID !== movieId.value;
    if (tmdbChanged) {
      if (nextTmdbID === undefined || nextTmdbID <= 0) {
        return;
      }
      const riskConfirm = await askTmdbRiskConfirm(movieId.value, nextTmdbID);
      if (!riskConfirm) {
        return;
      }
    }

    saving.value = true;
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

      const currentId = movieId.value;
      await updateMovie(currentId, payload);
      clearMovieCache(currentId);
      if (tmdbChanged && nextTmdbID !== undefined) {
        clearMovieCache(nextTmdbID);
      }
      saveMessage.value = "已保存到本地数据库";
      isEditing.value = false;
      comparedRemoteId.value = null;
      if (tmdbChanged && nextTmdbID !== undefined) {
        await router.replace(`/movie/${nextTmdbID}`);
        return;
      }
      await loadData({ force: true });
    } catch {
      /* handled by global toast */
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

  return {
    loading,
    loadError,
    refreshError,
    detail,
    castMembers,
    creditsLoading,
    creditsLoaded,
    creditsError,
    isEditing,
    saving,
    deleting,
    uploadingField,
    saveMessage,
    checkingRemoteDiff,
    remoteDiffNotice,
    remoteDiffMessage,
    remoteDiffDecision,
    showRemoteDiffDetails,
    showLocalOverrideDiffDetails,
    tmdbRiskModalVisible,
    tmdbRiskCurrentId,
    tmdbRiskNextId,
    deleteConfirmModalVisible,
    genreOptions,
    genreKeyword,
    filteredGenreOptions,
    editForm,
    movieId,
    currentTmdbId,
    originalTmdbId,
    hasRewrittenTmdbId,
    shouldShowSyncPanel,
    allowedSyncModes,
    goBack,
    personLink,
    updateGenreKeyword,
    toggleRemoteDiffDetails,
    toggleLocalOverrideDiffDetails,
    closeTmdbRiskModal,
    closeDeleteConfirmModal,
    keepLocalData,
    handleSynced,
    deleteCurrentMovie,
    enterEditMode,
    cancelEditMode,
    uploadEditImage,
    saveMovieChanges,
    loadMovieCredits,
    loadData,
    confirmDeleteCurrentMovie,
  };
}
