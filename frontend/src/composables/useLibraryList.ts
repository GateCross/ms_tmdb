import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import { cancelPrefetch, prefetchMediaDetail, schedulePrefetch } from "@/api/prefetch";
import {
  deleteMovie,
  deleteTV,
  listMovies,
  listTV,
  type AdminListResp,
} from "@/api/admin";
import { clearMovieCache } from "@/api/movie";
import { clearTVCache } from "@/api/tv";
import type { LibraryListItem, MediaTab } from "@/components/library/types";
import { resolveErrorMessage } from "@/utils/errors";
import { isSameQuery, readQueryString } from "@/utils/routeQuery";

export type ViewMode = "grid" | "table";
export type SearchMode = "contains" | "prefix";
export type { LibraryListItem };

type UseLibraryListOptions = {
  onExternalTabChange?: (tab: MediaTab) => void;
};

export function useLibraryList(options: UseLibraryListOptions = {}) {
  const route = useRoute();
  const router = useRouter();

  const activeTab = ref<MediaTab>(normalizeTab(route.query.tab));
  const viewMode = ref<ViewMode>(normalizeViewMode(route.query.view));
  const searchMode = ref<SearchMode>(normalizeSearchMode(route.query.mode));
  const keywordInput = ref(readQueryString(route.query.q));
  const keyword = ref(readQueryString(route.query.q));
  let searchDebounceTimer: number | undefined;
  const loading = ref(false);
  const listLoaded = ref(false);
  const loadError = ref("");
  const refreshError = ref("");
  const items = ref<LibraryListItem[]>([]);
  const total = ref(0);
  const page = ref(normalizePage(route.query.page));
  const pageSizeOptions = [10, 20, 50, 100];
  const pageSize = ref(normalizePageSize(route.query.page_size));
  const initialLoading = computed(() => loading.value && !listLoaded.value);

  // 卡片视图每页行数固定，列数随窗口宽度自适应，从而按屏幕大小决定每页卡片数量
  const libraryGridRowCount = 2;
  const gridColumnCount = ref(currentGridColumnCount());
  const gridPageSize = computed(() => Math.max(gridColumnCount.value, 1) * libraryGridRowCount);
  // 卡片视图按屏幕自适应每页数量；表格视图使用手动选择的每页数量
  const effectivePageSize = computed(() => (viewMode.value === "grid" ? gridPageSize.value : pageSize.value));

  const deletingId = ref<number | null>(null);
  const deleteModalVisible = ref(false);
  const pendingDeleteItem = ref<LibraryListItem | null>(null);

  const searchModeOptions = [
    { label: "模糊包含", value: "contains" },
    { label: "前缀匹配", value: "prefix" },
  ] as const;

  let loadReqSeq = 0;

  function normalizeTab(value: unknown): MediaTab {
    return readQueryString(value) === "tv" ? "tv" : "movie";
  }

  function normalizeViewMode(value: unknown): ViewMode {
    return readQueryString(value) === "table" ? "table" : "grid";
  }

  function normalizeSearchMode(value: unknown): SearchMode {
    return readQueryString(value) === "prefix" ? "prefix" : "contains";
  }

  function normalizePage(value: unknown): number {
    const parsed = Number(readQueryString(value));
    return Number.isInteger(parsed) && parsed > 0 ? parsed : 1;
  }

  function normalizePageSize(value: unknown): number {
    const parsed = Number(readQueryString(value));
    if (pageSizeOptions.includes(parsed)) return parsed;
    return 20;
  }

  // 卡片网格列数与样式断点保持一致，随窗口宽度逐档变化
  function currentGridColumnCount(): number {
    if (typeof window === "undefined") return 5;
    if (window.innerWidth >= 2560) return 12;
    if (window.innerWidth >= 1920) return 10;
    if (window.innerWidth >= 1536) return 8;
    if (window.innerWidth >= 1280) return 6;
    if (window.innerWidth >= 1024) return 5;
    if (window.innerWidth >= 768) return 4;
    if (window.innerWidth >= 640) return 3;
    return 2;
  }

  function syncGridColumnCount() {
    const next = currentGridColumnCount();
    if (next === gridColumnCount.value) return;
    gridColumnCount.value = next;
    // 每页容量随列数变化时回到第一页，避免停留在超出范围的空页
    if (page.value !== 1) {
      page.value = 1;
    }
  }

  function buildLibraryQuery(): LocationQueryRaw {
    const nextQuery: LocationQueryRaw = {};
    if (activeTab.value !== "movie") nextQuery.tab = activeTab.value;
    if (page.value > 1) nextQuery.page = String(page.value);
    if (pageSize.value !== 20) nextQuery.page_size = String(pageSize.value);
    if (keyword.value) nextQuery.q = keyword.value;
    if (searchMode.value !== "contains") nextQuery.mode = searchMode.value;
    if (viewMode.value !== "grid") nextQuery.view = viewMode.value;
    return nextQuery;
  }

  function isSameLibraryQuery(nextQuery: LocationQueryRaw): boolean {
    return isSameQuery(route.query, nextQuery);
  }

  function syncLibraryQuery() {
    const nextQuery = buildLibraryQuery();
    if (isSameLibraryQuery(nextQuery)) return;
    void router.replace({ path: "/library", query: nextQuery });
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
        skip_sync: Boolean(raw.skip_sync),
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
    return raw.map((item) => normalizeListItem(item)).filter((item): item is LibraryListItem => item !== null);
  }

  function normalizeListResponse(raw: unknown): AdminListResp<LibraryListItem> {
    const payload = raw && typeof raw === "object" ? (raw as Record<string, unknown>) : {};
    return {
      total: Number(payload.total ?? 0),
      page: Number(payload.page ?? 1),
      page_size: Number(payload.page_size ?? effectivePageSize.value),
      results: normalizeListResults(payload.results),
    };
  }

  async function loadData() {
    const requestSeq = ++loadReqSeq;
    const targetTab = activeTab.value;
    const targetPage = page.value;
    const targetPageSize = effectivePageSize.value;
    const targetKeyword = keyword.value;
    const targetSearchMode = searchMode.value;
    const hadData = listLoaded.value;
    loading.value = true;
    loadError.value = "";
    refreshError.value = "";
    try {
      const resp =
        targetTab === "movie"
          ? await listMovies(targetPage, targetPageSize, targetKeyword, targetSearchMode)
          : await listTV(targetPage, targetPageSize, targetKeyword, targetSearchMode);
      if (requestSeq !== loadReqSeq) {
        return;
      }
      const normalized = normalizeListResponse(resp.data);
      items.value = normalized.results;
      total.value = normalized.total;
      listLoaded.value = true;
      loadError.value = "";
      refreshError.value = "";
    } catch (error) {
      if (requestSeq !== loadReqSeq) {
        return;
      }
      const message = resolveErrorMessage(error, "请求失败，请重试");
      if (hadData) {
        refreshError.value = message;
        loadError.value = "";
      } else {
        loadError.value = message;
        refreshError.value = "";
      }
    } finally {
      if (requestSeq === loadReqSeq) {
        loading.value = false;
      }
    }
  }

  function switchTab(tab: MediaTab) {
    if (tab === activeTab.value) return;
    activeTab.value = tab;
    page.value = 1;
    // 切换分类后保留旧列表展示，刷新失败时不降级为首载空态
  }

  function applySearch() {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = undefined;
    keyword.value = keywordInput.value.trim();
    page.value = 1;
  }

  function resetSearch() {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = undefined;
    keywordInput.value = "";
    keyword.value = "";
    page.value = 1;
  }

  function totalPages() {
    return Math.ceil(total.value / effectivePageSize.value) || 1;
  }

  function gotoPage(p: number) {
    if (p < 1 || p > totalPages()) return;
    page.value = p;
  }

  function changePageSize(nextPageSize: number) {
    const normalized = normalizePageSize(String(nextPageSize));
    if (normalized === pageSize.value) return;
    pageSize.value = normalized;
    page.value = 1;
  }

  function routeByItem(item: LibraryListItem) {
    if (activeTab.value === "movie") return `/movie/${item.tmdb_id}`;
    return `/tv/${item.tmdb_id}`;
  }

  function scheduleItemDetail(item: LibraryListItem) {
    schedulePrefetch(activeTab.value, Number(item.tmdb_id));
  }

  function cancelItemDetail(item: LibraryListItem) {
    cancelPrefetch(activeTab.value, Number(item.tmdb_id));
  }

  function touchItemDetail(item: LibraryListItem) {
    prefetchMediaDetail(activeTab.value, Number(item.tmdb_id));
  }

  function openItemDetail(item: LibraryListItem) {
    void router.push(routeByItem(item));
  }

  function canDeleteItem(item: LibraryListItem): boolean {
    const id = Number(item?.tmdb_id);
    return Number.isInteger(id) && id !== 0;
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
    try {
      if (activeTab.value === "movie") {
        await deleteMovie(id);
        clearMovieCache(id);
      } else {
        await deleteTV(id);
        clearTVCache(id);
      }

      if (items.value.length <= 1 && page.value > 1) {
        page.value = page.value - 1;
        closeDeleteModal();
        return;
      }
      await loadData();
      closeDeleteModal();
    } catch {
      /* handled by global toast */
    } finally {
      deletingId.value = null;
    }
  }

  watch(
    () => route.query,
    (query) => {
      const nextTab = normalizeTab(query.tab);
      if (nextTab !== activeTab.value) {
        activeTab.value = nextTab;
        page.value = 1;
        options.onExternalTabChange?.(nextTab);
      }

      const nextPage = normalizePage(query.page);
      if (nextPage !== page.value) {
        page.value = nextPage;
      }

      const nextKeyword = readQueryString(query.q);
      if (nextKeyword !== keyword.value) {
        keyword.value = nextKeyword;
        keywordInput.value = nextKeyword;
      }

      const nextSearchMode = normalizeSearchMode(query.mode);
      if (nextSearchMode !== searchMode.value) {
        searchMode.value = nextSearchMode;
      }

      const nextViewMode = normalizeViewMode(query.view);
      if (nextViewMode !== viewMode.value) {
        viewMode.value = nextViewMode;
      }

      const nextPageSize = normalizePageSize(query.page_size);
      if (nextPageSize !== pageSize.value) {
        pageSize.value = nextPageSize;
      }
    },
  );

  watch(keywordInput, (val) => {
    const next = val.trim();
    if (next === keyword.value) return;
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = window.setTimeout(() => {
      keyword.value = next;
      page.value = 1;
    }, 350);
  });

  onBeforeUnmount(() => {
    clearTimeout(searchDebounceTimer);
  });
  watch([activeTab, page, effectivePageSize, keyword, searchMode], loadData);
  watch([activeTab, page, pageSize, keyword, searchMode, viewMode], syncLibraryQuery);
  onMounted(() => {
    syncGridColumnCount();
    window.addEventListener("resize", syncGridColumnCount, { passive: true });
    loadData();
  });
  onBeforeUnmount(() => {
    window.removeEventListener("resize", syncGridColumnCount);
  });

  return {
    activeTab,
    viewMode,
    searchMode,
    keywordInput,
    keyword,
    loading,
    listLoaded,
    loadError,
    refreshError,
    items,
    total,
    page,
    pageSize,
    pageSizeOptions,
    initialLoading,
    deletingId,
    deleteModalVisible,
    pendingDeleteItem,
    searchModeOptions,
    loadData,
    switchTab,
    applySearch,
    resetSearch,
    totalPages,
    gotoPage,
    changePageSize,
    routeByItem,
    scheduleItemDetail,
    cancelItemDetail,
    touchItemDetail,
    openItemDetail,
    canDeleteItem,
    requestDeleteItem,
    closeDeleteModal,
    confirmDeleteItem,
  };
}
