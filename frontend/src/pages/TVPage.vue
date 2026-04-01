<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { compareTVRemote, deleteTV, getTVSeasonLocal, saveTVSeasonLocal, updateTV, updateTVSeasonLocal } from "@/api/admin";
import type { AdminCompareFieldDetail, AdminSyncMode } from "@/api/admin";
import DetailSyncPanel from "@/components/DetailSyncPanel.vue";
import GlassSelect from "@/components/GlassSelect.vue";
import { getTVDetail, getTVGenreList, getTVSeasonDetail } from "@/api/tv";
import { tmdbImg } from "@/api/tmdb";
import { formatStatusLabel, formatTvTypeLabel, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";

type GenreOption = {
  id: number;
  name: string;
};

type TVEditForm = {
  tmdb_id: string;
  name: string;
  original_name: string;
  genre_names: string[];
  type: string;
  tagline: string;
  first_air_date: string;
  status: string;
  number_of_seasons: string;
  number_of_episodes: string;
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

type TVSeasonSummary = {
  id: number;
  season_number: number;
  name: string;
  poster_path: string;
  episode_count: number;
};

type TVEpisodeItem = {
  id: number;
  episode_number: number;
  name: string;
  air_date: string;
  runtime: number | null;
  vote_average: number | null;
  overview: string;
  still_path: string;
};

type TVSeasonDetail = {
  id: number;
  season_number: number;
  name: string;
  air_date: string;
  overview: string;
  episodes: TVEpisodeItem[];
};

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const error = ref("");
const detail = ref<any>(null);
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
const selectedSeasonNumber = ref<number | null>(null);
const selectedSeasonDetail = ref<TVSeasonDetail | null>(null);
const selectedSeasonPayload = ref<Record<string, unknown> | null>(null);
const seasonDetailLoading = ref(false);
const seasonDetailError = ref("");
const seasonLocalSaved = ref(false);
const seasonLocalSaving = ref(false);
const seasonLocalMessage = ref("");
const editingEpisodeNumber = ref<number | null>(null);
const editingEpisodeName = ref("");
const editingEpisodeOverview = ref("");
const genreOptions = ref<GenreOption[]>([]);
const genreKeyword = ref("");
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

const tvId = computed(() => Number(route.params.id));
const currentTmdbId = computed(() => Number(detail.value?.id ?? tvId.value ?? 0));
const originalTmdbId = computed(() => Number(detail.value?.sync_tmdb_id ?? detail.value?.id ?? tvId.value ?? 0));
const hasRewrittenTmdbId = computed(() => {
  return originalTmdbId.value > 0 && currentTmdbId.value > 0 && originalTmdbId.value !== currentTmdbId.value;
});
const seasonOptions = computed<TVSeasonSummary[]>(() => normalizeSeasonList(detail.value?.seasons));
const selectedSeasonEpisodes = computed<TVEpisodeItem[]>(() => selectedSeasonDetail.value?.episodes ?? []);
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

function resetEditForm(data: any) {
  editForm.value = {
    tmdb_id: data?.id != null ? String(data.id) : String(tvId.value || ""),
    name: data?.name ?? "",
    original_name: data?.original_name ?? "",
    genre_names: Array.isArray(data?.genres) ? data.genres.map((g: any) => String(g?.name ?? "").trim()).filter(Boolean) : [],
    type: data?.type ?? "",
    tagline: data?.tagline ?? "",
    first_air_date: data?.first_air_date ?? "",
    status: data?.status ?? "",
    number_of_seasons: data?.number_of_seasons != null ? String(data.number_of_seasons) : "",
    number_of_episodes: data?.number_of_episodes != null ? String(data.number_of_episodes) : "",
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

function normalizeSeasonList(raw: unknown): TVSeasonSummary[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item: any) => ({
      id: Number(item?.id) || 0,
      season_number: Number(item?.season_number) || 0,
      name: String(item?.name ?? "").trim() || "未知季",
      poster_path: String(item?.poster_path ?? ""),
      episode_count: Number(item?.episode_count) || 0,
    }))
    .sort((a, b) => a.season_number - b.season_number);
}

function normalizeSeasonDetail(raw: any, fallbackSeasonNumber: number): TVSeasonDetail {
  const episodes = Array.isArray(raw?.episodes)
    ? raw.episodes.map((item: any) => ({
      id: Number(item?.id) || 0,
      episode_number: Number(item?.episode_number) || 0,
      name: String(item?.name ?? "").trim(),
      air_date: String(item?.air_date ?? ""),
      runtime: Number.isFinite(Number(item?.runtime)) ? Number(item.runtime) : null,
      vote_average: Number.isFinite(Number(item?.vote_average)) ? Number(item.vote_average) : null,
      overview: String(item?.overview ?? "").trim(),
      still_path: String(item?.still_path ?? ""),
    }))
    : [];

  return {
    id: Number(raw?.id) || 0,
    season_number: Number(raw?.season_number) || fallbackSeasonNumber,
    name: String(raw?.name ?? "").trim() || `第 ${fallbackSeasonNumber} 季`,
    air_date: String(raw?.air_date ?? ""),
    overview: String(raw?.overview ?? "").trim(),
    episodes,
  };
}

function pickDefaultSeasonNumber(seasons: TVSeasonSummary[]): number | null {
  if (seasons.length === 0) return null;
  const normalSeason = seasons.find((item) => item.season_number > 0);
  return normalSeason?.season_number ?? seasons[0].season_number;
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

function resetSeasonLocalState() {
  selectedSeasonPayload.value = null;
  seasonLocalSaved.value = false;
  seasonLocalSaving.value = false;
  seasonLocalMessage.value = "";
  editingEpisodeNumber.value = null;
  editingEpisodeName.value = "";
  editingEpisodeOverview.value = "";
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

function startEpisodeEdit(ep: TVEpisodeItem) {
  editingEpisodeNumber.value = ep.episode_number;
  editingEpisodeName.value = ep.name ?? "";
  editingEpisodeOverview.value = ep.overview ?? "";
  seasonLocalMessage.value = "";
  seasonDetailError.value = "";
}

function cancelEpisodeEdit() {
  editingEpisodeNumber.value = null;
  editingEpisodeName.value = "";
  editingEpisodeOverview.value = "";
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
    cancelEpisodeEdit();
    seasonLocalMessage.value = "当前季明细已保存到本地数据库";
  } catch (err: any) {
    seasonDetailError.value = err.message ?? "保存季明细失败";
  } finally {
    seasonLocalSaving.value = false;
  }
}

async function saveEpisodeEdit() {
  if (!tvId.value || selectedSeasonNumber.value == null || !selectedSeasonDetail.value) return;
  if (editingEpisodeNumber.value == null) {
    seasonDetailError.value = "请先选择要编辑的集";
    return;
  }

  const basePayload = toPlainRecord(selectedSeasonPayload.value ?? selectedSeasonDetail.value);
  const targetEpisodeNumber = editingEpisodeNumber.value;
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
      episode_number: ep.episode_number || idx + 1,
      name: editingEpisodeName.value.trim(),
      overview: editingEpisodeOverview.value.trim(),
    };
  });
  if (!updated) {
    seasonDetailError.value = "未找到要编辑的目标集";
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
  try {
    const resp = await updateTVSeasonLocal(tvId.value, selectedSeasonNumber.value, payload);
    selectedSeasonPayload.value = toPlainRecord(resp.data?.data);
    selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, selectedSeasonNumber.value);
    seasonLocalSaved.value = true;
    seasonLocalMessage.value = `第 ${targetEpisodeNumber} 集本地修改已保存`;
    cancelEpisodeEdit();
  } catch (err: any) {
    seasonDetailError.value = err.message ?? "保存本集修改失败";
  } finally {
    seasonLocalSaving.value = false;
  }
}

let seasonDetailReqSeq = 0;

async function loadSeasonDetail(seasonNumber: number) {
  if (!tvId.value) return;
  selectedSeasonNumber.value = seasonNumber;
  seasonDetailLoading.value = true;
  seasonDetailError.value = "";
  seasonLocalSaved.value = false;
  seasonLocalMessage.value = "";
  cancelEpisodeEdit();
  const requestSeq = ++seasonDetailReqSeq;
  try {
    try {
      const localResp = await getTVSeasonLocal(tvId.value, seasonNumber);
      if (requestSeq !== seasonDetailReqSeq) {
        return;
      }
      if (localResp.data?.saved && localResp.data?.data) {
        selectedSeasonPayload.value = toPlainRecord(localResp.data.data);
        selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, seasonNumber);
        seasonLocalSaved.value = true;
        return;
      }
    } catch {
      // 本地查询失败时，降级走 TMDB 季详情接口
    }

    const resp = await getTVSeasonDetail(tvId.value, seasonNumber);
    if (requestSeq !== seasonDetailReqSeq) {
      return;
    }
    selectedSeasonPayload.value = toPlainRecord(resp.data);
    selectedSeasonDetail.value = normalizeSeasonDetail(selectedSeasonPayload.value, seasonNumber);
    seasonLocalSaved.value = false;
  } catch (err: any) {
    if (requestSeq !== seasonDetailReqSeq) {
      return;
    }
    selectedSeasonPayload.value = null;
    selectedSeasonDetail.value = null;
    seasonLocalSaved.value = false;
    seasonDetailError.value = err.message ?? "加载分集明细失败";
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

async function loadGenreOptions() {
  try {
    const resp = await getTVGenreList();
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
  } catch (err: any) {
    deleteError.value = err.message ?? "删除失败";
  } finally {
    deleting.value = false;
  }
}

async function checkRemoteDiffAndPrompt() {
  if (!tvId.value || checkingRemoteDiff.value || comparedRemoteId.value === tvId.value) {
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
    const localOverrideFields = Array.isArray(resp.data?.local_override_diff_fields) ? resp.data.local_override_diff_fields : [];
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
    comparedRemoteId.value = tvId.value;
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
  void loadData();
}

async function loadData(options: { checkRemoteDiff?: boolean } = {}) {
  const { checkRemoteDiff = true } = options;
  if (!tvId.value) {
    error.value = "无效剧集 ID";
    seasonDetailReqSeq++;
    selectedSeasonNumber.value = null;
    selectedSeasonDetail.value = null;
    seasonDetailLoading.value = false;
    seasonDetailError.value = "";
    resetSeasonLocalState();
    return;
  }
  loading.value = true;
  error.value = "";
  remoteDiffError.value = "";
  try {
    const resp = await getTVDetail(tvId.value);
    detail.value = resp.data;
    resetEditForm(resp.data);
    await loadGenreOptions();
    genreKeyword.value = "";
    isEditing.value = false;
    const seasons = normalizeSeasonList(resp.data?.seasons);
    const targetSeasonNumber = seasons.some((item) => item.season_number === selectedSeasonNumber.value)
      ? selectedSeasonNumber.value
      : pickDefaultSeasonNumber(seasons);
    if (targetSeasonNumber != null) {
      await loadSeasonDetail(targetSeasonNumber);
    } else {
      seasonDetailReqSeq++;
      selectedSeasonNumber.value = null;
      selectedSeasonDetail.value = null;
      seasonDetailError.value = "";
      seasonDetailLoading.value = false;
      resetSeasonLocalState();
    }
    if (checkRemoteDiff) {
      await checkRemoteDiffAndPrompt();
    }
  } catch (err: any) {
    error.value = err.message ?? "加载失败";
    seasonDetailReqSeq++;
    selectedSeasonNumber.value = null;
    selectedSeasonDetail.value = null;
    seasonDetailError.value = "";
    seasonDetailLoading.value = false;
    resetSeasonLocalState();
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
    await loadData();
  } catch (err: any) {
    saveError.value = err.message ?? "保存失败";
  } finally {
    saving.value = false;
  }
}

onMounted(loadData);
watch(tvId, () => {
  void loadData();
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
        :alt="detail.name || detail.original_name"
        class="hero-banner-media"
      />
      <div class="absolute left-4 top-4 z-10">
        <button
          class="rounded-lg border border-white/40 bg-black/40 px-3 py-1.5 text-xs text-white backdrop-blur hover:bg-black/55"
          @click="goBack"
        >
          返回上一页
        </button>
      </div>
      <div class="hero-overlay">
        <h1 class="text-2xl font-bold text-white md:text-3xl">{{ detail.name || detail.original_name }}</h1>
        <p class="mt-1 text-sm text-white/70">{{ detail.tagline }}</p>
      </div>
    </section>

    <section class="card mt-4">
      <div class="detail-layout">
        <div class="detail-poster">
          <img
            :src="tmdbImg(detail.poster_path, 'w342')"
            :alt="detail.name"
            class="w-full rounded-xl shadow-soft"
          />
        </div>

        <div class="detail-info">
          <h2 class="text-xl font-bold">{{ detail.name }}</h2>
          <p v-if="detail.original_name !== detail.name" class="text-sm text-black/55">
            {{ detail.original_name }}
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
            <span class="badge">📅 {{ detail.first_air_date ?? "-" }}</span>
            <span v-if="detail.number_of_seasons" class="badge">
              {{ detail.number_of_seasons }} 季 · {{ detail.number_of_episodes }} 集
            </span>
            <span class="badge">{{ formatStatusLabel(detail.status) }}</span>
            <span v-if="detail.type" class="badge">{{ formatTvTypeLabel(detail.type) }}</span>
          </div>

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
                检测到远程剧集数据与本地不一致
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
              media-type="tv"
              :target-id="tvId"
              :allowed-modes="allowedSyncModes"
              :preset-changed-fields="remoteDiffNotice?.localOverrideFields ?? []"
              :embedded="true"
              @synced="handleSynced"
            />

            <p v-if="!checkingRemoteDiff && !remoteDiffNotice && remoteDiffMessage" class="text-xs text-green-700">
              {{ remoteDiffMessage }}
            </p>
            <p v-if="remoteDiffError" class="mt-1 text-xs text-red-600">
              {{ remoteDiffError }}
            </p>
          </div>

          <div class="panel-glass mt-6 rounded-xl p-4">
            <div class="flex items-center justify-between gap-3">
              <h3 class="text-sm font-semibold">本地信息编辑</h3>
              <div class="flex items-center gap-2">
                <button
                  class="btn-danger-soft-xs disabled:opacity-60"
                  :disabled="deleting || saving"
                  @click="deleteCurrentTV"
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
                    placeholder="例如：1399"
                  />
                  <p class="mt-1 text-[11px] text-amber-700">
                    高风险：改动后，后续同步仍使用旧 TMDB ID 拉取；对外返回与访问使用新 TMDB ID。
                  </p>
                </label>
                <label class="text-xs text-black/60">
                  剧名
                  <input
                    v-model="editForm.name"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="剧集标题"
                  />
                </label>
                <label class="text-xs text-black/60">
                  原始剧名
                  <input
                    v-model="editForm.original_name"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Original Name"
                  />
                </label>
                <label class="text-xs text-black/60 md:col-span-2">
                  类型（多选）
                  <div class="mt-1 flex flex-wrap gap-2 rounded-lg border border-white/70 bg-white/55 p-2 backdrop-blur">
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
                  首播日期
                  <input
                    v-model="editForm.first_air_date"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="YYYY-MM-DD"
                  />
                </label>
                <label class="text-xs text-black/60">
                  状态
                  <GlassSelect v-model="editForm.status" :options="tvStatusOptions" class="mt-1 w-full" />
                </label>
                <label class="text-xs text-black/60">
                  剧集类型
                  <GlassSelect v-model="editForm.type" :options="tvTypeOptions" class="mt-1 w-full" />
                </label>
                <label class="text-xs text-black/60">
                  季数
                  <input
                    v-model="editForm.number_of_seasons"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Seasons"
                  />
                </label>
                <label class="text-xs text-black/60">
                  集数
                  <input
                    v-model="editForm.number_of_episodes"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="Episodes"
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
                    placeholder="8.4"
                  />
                </label>
                <label class="text-xs text-black/60">
                  热度
                  <input
                    v-model="editForm.popularity"
                    class="field-control mt-1 w-full text-sm"
                    placeholder="210.5"
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
                  @click="saveTVChanges"
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

          <!-- 季列表 -->
          <div v-if="seasonOptions.length" class="mt-6">
            <div class="mb-2 flex items-center justify-between gap-2">
              <h3 class="text-sm font-semibold">季列表</h3>
              <span class="text-xs text-black/50">点击季卡可切换分集明细</span>
            </div>
            <div class="cast-grid">
              <button
                v-for="s in seasonOptions"
                :key="s.id || s.season_number"
                type="button"
                class="cast-card rounded-xl p-1 text-left transition"
                :class="selectedSeasonNumber === s.season_number ? 'bg-sand/60 ring-2 ring-pine/25' : 'hover:bg-white/70'"
                @click="selectSeason(s.season_number)"
              >
                <img
                  :src="tmdbImg(s.poster_path, 'w185')"
                  :alt="s.name"
                  class="cast-img"
                  loading="lazy"
                />
                <p class="mt-1 truncate text-xs font-medium">{{ s.name }}</p>
                <p class="truncate text-xs text-black/50">{{ s.episode_count }} 集</p>
              </button>
            </div>
          </div>

          <div v-if="seasonOptions.length" class="panel-glass mt-4 rounded-xl p-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-semibold">
                {{ selectedSeasonDetail?.name || "分集明细" }}
              </h3>
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-xs text-black/55">
                  共 {{ selectedSeasonEpisodes.length }} 集
                </span>
                <button
                  v-if="!seasonLocalSaved"
                  type="button"
                  class="btn-soft-xs px-3 py-1 disabled:opacity-60"
                  :disabled="seasonLocalSaving || seasonDetailLoading || !selectedSeasonDetail"
                  @click="saveSeasonToLocalFromTMDB"
                >
                  {{ seasonLocalSaving ? "保存中..." : "保存到本地数据库" }}
                </button>
                <button
                  v-else
                  type="button"
                  class="rounded-lg border border-amber-300 bg-amber-50 px-3 py-1 text-xs text-amber-700 hover:bg-amber-100 disabled:opacity-60"
                  :disabled="seasonLocalSaving || seasonDetailLoading || !selectedSeasonDetail"
                  @click="saveSeasonToLocalFromTMDB"
                >
                  {{ seasonLocalSaving ? "覆盖中..." : "用 TMDB 覆盖本地" }}
                </button>
                <span v-if="seasonLocalSaved" class="text-xs text-black/50">
                  点击每集“编辑本集”可单独保存
                </span>
              </div>
            </div>
            <p v-if="selectedSeasonDetail?.overview" class="mt-2 text-xs leading-relaxed text-black/60">
              {{ selectedSeasonDetail.overview }}
            </p>
            <p v-if="seasonLocalSaved" class="mt-1 text-xs text-green-700">
              当前季已保存到本地数据库
            </p>
            <p v-if="seasonLocalMessage" class="mt-1 text-xs text-green-700">
              {{ seasonLocalMessage }}
            </p>

            <p v-if="seasonDetailLoading" class="mt-3 text-xs text-black/60">正在加载分集明细...</p>
            <p v-else-if="seasonDetailError" class="mt-3 text-xs text-red-600">{{ seasonDetailError }}</p>
            <p
              v-else-if="!selectedSeasonEpisodes.length"
              class="mt-3 rounded-lg border border-white/70 bg-white/55 px-3 py-2 text-xs text-black/55"
            >
              当前季暂无可展示的分集数据
            </p>

            <div class="mt-3 space-y-3">
              <article
                v-for="ep in selectedSeasonEpisodes"
                :key="ep.id || ep.episode_number"
                class="rounded-xl border border-white/70 bg-white/62 p-3 backdrop-blur md:flex md:gap-3"
              >
                <img
                  :src="tmdbImg(ep.still_path, 'w342')"
                  :alt="ep.name || `第${ep.episode_number}集`"
                  class="aspect-video w-full rounded-lg object-contain md:w-48"
                  loading="lazy"
                />
                <div class="mt-2 min-w-0 md:mt-0 md:flex-1">
                  <div class="flex flex-wrap gap-1.5">
                    <span class="badge">{{ formatEpisodeCode(ep.episode_number) }}</span>
                    <span class="badge">📅 {{ ep.air_date || "-" }}</span>
                    <span class="badge">⏱ {{ formatEpisodeRuntime(ep.runtime) }}</span>
                    <span class="badge">⭐ {{ formatEpisodeRating(ep.vote_average) }}</span>
                  </div>
                  <template v-if="seasonLocalSaved && editingEpisodeNumber === ep.episode_number">
                    <label class="mt-2 block text-xs text-black/60">
                      标题
                      <input
                        v-model="editingEpisodeName"
                        class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                        placeholder="请输入本集标题"
                      />
                    </label>
                    <label class="mt-2 block text-xs text-black/60">
                      简介
                      <textarea
                        v-model="editingEpisodeOverview"
                        rows="3"
                        class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                        placeholder="请输入本集简介"
                      />
                    </label>
                    <div class="mt-2 flex items-center gap-2">
                      <button
                        type="button"
                        class="btn-primary-xs disabled:opacity-60"
                        :disabled="seasonLocalSaving"
                        @click="saveEpisodeEdit"
                      >
                        {{ seasonLocalSaving ? "保存中..." : "保存本集" }}
                      </button>
                      <button
                        type="button"
                        class="btn-soft-xs disabled:opacity-60"
                        :disabled="seasonLocalSaving"
                        @click="cancelEpisodeEdit"
                      >
                        取消
                      </button>
                    </div>
                  </template>
                  <template v-else>
                    <h4 class="mt-2 truncate text-sm font-semibold">{{ ep.name || `第${ep.episode_number}集` }}</h4>
                    <p class="mt-1 text-xs leading-relaxed text-black/65">
                      {{ ep.overview || "暂无简介" }}
                    </p>
                    <button
                      v-if="seasonLocalSaved"
                      type="button"
                      class="btn-soft-xs mt-2 px-3 py-1 disabled:opacity-60"
                      :disabled="seasonLocalSaving"
                      @click="startEpisodeEdit(ep)"
                    >
                      编辑本集
                    </button>
                  </template>
                </div>
              </article>
            </div>
          </div>

          <!-- 演员 -->
          <div v-if="detail.credits?.cast?.length" class="mt-6">
            <h3 class="mb-2 text-sm font-semibold">主要演员</h3>
            <div class="cast-grid">
              <div v-for="c in detail.credits.cast.slice(0, 8)" :key="c.id" class="cast-card">
                <RouterLink :to="personLink(c.id)">
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
        你正在修改剧集 TMDB ID：
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
        确认删除剧集
        <span class="font-medium">{{ detail?.name || detail?.original_name || `ID ${tvId}` }}</span>
        的本地数据吗？
      </p>
      <p class="mt-2 text-xs text-red-700">删除后不可恢复。</p>

      <div class="mt-4 flex items-center justify-end gap-2">
        <button class="btn-soft" :disabled="deleting" @click="closeDeleteConfirmModal">取消</button>
        <button class="btn-danger-soft" :disabled="deleting" @click="confirmDeleteCurrentTV">
          {{ deleting ? "删除中..." : "确认删除" }}
        </button>
      </div>
    </section>
  </div>
</template>
