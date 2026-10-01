<script setup lang="ts">
import GlassSelect from "@/components/GlassSelect.vue";
import LoadState from "@/components/common/LoadState.vue";
import ModalShell from "@/components/common/ModalShell.vue";
import LibraryDataList from "@/components/library/LibraryDataList.vue";
import LibraryDeleteDialog from "@/components/library/LibraryDeleteDialog.vue";
import LocalMediaCreateForm from "@/components/library/LocalMediaCreateForm.vue";
import LogsPagination from "@/components/logs/LogsPagination.vue";
import type { MediaTab } from "@/components/library/types";
import { tmdbImg } from "@/api/tmdb";
import { useLibraryList } from "@/composables/useLibraryList";
import { useLocalMediaCreate } from "@/composables/useLocalMediaCreate";
import { ratingTierClass } from "@/utils/ratingTier";

let handleExternalTabChange: ((tab: MediaTab) => void) | undefined;

const {
  activeTab,
  viewMode,
  searchMode,
  keywordInput,
  keyword,
  loading,
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
} = useLibraryList({
  onExternalTabChange: (tab) => handleExternalTabChange?.(tab),
});

const {
  createPanelVisible,
  creating,
  uploadingKey,
  movieCreateForm,
  tvCreateForm,
  movieGenreOptions,
  tvGenreOptions,
  languageOptions,
  movieStatusOptions,
  tvStatusOptions,
  tvTypeOptions,
  createTitle,
  openCreatePanel,
  closeCreatePanel,
  onExternalTabChange,
  uploadCreateImage,
  submitCreate,
} = useLocalMediaCreate({
  activeTab,
  loadData,
});

handleExternalTabChange = onExternalTabChange;
</script>

<template>
  <section :class="viewMode === 'table' ? 'library-page-shell' : 'library-page-natural'">
    <section class="library-toolbar card">
      <div class="library-toolbar-main">
        <div class="library-toolbar-copy">
          <p class="section-label">本地库</p>
          <h2 class="library-toolbar-title">{{ activeTab === "movie" ? "电影库" : "剧集库" }}</h2>
          <p class="mt-1 text-sm text-muted">管理本地缓存、手动新建条目，并快速进入详情页处理字段覆盖。</p>
        </div>
      </div>

      <div class="library-filter-panel">
        <div v-if="keyword" class="library-filter-header">
          <span v-if="keyword" class="badge">当前关键词：{{ keyword }}</span>
        </div>
        <div class="library-filter-form">
          <div class="library-search-combo">
            <div class="library-switch library-media-switch library-search-media-switch" role="group" aria-label="媒体类型">
              <button
                v-for="tab in [
                  { key: 'movie', label: '电影' },
                  { key: 'tv', label: '剧集' },
                ] as const"
                :key="tab.key"
                type="button"
                class="library-switch-btn"
                :class="activeTab === tab.key ? 'library-switch-btn-active' : ''"
                @click="switchTab(tab.key as MediaTab)"
              >
                {{ tab.label }}
              </button>
            </div>
            <input
              v-model="keywordInput"
              class="library-search-input text-sm"
              placeholder="快速筛选，支持 TMDB ID、片名或剧名；可切换包含/前缀匹配。"
              @keyup.enter="applySearch"
            />
          </div>
          <GlassSelect v-model="searchMode" :options="searchModeOptions" />
          <div class="library-filter-actions">
            <button class="btn-primary library-action-btn" @click="applySearch">搜索</button>
            <button class="btn-soft library-action-btn" @click="resetSearch">重置</button>
          </div>
        </div>
      </div>
    </section>

    <!-- 卡片视图：保持原来的整页自然滚动布局 -->
    <template v-if="viewMode === 'grid'">
      <section class="library-list-summary mt-4">
        <p class="text-sm text-muted">
          共 <strong>{{ total }}</strong> 条记录 · 第 {{ page }}/{{ totalPages() }} 页
        </p>
        <div class="library-list-controls">
          <div class="library-switch library-view-switch" role="group" aria-label="视图模式">
            <button type="button" class="library-switch-btn library-switch-btn-active" @click="viewMode = 'grid'">
              卡片
            </button>
            <button type="button" class="library-switch-btn" @click="viewMode = 'table'">表格</button>
          </div>

          <button type="button" class="btn-primary library-list-create-btn" @click="openCreatePanel">
            {{ createTitle }}
          </button>
        </div>
      </section>

      <div
        v-if="refreshError && !loadError"
        class="logs-refresh-error mt-4"
        role="status"
        aria-live="polite"
      >
        <span>刷新失败：{{ refreshError }}</span>
        <button type="button" class="btn-soft-xs" :disabled="loading" @click="loadData">重试</button>
      </div>

      <LoadState class="mt-4" :loading="initialLoading" loading-text="媒体库加载中...">
        <section v-if="items.length" class="poster-grid">
          <div v-for="item in items" :key="item.tmdb_id" class="poster-card group relative">
            <button
              v-if="canDeleteItem(item)"
              type="button"
              class="library-delete-btn pointer-events-none absolute right-2 top-2 z-20 opacity-0 group-hover:pointer-events-auto group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:opacity-100"
              :class="deletingId === item.tmdb_id ? 'opacity-100 pointer-events-none' : ''"
              :disabled="deletingId === item.tmdb_id"
              :title="deletingId === item.tmdb_id ? '删除中' : '删除本地数据'"
              @click.stop="requestDeleteItem(item)"
            >
              <span v-if="deletingId === item.tmdb_id" class="text-[11px]">...</span>
              <svg
                v-else
                viewBox="0 0 24 24"
                class="h-4 w-4 fill-none stroke-current"
                stroke-width="1.8"
                aria-hidden="true"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"
                />
              </svg>
            </button>
            <RouterLink
              :to="routeByItem(item)"
              class="block"
              @pointerenter="scheduleItemDetail(item)"
              @pointerleave="cancelItemDetail(item)"
              @focus="scheduleItemDetail(item)"
              @blur="cancelItemDetail(item)"
              @touchstart.passive="touchItemDetail(item)"
            >
              <img
                :src="tmdbImg(item.poster_path, 'w342')"
                :srcset="`${tmdbImg(item.poster_path, 'w342')} 1x, ${tmdbImg(item.poster_path, 'w500')} 2x`"
                :alt="item.title || item.name"
                class="poster-img"
                loading="lazy"
              />
              <div class="poster-info">
                <p class="truncate text-sm font-medium">{{ item.title || item.name }}</p>
                <p class="poster-meta">
                  <span class="rating-badge" :class="ratingTierClass(item.vote_average)">{{ (item.vote_average ?? 0).toFixed(1) }} 分</span>
                  <span>{{ (item.release_date || item.first_air_date || "").slice(0, 4) }}</span>
                </p>
                <span v-if="item.tmdb_id < 0" class="chip-local-new mt-1 text-[10px]"> 本地新建 </span>
                <span v-else-if="item.is_modified" class="chip-modified mt-1 text-[10px]"> 已修改 </span>
              </div>
            </RouterLink>
          </div>
        </section>

        <section v-else class="empty-state mt-4">
          暂无本地数据，可以尝试切换分类、重置搜索，或新建一条本地记录。
        </section>

        <section class="mt-6">
          <LogsPagination
            :total="total"
            :page="page"
            :total-pages="totalPages()"
            :busy="loading"
            :show-stats="false"
            @change-page="gotoPage"
          />
        </section>
      </LoadState>
    </template>

    <!-- 表格视图：限高，列表内部纵向滚动；顶部汇总条与卡片视图同构 -->
    <section v-else class="card settings-card-wide settings-log-card library-result-card">
      <section class="library-list-summary">
        <p class="text-sm text-muted">
          共 <strong>{{ total }}</strong> 条记录 · 第 {{ page }}/{{ totalPages() }} 页
        </p>
        <div class="library-list-controls">
          <div class="library-switch library-view-switch" role="group" aria-label="视图模式">
            <button type="button" class="library-switch-btn" @click="viewMode = 'grid'">卡片</button>
            <button type="button" class="library-switch-btn library-switch-btn-active" @click="viewMode = 'table'">
              表格
            </button>
          </div>

          <button type="button" class="btn-primary library-list-create-btn" @click="openCreatePanel">
            {{ createTitle }}
          </button>
        </div>
      </section>

      <div
        v-if="refreshError && !loadError"
        class="logs-refresh-error"
        role="status"
        aria-live="polite"
      >
        <span>刷新失败：{{ refreshError }}</span>
        <button type="button" class="btn-soft-xs" :disabled="loading" @click="loadData">重试</button>
      </div>

      <LoadState :loading="initialLoading" loading-text="媒体库加载中...">
        <LibraryDataList
          :items="items"
          :deleting-id="deletingId"
          :can-delete-item="canDeleteItem"
          :schedule-item-detail="scheduleItemDetail"
          :cancel-item-detail="cancelItemDetail"
          :touch-item-detail="touchItemDetail"
          :open-item-detail="openItemDetail"
          :request-delete-item="requestDeleteItem"
        />
      </LoadState>

      <LogsPagination
        v-if="!loadError"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :page-size-options="pageSizeOptions"
        :total-pages="totalPages()"
        :busy="loading"
        :show-stats="false"
        @change-page="gotoPage"
        @change-page-size="changePageSize"
      />
    </section>
  </section>

  <ModalShell
    :visible="createPanelVisible"
    :title="createTitle"
    variant="vben"
    max-width-class="max-w-4xl"
    content-class="modal-scroll-content max-h-[calc(86vh-120px)] overflow-y-auto px-5 py-4"
    @close="closeCreatePanel"
  >
    <LocalMediaCreateForm
      v-model:movie-form="movieCreateForm"
      v-model:tv-form="tvCreateForm"
      :media-type="activeTab"
      :movie-genre-options="movieGenreOptions"
      :tv-genre-options="tvGenreOptions"
      :uploading-key="uploadingKey"
      :language-options="languageOptions"
      :movie-status-options="movieStatusOptions"
      :tv-status-options="tvStatusOptions"
      :tv-type-options="tvTypeOptions"
      @upload="uploadCreateImage"
    />

    <template #footer>
      <button class="btn-soft library-modal-action-btn" @click="closeCreatePanel">取消</button>
      <button class="btn-primary disabled:opacity-60" :disabled="creating || uploadingKey !== ''" @click="submitCreate">
        {{ creating ? "创建中..." : "创建并进入详情" }}
      </button>
    </template>
  </ModalShell>

  <LibraryDeleteDialog
    :visible="deleteModalVisible"
    :deleting-id="deletingId"
    :pending-delete-item="pendingDeleteItem"
    :on-close="closeDeleteModal"
    :on-confirm="confirmDeleteItem"
  />
</template>
