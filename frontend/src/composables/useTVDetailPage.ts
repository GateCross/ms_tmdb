import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  compareTVRemote,
  deleteTV,
  deleteTVSeasonLocal,
  getTVSeasonLocal,
  saveTVSeasonLocal,
  updateTV,
  updateTVSeasonLocal,
} from "@/api/admin";
import { prefetchMediaDetail } from "@/api/prefetch";
import type { AdminCompareFieldDetail, AdminSyncMode } from "@/api/admin";
import { getTVCredits, getTVDetail, getTVGenreList, getTVSeasonDetail } from "@/api/tv";
import type {
  GenreOption,
  RemoteDiffDecision,
  RemoteDiffNotice,
  TVCastMember,
  TVEditForm,
  TVEpisodeForm,
  TVEpisodeItem,
  TVSeasonDetail,
  TVSeasonForm,
  TVSeasonSummary,
} from "@/components/tv/types";
import { resolveErrorMessage } from "@/utils/errors";
import { normalizeCastMembers, normalizeGenreOptions, normalizeTVEditForm } from "@/utils/mediaNormalizers";
import { scheduleAfterPaint } from "@/utils/schedule";

type CachedSeasonState = {
  payload: Record<string, unknown>;
  saved: boolean;
};

type TVDetail = {
  id?: number;
  sync_tmdb_id?: number;
  name?: string;
  original_name?: string;
  tagline?: string;
  poster_path?: string;
  backdrop_path?: string;
  vote_average?: number;
  first_air_date?: string;
  number_of_seasons?: number;
  number_of_episodes?: number;
  status?: string;
  type?: string;
  overview?: string;
  genres?: GenreOption[];
  seasons?: unknown[];
};

export function useTVDetailPage() {
  const route = useRoute();
  const router = useRouter();

  const loading = ref(false);
  const error = ref("");
  const detail = ref<TVDetail | null>(null);
  const castMembers = ref<TVCastMember[]>([]);
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
  const localDeleteConfirmModalVisible = ref(false);
  const localDeleteConfirmTitle = ref("");
  const localDeleteConfirmMessage = ref("");
  const localDeleteConfirmActionText = ref("确认删除");
  let localDeleteConfirmResolver: ((confirmed: boolean) => void) | null = null;
  const selectedSeasonNumber = ref<number | null>(null);
  const selectedSeasonDetail = ref<TVSeasonDetail | null>(null);
  const selectedSeasonPayload = ref<Record<string, unknown> | null>(null);
  const seasonDetailLoading = ref(false);
  const seasonDetailError = ref("");
  const seasonLocalSaved = ref(false);
  const seasonLocalSaving = ref(false);
  const seasonLocalMessage = ref("");
  const seasonEditorVisible = ref(false);
  const seasonEditorMode = ref<"create" | "edit">("create");
  const seasonFormError = ref("");
  const episodeCreatorVisible = ref(false);
  const episodeFormError = ref("");
  const editingEpisodeNumber = ref<number | null>(null);
  const editingEpisodeInitialForm = ref<TVEpisodeForm | null>(null);
  const episodeEditableFields: Array<keyof TVEpisodeForm> = [
    "episode_number",
    "name",
    "air_date",
    "runtime",
    "vote_average",
    "still_path",
    "overview",
  ];
  const genreOptions = ref<GenreOption[]>([]);
  const genreOptionsLoaded = ref(false);
  const genreKeyword = ref("");
  const seasonDetailCache = new Map<number, CachedSeasonState>();
  let currentDetailId = 0;
  let loadReqSeq = 0;
  let creditsReqSeq = 0;
  let cancelDeferredLoads: (() => void) | null = null;
  let seasonDetailReqSeq = 0;

  const filteredGenreOptions = computed(() => {
    const keyword = genreKeyword.value.trim().toLowerCase();
    if (!keyword) {
      return genreOptions.value;
    }
    return genreOptions.value.filter((genre) => genre.name.toLowerCase().includes(keyword));
  });

  const editForm = ref<TVEditForm>({
    tmdb_id: "",
    name: "",
    original_name: "",
    genre_names: [],
    type: "",
    tagline: "",
    first_air_date: "",
    status: "",
    number_of_seasons: "",
    number_of_episodes: "",
    original_language: "",
    homepage: "",
    poster_path: "",
    backdrop_path: "",
    vote_average: "",
    popularity: "",
    overview: "",
  });

  const seasonForm = ref<TVSeasonForm>({
    season_number: "",
    name: "",
    air_date: "",
    poster_path: "",
    overview: "",
  });

  const episodeForm = ref<TVEpisodeForm>({
    episode_number: "",
    name: "",
    air_date: "",
    runtime: "",
    vote_average: "",
    still_path: "",
    overview: "",
  });

  const tvId = computed(() => Number(route.params.id));
  const currentTmdbId = computed(() => Number(detail.value?.id ?? tvId.value ?? 0));
  const originalTmdbId = computed(() => Number(detail.value?.sync_tmdb_id ?? detail.value?.id ?? tvId.value ?? 0));
  const hasRewrittenTmdbId = computed(() => {
    return originalTmdbId.value > 0 && currentTmdbId.value > 0 && originalTmdbId.value !== currentTmdbId.value;
  });
  const seasonOptions = computed<TVSeasonSummary[]>(() => normalizeSeasonList(detail.value?.seasons));
  const seasonPanelVisible = computed(() => {
    return seasonOptions.value.length > 0 || seasonEditorVisible.value || !!selectedSeasonDetail.value;
  });
  const selectedSeasonEpisodes = computed<TVEpisodeItem[]>(() => selectedSeasonDetail.value?.episodes ?? []);
  const episodeEditChangedCount = computed(() => {
    if (editingEpisodeNumber.value == null || !editingEpisodeInitialForm.value) {
      return 0;
    }
    return episodeEditableFields.reduce((count, field) => {
      return count + (isEpisodeFieldChanged(field) ? 1 : 0);
    }, 0);
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
      query: { tab: "tv" },
    });
  }

  function personLink(personId: number) {
    return {
      path: `/person/${personId}`,
      query: {
        fromType: "tv",
        fromId: String(tvId.value),
      },
    };
  }

  function prefetchPerson(personId: number) {
    prefetchMediaDetail("person", personId);
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
    editForm.value = normalizeTVEditForm(data, tvId.value);
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
      void loadTVCredits();
    });
  }

  function normalizeSeasonList(raw: unknown): TVSeasonSummary[] {
    if (!Array.isArray(raw)) return [];
    return raw
      .map((item) => {
        const value = item && typeof item === "object" ? (item as Record<string, unknown>) : {};
        return {
          id: Number(value.id) || 0,
          season_number: Number(value.season_number) || 0,
          name: String(value.name ?? "").trim() || "未知季",
          poster_path: String(value.poster_path ?? ""),
          episode_count: Number(value.episode_count) || 0,
        };
      })
      .sort((a, b) => a.season_number - b.season_number);
  }

  function normalizeSeasonDetail(raw: unknown, fallbackSeasonNumber: number): TVSeasonDetail {
    const value = raw && typeof raw === "object" ? (raw as Record<string, unknown>) : {};
    const episodes = Array.isArray(value.episodes)
      ? value.episodes.map((item) => {
          const episode = item && typeof item === "object" ? (item as Record<string, unknown>) : {};
          return {
            id: Number(episode.id) || 0,
            episode_number: Number(episode.episode_number) || 0,
            name: String(episode.name ?? "").trim(),
            air_date: String(episode.air_date ?? ""),
            runtime: Number.isFinite(Number(episode.runtime)) ? Number(episode.runtime) : null,
            vote_average: Number.isFinite(Number(episode.vote_average)) ? Number(episode.vote_average) : null,
            overview: String(episode.overview ?? "").trim(),
            still_path: String(episode.still_path ?? ""),
          };
        })
      : [];

    return {
      id: Number(value.id) || 0,
      season_number: Number(value.season_number) || fallbackSeasonNumber,
      name: String(value.name ?? "").trim() || `第 ${fallbackSeasonNumber} 季`,
      air_date: String(value.air_date ?? ""),
      overview: String(value.overview ?? "").trim(),
      poster_path: String(value.poster_path ?? ""),
      episodes,
    };
  }

  function clearSelectedSeasonState() {
    seasonDetailReqSeq++;
    selectedSeasonNumber.value = null;
    selectedSeasonDetail.value = null;
    seasonDetailLoading.value = false;
    seasonDetailError.value = "";
    resetSeasonLocalState();
  }

  function cacheSeasonDetail(seasonNumber: number, payload: Record<string, unknown>, saved: boolean) {
    seasonDetailCache.set(seasonNumber, {
      payload: toPlainRecord(payload),
      saved,
    });
  }

  function applyCachedSeasonDetail(seasonNumber: number, cached: CachedSeasonState) {
    selectedSeasonNumber.value = seasonNumber;
    selectedSeasonPayload.value = toPlainRecord(cached.payload);
    selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, seasonNumber);
    seasonLocalSaved.value = cached.saved;
  }

  function resetSeasonForm(data?: Partial<TVSeasonForm>) {
    seasonForm.value = {
      season_number: data?.season_number ?? "",
      name: data?.name ?? "",
      air_date: data?.air_date ?? "",
      poster_path: data?.poster_path ?? "",
      overview: data?.overview ?? "",
    };
  }

  function resetEpisodeForm(data?: Partial<TVEpisodeForm>) {
    episodeForm.value = {
      episode_number: data?.episode_number ?? "",
      name: data?.name ?? "",
      air_date: data?.air_date ?? "",
      runtime: data?.runtime ?? "",
      vote_average: data?.vote_average ?? "",
      still_path: data?.still_path ?? "",
      overview: data?.overview ?? "",
    };
  }

  function nextSeasonNumberCandidate(): number {
    const regularSeasonNumbers = seasonOptions.value
      .map((item) => item.season_number)
      .filter((seasonNumber) => seasonNumber > 0);
    if (regularSeasonNumbers.length === 0) {
      return 1;
    }
    return Math.max(...regularSeasonNumbers) + 1;
  }

  function nextEpisodeNumberCandidate(): number {
    const episodeNumbers = selectedSeasonEpisodes.value
      .map((item) => item.episode_number)
      .filter((episodeNumber) => episodeNumber > 0);
    if (episodeNumbers.length === 0) {
      return 1;
    }
    return Math.max(...episodeNumbers) + 1;
  }

  function buildSeasonSummary(detailValue: TVSeasonDetail): TVSeasonSummary {
    return {
      id: detailValue.id || 0,
      season_number: detailValue.season_number,
      name: detailValue.name || `第 ${detailValue.season_number} 季`,
      poster_path: detailValue.poster_path || "",
      episode_count: detailValue.episodes.length,
    };
  }

  function syncSeasonSummaryInDetail(detailValue: TVSeasonDetail) {
    if (!detail.value) return;

    const summary = buildSeasonSummary(detailValue);
    const current = normalizeSeasonList(detail.value?.seasons);
    const targetIndex = current.findIndex((item) => item.season_number === summary.season_number);
    if (targetIndex >= 0) {
      current[targetIndex] = summary;
    } else {
      current.push(summary);
    }
    current.sort((a, b) => a.season_number - b.season_number);

    const regularSeasonCount = current.filter((item) => item.season_number > 0).length;
    const totalEpisodeCount = current.reduce((sum, item) => sum + Math.max(0, item.episode_count || 0), 0);
    detail.value = {
      ...detail.value,
      seasons: current,
      number_of_seasons: regularSeasonCount,
      number_of_episodes: totalEpisodeCount,
    };
  }

  function formatEpisodeCode(episodeNumber: number): string {
    return `E${String(episodeNumber || 0).padStart(2, "0")}`;
  }

  function formatEpisodeRuntime(runtime: number | null): string {
    if (!Number.isFinite(runtime) || runtime == null || runtime <= 0) {
      return "-";
    }
    return `${Math.round(runtime)} 分钟`;
  }

  function formatEpisodeRating(voteAverage: number | null): string {
    if (!Number.isFinite(voteAverage) || voteAverage == null || voteAverage <= 0) {
      return "-";
    }
    return voteAverage.toFixed(1);
  }

  function toPlainRecord(raw: unknown): Record<string, unknown> {
    if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
      return {};
    }
    try {
      return JSON.parse(JSON.stringify(raw)) as Record<string, unknown>;
    } catch {
      return {};
    }
  }

  function normalizeEpisodeFormValue(raw: string): string {
    return String(raw ?? "").trim();
  }

  function isEpisodeFieldChanged(field: keyof TVEpisodeForm): boolean {
    if (editingEpisodeNumber.value == null || !editingEpisodeInitialForm.value) {
      return false;
    }
    return (
      normalizeEpisodeFormValue(episodeForm.value[field]) !==
      normalizeEpisodeFormValue(editingEpisodeInitialForm.value[field])
    );
  }

  function episodeEditFieldClass(field: keyof TVEpisodeForm): string {
    if (isEpisodeFieldChanged(field)) {
      return "rounded-lg border border-amber-300 bg-amber-50/80 p-2";
    }
    return "rounded-lg border border-transparent bg-white/65 p-2";
  }

  function resetSeasonLocalState() {
    selectedSeasonPayload.value = null;
    seasonLocalSaved.value = false;
    seasonLocalSaving.value = false;
    seasonLocalMessage.value = "";
    seasonEditorVisible.value = false;
    seasonEditorMode.value = "create";
    seasonFormError.value = "";
    resetSeasonForm();
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    resetEpisodeForm();
    editingEpisodeNumber.value = null;
    editingEpisodeInitialForm.value = null;
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

  function closeLocalDeleteConfirmModal(confirmed: boolean) {
    localDeleteConfirmModalVisible.value = false;
    const resolver = localDeleteConfirmResolver;
    localDeleteConfirmResolver = null;
    localDeleteConfirmTitle.value = "";
    localDeleteConfirmMessage.value = "";
    localDeleteConfirmActionText.value = "确认删除";
    if (resolver) {
      resolver(confirmed);
    }
  }

  function askLocalDeleteConfirm(params: { title: string; message: string; actionText?: string }): Promise<boolean> {
    if (localDeleteConfirmResolver) {
      localDeleteConfirmResolver(false);
      localDeleteConfirmResolver = null;
    }
    localDeleteConfirmTitle.value = params.title;
    localDeleteConfirmMessage.value = params.message;
    localDeleteConfirmActionText.value = params.actionText?.trim() || "确认删除";
    localDeleteConfirmModalVisible.value = true;
    return new Promise((resolve) => {
      localDeleteConfirmResolver = resolve;
    });
  }

  function startEpisodeEdit(ep: TVEpisodeItem) {
    const formData: TVEpisodeForm = {
      episode_number: String(ep.episode_number || ""),
      name: ep.name ?? "",
      air_date: ep.air_date ?? "",
      runtime: ep.runtime == null || !Number.isFinite(ep.runtime) ? "" : String(ep.runtime),
      vote_average: ep.vote_average == null || !Number.isFinite(ep.vote_average) ? "" : String(ep.vote_average),
      still_path: ep.still_path ?? "",
      overview: ep.overview ?? "",
    };
    editingEpisodeNumber.value = ep.episode_number;
    resetEpisodeForm(formData);
    editingEpisodeInitialForm.value = { ...formData };
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    seasonEditorVisible.value = false;
    seasonFormError.value = "";
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
  }

  function cancelEpisodeEdit() {
    editingEpisodeNumber.value = null;
    editingEpisodeInitialForm.value = null;
    resetEpisodeForm();
  }

  function openSeasonCreateEditor() {
    seasonEditorMode.value = "create";
    seasonEditorVisible.value = true;
    seasonFormError.value = "";
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    cancelEpisodeEdit();
    resetSeasonForm({
      season_number: String(nextSeasonNumberCandidate()),
      name: "",
      air_date: "",
      poster_path: "",
      overview: "",
    });
  }

  function openSeasonEditEditor() {
    if (!selectedSeasonDetail.value) {
      seasonDetailError.value = "请先选择要编辑的季";
      return;
    }

    seasonEditorMode.value = "edit";
    seasonEditorVisible.value = true;
    seasonFormError.value = "";
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    cancelEpisodeEdit();
    resetSeasonForm({
      season_number: String(selectedSeasonDetail.value.season_number),
      name: selectedSeasonDetail.value.name ?? "",
      air_date: selectedSeasonDetail.value.air_date ?? "",
      poster_path: selectedSeasonDetail.value.poster_path ?? "",
      overview: selectedSeasonDetail.value.overview ?? "",
    });
  }

  function closeSeasonEditor() {
    seasonEditorVisible.value = false;
    seasonFormError.value = "";
  }

  function openEpisodeCreateEditor() {
    if (!selectedSeasonDetail.value || selectedSeasonNumber.value == null) {
      seasonDetailError.value = "请先选择要编辑的季";
      return;
    }
    if (!seasonLocalSaved.value) {
      seasonDetailError.value = "请先将当前季保存到本地数据库";
      return;
    }

    episodeCreatorVisible.value = true;
    episodeFormError.value = "";
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    seasonEditorVisible.value = false;
    seasonFormError.value = "";
    cancelEpisodeEdit();
    resetEpisodeForm({
      episode_number: String(nextEpisodeNumberCandidate()),
      name: "",
      air_date: "",
      runtime: "",
      vote_average: "",
      still_path: "",
      overview: "",
    });
  }

  function closeEpisodeCreator() {
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    resetEpisodeForm();
  }

  async function saveSeasonEditor() {
    if (!tvId.value) {
      seasonFormError.value = "无效剧集 ID";
      return;
    }

    const seasonNumber = parseOptionalInt(seasonForm.value.season_number);
    if (seasonNumber === undefined || seasonNumber < 0) {
      seasonFormError.value = "季号必须是大于等于 0 的整数";
      return;
    }
    if (
      seasonEditorMode.value === "create" &&
      seasonOptions.value.some((item) => item.season_number === seasonNumber)
    ) {
      seasonFormError.value = "该季号已存在，请直接编辑";
      return;
    }

    const basePayload =
      seasonEditorMode.value === "edit"
        ? toPlainRecord(selectedSeasonPayload.value ?? selectedSeasonDetail.value)
        : { season_number: seasonNumber, episodes: [] };
    const payload: Record<string, unknown> = {
      ...basePayload,
      season_number: seasonNumber,
      name: seasonForm.value.name.trim(),
      air_date: seasonForm.value.air_date.trim(),
      poster_path: seasonForm.value.poster_path.trim(),
      overview: seasonForm.value.overview.trim(),
    };
    if (!Array.isArray(payload.episodes)) {
      payload.episodes = [];
    }

    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    seasonFormError.value = "";
    try {
      const resp = await updateTVSeasonLocal(tvId.value, seasonNumber, payload);
      selectedSeasonNumber.value = seasonNumber;
      selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
      selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, seasonNumber);
      seasonLocalSaved.value = true;
      cacheSeasonDetail(seasonNumber, selectedSeasonPayload.value, true);
      syncSeasonSummaryInDetail(selectedSeasonDetail.value);
      seasonLocalMessage.value =
        seasonEditorMode.value === "create"
          ? `第 ${seasonNumber} 季已新增并保存到本地数据库`
          : `第 ${seasonNumber} 季修改已保存到本地数据库`;
      closeSeasonEditor();
      closeEpisodeCreator();
    } catch (err: unknown) {
      seasonFormError.value = resolveErrorMessage(err, "保存季信息失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function deleteSeasonLocalData() {
    if (!tvId.value || selectedSeasonNumber.value == null) return;
    if (!seasonLocalSaved.value) {
      seasonDetailError.value = "当前季还未保存到本地数据库";
      return;
    }
    const confirmed = await askLocalDeleteConfirm({
      title: "删除本季确认",
      message: `确认删除第 ${selectedSeasonNumber.value} 季的本地数据吗？`,
    });
    if (!confirmed) {
      return;
    }

    const deletingSeasonNumber = selectedSeasonNumber.value;
    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    seasonFormError.value = "";
    episodeFormError.value = "";
    try {
      await deleteTVSeasonLocal(tvId.value, deletingSeasonNumber);
      closeSeasonEditor();
      closeEpisodeCreator();
      cancelEpisodeEdit();
      seasonDetailCache.delete(deletingSeasonNumber);
      await loadSeasonDetail(deletingSeasonNumber, true);
      await loadData({ force: true, checkRemoteDiff: false });
      seasonLocalMessage.value = `第 ${deletingSeasonNumber} 季本地数据已删除`;
    } catch (err: unknown) {
      seasonDetailError.value = resolveErrorMessage(err, "删除本地季失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function saveEpisodeCreate() {
    if (!tvId.value || selectedSeasonNumber.value == null || !selectedSeasonDetail.value) return;
    if (!seasonLocalSaved.value) {
      episodeFormError.value = "请先将当前季保存到本地数据库";
      return;
    }

    const episodeNumber = parseOptionalInt(episodeForm.value.episode_number);
    if (episodeNumber === undefined || episodeNumber <= 0) {
      episodeFormError.value = "集号必须是大于 0 的整数";
      return;
    }
    if (selectedSeasonEpisodes.value.some((item) => item.episode_number === episodeNumber)) {
      episodeFormError.value = "该集号已存在，请使用其他集号";
      return;
    }

    const runtime = parseOptionalInt(episodeForm.value.runtime);
    if (episodeForm.value.runtime.trim() && runtime === undefined) {
      episodeFormError.value = "时长必须是数字";
      return;
    }
    const voteAverage = parseOptionalFloat(episodeForm.value.vote_average);
    if (episodeForm.value.vote_average.trim() && voteAverage === undefined) {
      episodeFormError.value = "评分必须是数字";
      return;
    }

    const basePayload = toPlainRecord(selectedSeasonPayload.value ?? selectedSeasonDetail.value);
    const updatedEpisodes = [
      ...selectedSeasonEpisodes.value.map((ep) => ({
        ...ep,
        id: ep.id || 0,
        episode_number: ep.episode_number,
      })),
      {
        id: 0,
        episode_number: episodeNumber,
        name: episodeForm.value.name.trim(),
        air_date: episodeForm.value.air_date.trim(),
        runtime: runtime ?? null,
        vote_average: voteAverage ?? null,
        overview: episodeForm.value.overview.trim(),
        still_path: episodeForm.value.still_path.trim(),
      },
    ].sort((a, b) => a.episode_number - b.episode_number);

    const payload: Record<string, unknown> = {
      ...basePayload,
      season_number: selectedSeasonNumber.value,
      episodes: updatedEpisodes,
    };

    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    episodeFormError.value = "";
    try {
      const resp = await updateTVSeasonLocal(tvId.value, selectedSeasonNumber.value, payload);
      selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
      selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, selectedSeasonNumber.value);
      seasonLocalSaved.value = true;
      cacheSeasonDetail(selectedSeasonNumber.value, selectedSeasonPayload.value, true);
      syncSeasonSummaryInDetail(selectedSeasonDetail.value);
      seasonLocalMessage.value = `第 ${episodeNumber} 集已新增到本地数据库`;
      closeEpisodeCreator();
    } catch (err: unknown) {
      episodeFormError.value = resolveErrorMessage(err, "新增本集失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function saveSeasonToLocalFromTMDB() {
    if (!tvId.value || selectedSeasonNumber.value == null) return;
    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    try {
      const resp = await saveTVSeasonLocal(tvId.value, selectedSeasonNumber.value);
      selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
      selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, selectedSeasonNumber.value);
      seasonLocalSaved.value = true;
      cacheSeasonDetail(selectedSeasonNumber.value, selectedSeasonPayload.value, true);
      syncSeasonSummaryInDetail(selectedSeasonDetail.value);
      cancelEpisodeEdit();
      closeEpisodeCreator();
      seasonLocalMessage.value = "当前季明细已保存到本地数据库";
    } catch (err: unknown) {
      seasonDetailError.value = resolveErrorMessage(err, "保存季明细失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function saveEpisodeEdit() {
    if (!tvId.value || selectedSeasonNumber.value == null || !selectedSeasonDetail.value) return;
    if (editingEpisodeNumber.value == null) {
      episodeFormError.value = "请先选择要编辑的集";
      return;
    }
    if (!seasonLocalSaved.value) {
      episodeFormError.value = "请先将当前季保存到本地数据库";
      return;
    }

    const targetEpisodeNumber = editingEpisodeNumber.value;
    const episodeNumber = parseOptionalInt(episodeForm.value.episode_number);
    if (episodeNumber === undefined || episodeNumber <= 0) {
      episodeFormError.value = "集号必须是大于 0 的整数";
      return;
    }
    if (
      episodeNumber !== targetEpisodeNumber &&
      selectedSeasonEpisodes.value.some((item) => item.episode_number === episodeNumber)
    ) {
      episodeFormError.value = "该集号已存在，请使用其他集号";
      return;
    }

    const runtime = parseOptionalInt(episodeForm.value.runtime);
    if (episodeForm.value.runtime.trim() && runtime === undefined) {
      episodeFormError.value = "时长必须是数字";
      return;
    }
    const voteAverage = parseOptionalFloat(episodeForm.value.vote_average);
    if (episodeForm.value.vote_average.trim() && voteAverage === undefined) {
      episodeFormError.value = "评分必须是数字";
      return;
    }

    const basePayload = toPlainRecord(selectedSeasonPayload.value ?? selectedSeasonDetail.value);
    let updated = false;
    const updatedEpisodes = selectedSeasonEpisodes.value.map((ep, idx) => {
      if (ep.episode_number !== targetEpisodeNumber) {
        return {
          ...ep,
          id: ep.id || idx + 1,
          episode_number: ep.episode_number || idx + 1,
        };
      }
      updated = true;
      return {
        ...ep,
        id: ep.id || idx + 1,
        episode_number: episodeNumber,
        name: episodeForm.value.name.trim(),
        air_date: episodeForm.value.air_date.trim(),
        runtime: runtime ?? null,
        vote_average: voteAverage ?? null,
        still_path: episodeForm.value.still_path.trim(),
        overview: episodeForm.value.overview.trim(),
      };
    });
    if (!updated) {
      episodeFormError.value = "未找到要编辑的目标集";
      return;
    }

    const payload: Record<string, unknown> = {
      ...basePayload,
      season_number: selectedSeasonNumber.value,
      episodes: updatedEpisodes,
    };

    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    episodeFormError.value = "";
    try {
      const resp = await updateTVSeasonLocal(tvId.value, selectedSeasonNumber.value, payload);
      selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
      selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, selectedSeasonNumber.value);
      seasonLocalSaved.value = true;
      cacheSeasonDetail(selectedSeasonNumber.value, selectedSeasonPayload.value, true);
      syncSeasonSummaryInDetail(selectedSeasonDetail.value);
      seasonLocalMessage.value = `第 ${episodeNumber} 集本地修改已保存`;
      cancelEpisodeEdit();
      closeEpisodeCreator();
    } catch (err: unknown) {
      episodeFormError.value = resolveErrorMessage(err, "保存本集修改失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function deleteEpisode(targetEpisodeNumber: number) {
    if (!tvId.value || selectedSeasonNumber.value == null || !selectedSeasonDetail.value) return;
    if (!seasonLocalSaved.value) {
      seasonDetailError.value = "请先将当前季保存到本地数据库";
      return;
    }
    const confirmed = await askLocalDeleteConfirm({
      title: "删除本集确认",
      message: `确认删除第 ${targetEpisodeNumber} 集吗？`,
    });
    if (!confirmed) {
      return;
    }

    const basePayload = toPlainRecord(selectedSeasonPayload.value ?? selectedSeasonDetail.value);
    const updatedEpisodes = selectedSeasonEpisodes.value
      .filter((ep) => ep.episode_number !== targetEpisodeNumber)
      .map((ep, idx) => ({
        ...ep,
        id: ep.id || idx + 1,
        episode_number: ep.episode_number || idx + 1,
      }));

    if (updatedEpisodes.length === selectedSeasonEpisodes.value.length) {
      seasonDetailError.value = "未找到要删除的目标集";
      return;
    }

    const payload: Record<string, unknown> = {
      ...basePayload,
      season_number: selectedSeasonNumber.value,
      episodes: updatedEpisodes,
    };

    seasonLocalSaving.value = true;
    seasonLocalMessage.value = "";
    seasonDetailError.value = "";
    episodeFormError.value = "";
    try {
      const resp = await updateTVSeasonLocal(tvId.value, selectedSeasonNumber.value, payload);
      selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
      selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, selectedSeasonNumber.value);
      seasonLocalSaved.value = true;
      cacheSeasonDetail(selectedSeasonNumber.value, selectedSeasonPayload.value, true);
      syncSeasonSummaryInDetail(selectedSeasonDetail.value);
      closeEpisodeCreator();
      if (editingEpisodeNumber.value === targetEpisodeNumber) {
        cancelEpisodeEdit();
      }
      seasonLocalMessage.value = `第 ${targetEpisodeNumber} 集已从本地数据库删除`;
    } catch (err: unknown) {
      seasonDetailError.value = resolveErrorMessage(err, "删除本集失败");
    } finally {
      seasonLocalSaving.value = false;
    }
  }

  async function loadSeasonDetail(seasonNumber: number, force = false) {
    if (!tvId.value) return;
    selectedSeasonNumber.value = seasonNumber;
    seasonDetailLoading.value = true;
    seasonDetailError.value = "";
    seasonLocalSaved.value = false;
    seasonLocalMessage.value = "";
    seasonEditorVisible.value = false;
    seasonFormError.value = "";
    episodeCreatorVisible.value = false;
    episodeFormError.value = "";
    cancelEpisodeEdit();
    const requestSeq = ++seasonDetailReqSeq;
    try {
      if (!force) {
        const cached = seasonDetailCache.get(seasonNumber);
        if (cached) {
          applyCachedSeasonDetail(seasonNumber, cached);
          return;
        }
      }

      try {
        const localResp = await getTVSeasonLocal(tvId.value, seasonNumber);
        if (requestSeq !== seasonDetailReqSeq) {
          return;
        }
        if (localResp.data?.saved && localResp.data?.data) {
          cacheSeasonDetail(seasonNumber, toPlainRecord(localResp.data.data), true);
          applyCachedSeasonDetail(seasonNumber, seasonDetailCache.get(seasonNumber)!);
          return;
        }
      } catch {
        // 本地查询失败时，降级走 TMDB 季详情接口
      }

      const resp = await getTVSeasonDetail(tvId.value, seasonNumber);
      if (requestSeq !== seasonDetailReqSeq) {
        return;
      }
      cacheSeasonDetail(seasonNumber, toPlainRecord(resp.data), false);
      applyCachedSeasonDetail(seasonNumber, seasonDetailCache.get(seasonNumber)!);
    } catch (err: unknown) {
      if (requestSeq !== seasonDetailReqSeq) {
        return;
      }
      selectedSeasonPayload.value = null;
      selectedSeasonDetail.value = null;
      seasonLocalSaved.value = false;
      seasonDetailError.value = resolveErrorMessage(err, "加载分集明细失败");
    } finally {
      if (requestSeq === seasonDetailReqSeq) {
        seasonDetailLoading.value = false;
      }
    }
  }

  function selectSeason(seasonNumber: number) {
    if (seasonNumber === selectedSeasonNumber.value && selectedSeasonDetail.value) {
      return;
    }
    void loadSeasonDetail(seasonNumber);
  }

  async function loadTVCredits(force = false) {
    if (!tvId.value || creditsLoading.value || (creditsLoaded.value && !force)) {
      return;
    }

    const requestSeq = ++creditsReqSeq;
    const targetId = tvId.value;
    creditsLoading.value = true;
    creditsError.value = "";
    try {
      const resp = await getTVCredits(targetId, "zh-CN", { force });
      if (requestSeq !== creditsReqSeq || targetId !== tvId.value) {
        return;
      }
      castMembers.value = normalizeCastMembers(resp.data);
      creditsLoaded.value = true;
    } catch (err: unknown) {
      if (requestSeq !== creditsReqSeq || targetId !== tvId.value) {
        return;
      }
      creditsError.value = resolveErrorMessage(err, "加载演员失败");
    } finally {
      if (requestSeq === creditsReqSeq) {
        creditsLoading.value = false;
      }
    }
  }

  async function loadGenreOptions(force = false) {
    if (!force && genreOptionsLoaded.value) {
      return;
    }
    try {
      const resp = await getTVGenreList();
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

  async function deleteCurrentTV() {
    if (!tvId.value) {
      deleteError.value = "无效剧集 ID";
      return;
    }
    deleteConfirmModalVisible.value = true;
  }

  function closeDeleteConfirmModal() {
    deleteConfirmModalVisible.value = false;
  }

  async function confirmDeleteCurrentTV() {
    if (!tvId.value) {
      deleteError.value = "无效剧集 ID";
      deleteConfirmModalVisible.value = false;
      return;
    }

    deleting.value = true;
    deleteError.value = "";
    try {
      deleteConfirmModalVisible.value = false;
      await deleteTV(tvId.value);
      await router.push({
        path: "/library",
        query: { tab: "tv" },
      });
    } catch (err: unknown) {
      deleteError.value = resolveErrorMessage(err, "删除失败");
    } finally {
      deleting.value = false;
    }
  }

  async function checkRemoteDiffAndPrompt(force = false) {
    if (!tvId.value || checkingRemoteDiff.value || (!force && comparedRemoteId.value === tvId.value)) {
      return;
    }
    if (tvId.value < 0) {
      remoteDiffNotice.value = null;
      showRemoteDiffDetails.value = false;
      showLocalOverrideDiffDetails.value = false;
      remoteDiffDecision.value = "keep_local";
      remoteDiffError.value = "";
      remoteDiffMessage.value = "本地新建条目不参与 TMDB 远程差异检测";
      comparedRemoteId.value = tvId.value;
      return;
    }
    checkingRemoteDiff.value = true;
    remoteDiffError.value = "";
    try {
      const resp = await compareTVRemote(tvId.value);
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
        comparedRemoteId.value = tvId.value;
        return;
      }

      const remoteFieldPreview = remoteFields.slice(0, 6).join("、");
      const remoteSummary =
        remoteFields.length === 0
          ? "无"
          : remoteFields.length > 6
            ? `${remoteFieldPreview} 等 ${remoteFields.length} 项`
            : `${remoteFieldPreview}（共 ${remoteFields.length} 项）`;
      const localOverridePreview = localOverrideFields.slice(0, 6).join("、");
      const localOverrideSummary =
        localOverrideFields.length === 0
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
      comparedRemoteId.value = tvId.value;
    } catch (err: unknown) {
      remoteDiffError.value = resolveErrorMessage(err, "远程差异检测失败");
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
    if (!tvId.value) {
      error.value = "无效剧集 ID";
      currentDetailId = 0;
      seasonDetailCache.clear();
      clearSelectedSeasonState();
      return;
    }
    const targetId = tvId.value;
    if (targetId !== currentDetailId) {
      currentDetailId = targetId;
      seasonDetailCache.clear();
      clearSelectedSeasonState();
    }
    const requestSeq = ++loadReqSeq;
    stopDeferredLoads();
    loading.value = true;
    error.value = "";
    resetRemoteDiffState();
    resetCreditsState();
    try {
      const resp = await getTVDetail(targetId, "zh-CN", "", { force });
      if (requestSeq !== loadReqSeq) {
        return;
      }
      detail.value = resp.data;
      resetEditForm(resp.data);
      genreOptions.value = normalizeGenreOptions(resp.data?.genres);
      genreOptionsLoaded.value = false;
      genreKeyword.value = "";
      isEditing.value = false;
      const seasons = normalizeSeasonList(resp.data?.seasons);
      if (selectedSeasonNumber.value != null) {
        if (seasons.some((item) => item.season_number === selectedSeasonNumber.value)) {
          const cached = seasonDetailCache.get(selectedSeasonNumber.value);
          if (cached) {
            applyCachedSeasonDetail(selectedSeasonNumber.value, cached);
          }
        } else {
          clearSelectedSeasonState();
        }
      }
      if (checkRemoteDiff) {
        await checkRemoteDiffAndPrompt();
      }
      scheduleDeferredLoadsForDetail();
    } catch (err: unknown) {
      if (requestSeq === loadReqSeq) {
        error.value = resolveErrorMessage(err, "加载失败");
        clearSelectedSeasonState();
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

  async function saveTVChanges() {
    if (!tvId.value) {
      saveError.value = "无效剧集 ID";
      return;
    }
    const seasons = parseOptionalInt(editForm.value.number_of_seasons);
    if (editForm.value.number_of_seasons.trim() && seasons === undefined) {
      saveError.value = "季数必须是数字";
      return;
    }
    const episodes = parseOptionalInt(editForm.value.number_of_episodes);
    if (editForm.value.number_of_episodes.trim() && episodes === undefined) {
      saveError.value = "集数必须是数字";
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
    const tmdbChanged = nextTmdbID !== undefined && nextTmdbID !== tvId.value;
    if (tmdbChanged) {
      if (nextTmdbID === undefined || nextTmdbID <= 0) {
        saveError.value = "TMDB ID 必须是大于 0 的整数";
        return;
      }
      const riskConfirm = await askTmdbRiskConfirm(tvId.value, nextTmdbID);
      if (!riskConfirm) {
        return;
      }
    }

    saving.value = true;
    saveError.value = "";
    saveMessage.value = "";
    try {
      const payload: Record<string, unknown> = {
        name: editForm.value.name.trim(),
        original_name: editForm.value.original_name.trim(),
        genre_names: editForm.value.genre_names,
        type: editForm.value.type.trim(),
        tagline: editForm.value.tagline.trim(),
        first_air_date: editForm.value.first_air_date.trim(),
        status: editForm.value.status.trim(),
        original_language: editForm.value.original_language.trim(),
        homepage: editForm.value.homepage.trim(),
        poster_path: editForm.value.poster_path.trim(),
        backdrop_path: editForm.value.backdrop_path.trim(),
        overview: editForm.value.overview.trim(),
      };
      if (seasons !== undefined) {
        payload.number_of_seasons = seasons;
      }
      if (episodes !== undefined) {
        payload.number_of_episodes = episodes;
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

      await updateTV(tvId.value, payload);
      saveMessage.value = "已保存到本地数据库";
      isEditing.value = false;
      comparedRemoteId.value = null;
      if (tmdbChanged && nextTmdbID !== undefined) {
        await router.replace(`/tv/${nextTmdbID}`);
        return;
      }
      await loadData({ force: true });
    } catch (err: unknown) {
      saveError.value = resolveErrorMessage(err, "保存失败");
    } finally {
      saving.value = false;
    }
  }

  onMounted(loadData);
  watch(tvId, () => {
    void loadData();
  });

  onBeforeUnmount(() => {
    loadReqSeq++;
    creditsReqSeq++;
    stopDeferredLoads();
    seasonDetailReqSeq++;
  });

  return {
    loading,
    error,
    detail,
    castMembers,
    creditsLoading,
    creditsLoaded,
    creditsError,
    isEditing,
    saving,
    deleting,
    saveError,
    saveMessage,
    deleteError,
    checkingRemoteDiff,
    remoteDiffNotice,
    remoteDiffMessage,
    remoteDiffError,
    remoteDiffDecision,
    showRemoteDiffDetails,
    showLocalOverrideDiffDetails,
    tmdbRiskModalVisible,
    tmdbRiskCurrentId,
    tmdbRiskNextId,
    deleteConfirmModalVisible,
    localDeleteConfirmModalVisible,
    localDeleteConfirmTitle,
    localDeleteConfirmMessage,
    localDeleteConfirmActionText,
    selectedSeasonNumber,
    selectedSeasonDetail,
    seasonDetailLoading,
    seasonDetailError,
    seasonLocalSaved,
    seasonLocalSaving,
    seasonLocalMessage,
    seasonEditorVisible,
    seasonEditorMode,
    seasonFormError,
    episodeCreatorVisible,
    episodeFormError,
    editingEpisodeNumber,
    genreOptions,
    genreKeyword,
    filteredGenreOptions,
    editForm,
    seasonForm,
    episodeForm,
    tvId,
    currentTmdbId,
    originalTmdbId,
    hasRewrittenTmdbId,
    seasonOptions,
    seasonPanelVisible,
    selectedSeasonEpisodes,
    episodeEditChangedCount,
    shouldShowSyncPanel,
    allowedSyncModes,
    goBack,
    personLink,
    prefetchPerson,
    updateGenreKeyword,
    toggleRemoteDiffDetails,
    toggleLocalOverrideDiffDetails,
    closeTmdbRiskModal,
    closeDeleteConfirmModal,
    closeLocalDeleteConfirmModal,
    checkRemoteDiffAndPrompt,
    keepLocalData,
    handleSynced,
    deleteCurrentTV,
    enterEditMode,
    cancelEditMode,
    saveTVChanges,
    episodeEditFieldClass,
    formatEpisodeCode,
    formatEpisodeRuntime,
    formatEpisodeRating,
    openSeasonCreateEditor,
    selectSeason,
    openSeasonEditEditor,
    openEpisodeCreateEditor,
    saveSeasonToLocalFromTMDB,
    deleteSeasonLocalData,
    saveSeasonEditor,
    closeSeasonEditor,
    saveEpisodeCreate,
    closeEpisodeCreator,
    startEpisodeEdit,
    saveEpisodeEdit,
    cancelEpisodeEdit,
    deleteEpisode,
    loadTVCredits,
    confirmDeleteCurrentTV,
  };
}
