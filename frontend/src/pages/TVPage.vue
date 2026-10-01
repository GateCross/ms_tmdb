<script setup lang="ts">
import { defineAsyncComponent, watch } from "vue";
import { tmdbImg } from "@/api/tmdb";
import BaseDialog from "@/components/common/BaseDialog.vue";
import LoadState from "@/components/common/LoadState.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import { formatStatusLabel, formatTvTypeLabel, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";
import { useToastNotice } from "@/composables/useToastNotice";
import { useTVDetailPage } from "@/composables/useTVDetailPage";
import { ratingTierClass } from "@/utils/ratingTier";

const TVRemoteDiffCard = defineAsyncComponent(() => import("@/components/tv/TVRemoteDiffCard.vue"));
const TVLocalEditorCard = defineAsyncComponent(() => import("@/components/tv/TVLocalEditorCard.vue"));
const TVSeasonManager = defineAsyncComponent(() => import("@/components/tv/TVSeasonManager.vue"));
const TVCastSection = defineAsyncComponent(() => import("@/components/tv/TVCastSection.vue"));
const {
  loading,
  refreshError,
  detail,
  castMembers,
  creditsLoading,
  creditsLoaded,
  creditsError,
  isEditing,
  saving,
  deleting,
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
  episodeCreatorVisible,
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
  updateGenreKeyword,
  toggleRemoteDiffDetails,
  toggleLocalOverrideDiffDetails,
  closeTmdbRiskModal,
  closeDeleteConfirmModal,
  closeLocalDeleteConfirmModal,
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
  loadSeasonDetail,
  loadData,
  confirmDeleteCurrentTV,
} = useTVDetailPage();

const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

watch(saveMessage, (message) => {
  if (message.trim()) {
    showToastNotice(message);
  }
});

watch(seasonLocalMessage, (message) => {
  if (message.trim()) {
    showToastNotice(message);
  }
});
</script>

<template>
  <LoadState
    v-if="!detail"
    class="card"
    :loading="loading"
    loading-text="剧集详情加载中..."
  />

  <template v-else>
    <div
      v-if="refreshError"
      class="logs-refresh-error mb-4"
      role="status"
      aria-live="polite"
    >
      <span>刷新失败：{{ refreshError }}</span>
      <button type="button" class="btn-soft-xs" :disabled="loading" @click="() => loadData({ force: true })">
        重试
      </button>
    </div>
    <!-- 背景横幅 -->
    <section class="hero-banner hero-banner-detail">
      <img
        :src="tmdbImg(detail.backdrop_path, 'original')"
        :alt="detail.name || detail.original_name"
        class="hero-banner-media"
      />
      <div class="absolute left-4 top-4 z-10">
        <button class="detail-back-btn" @click="goBack">返回上一页</button>
      </div>
      <div class="hero-overlay">
        <h1 class="text-2xl font-bold text-white md:text-3xl">{{ detail.name || detail.original_name }}</h1>
        <p class="mt-1 text-sm text-white/70">{{ detail.tagline }}</p>
      </div>
    </section>

    <section class="card mt-4">
      <div class="detail-layout">
        <div class="detail-poster">
          <img :src="tmdbImg(detail.poster_path, 'w780')" :alt="detail.name" class="detail-poster-img" />
        </div>

        <div class="detail-info">
          <h2 class="text-xl font-bold">{{ detail.name }}</h2>
          <p v-if="detail.original_name !== detail.name" class="text-sm text-muted">
            {{ detail.original_name }}
          </p>
          <div class="mt-2 grid gap-1 text-xs text-muted sm:grid-cols-2">
            <template v-if="hasRewrittenTmdbId">
              <p>
                修改后 TMDB ID：
                <span class="font-medium text-ink">{{ currentTmdbId }}</span>
              </p>
              <p>
                原始 TMDB ID：
                <span class="font-medium text-ink">{{ originalTmdbId }}</span>
              </p>
            </template>
            <p v-else>
              TMDB ID：
              <span class="font-medium text-ink">{{ currentTmdbId }}</span>
            </p>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <span class="rating-badge" :class="ratingTierClass(detail.vote_average)">
              {{ detail.vote_average == null ? "-" : `${detail.vote_average.toFixed(1)} 分` }}
            </span>
            <span class="badge">首播 {{ detail.first_air_date ?? "-" }}</span>
            <span v-if="detail.number_of_seasons" class="badge">
              {{ detail.number_of_seasons }} 季 · {{ detail.number_of_episodes }} 集
            </span>
            <span class="badge">{{ formatStatusLabel(detail.status) }}</span>
            <span v-if="detail.type" class="badge">{{ formatTvTypeLabel(detail.type) }}</span>
          </div>

          <div v-if="detail.genres?.length" class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="g in detail.genres" :key="g.id" class="genre-pill">
              {{ g.name }}
            </span>
          </div>

          <p class="mt-4 text-sm leading-relaxed text-muted">
            {{ detail.overview || "暂无简介" }}
          </p>

          <TVRemoteDiffCard
            :target-id="tvId"
            :checking-remote-diff="checkingRemoteDiff"
            :remote-diff-notice="remoteDiffNotice"
            :remote-diff-message="remoteDiffMessage"
            :remote-diff-decision="remoteDiffDecision"
            :show-remote-diff-details="showRemoteDiffDetails"
            :show-local-override-diff-details="showLocalOverrideDiffDetails"
            :should-show-sync-panel="shouldShowSyncPanel"
            :allowed-sync-modes="allowedSyncModes"
            :on-toggle-remote-details="toggleRemoteDiffDetails"
            :on-toggle-local-details="toggleLocalOverrideDiffDetails"
            :on-keep-local="keepLocalData"
            :on-synced="handleSynced"
          />

          <TVLocalEditorCard
            :is-editing="isEditing"
            :deleting="deleting"
            :saving="saving"
            :edit-form="editForm"
            :genre-keyword="genreKeyword"
            :filtered-genre-options="filteredGenreOptions"
            :genre-options="genreOptions"
            :tv-status-options="tvStatusOptions"
            :tv-type-options="tvTypeOptions"
            :on-delete="deleteCurrentTV"
            :on-enter-edit="enterEditMode"
            :on-save="saveTVChanges"
            :on-cancel="cancelEditMode"
            :on-update-genre-keyword="updateGenreKeyword"
          />

          <TVSeasonManager
            :season-options="seasonOptions"
            :selected-season-number="selectedSeasonNumber"
            :season-local-saving="seasonLocalSaving"
            :season-detail-loading="seasonDetailLoading"
            :season-detail-error="seasonDetailError"
            :season-panel-visible="seasonPanelVisible"
            :selected-season-detail="selectedSeasonDetail"
            :selected-season-episodes="selectedSeasonEpisodes"
            :season-editor-visible="seasonEditorVisible"
            :season-editor-mode="seasonEditorMode"
            :season-local-saved="seasonLocalSaved"
            :season-form="seasonForm"
            :episode-creator-visible="episodeCreatorVisible"
            :episode-form="episodeForm"
            :editing-episode-number="editingEpisodeNumber"
            :episode-edit-changed-count="episodeEditChangedCount"
            :episode-edit-field-class="episodeEditFieldClass"
            :format-episode-code="formatEpisodeCode"
            :format-episode-runtime="formatEpisodeRuntime"
            :format-episode-rating="formatEpisodeRating"
            :on-open-season-create-editor="openSeasonCreateEditor"
            :on-select-season="selectSeason"
            :on-retry-season-detail="
              () => selectedSeasonNumber != null && loadSeasonDetail(selectedSeasonNumber, true)
            "
            :on-open-season-edit-editor="openSeasonEditEditor"
            :on-open-episode-create-editor="openEpisodeCreateEditor"
            :on-save-season-to-local="saveSeasonToLocalFromTMDB"
            :on-delete-season-local-data="deleteSeasonLocalData"
            :on-save-season-editor="saveSeasonEditor"
            :on-close-season-editor="closeSeasonEditor"
            :on-save-episode-create="saveEpisodeCreate"
            :on-close-episode-creator="closeEpisodeCreator"
            :on-start-episode-edit="startEpisodeEdit"
            :on-save-episode-edit="saveEpisodeEdit"
            :on-cancel-episode-edit="cancelEpisodeEdit"
            :on-delete-episode="deleteEpisode"
          />

          <TVCastSection
            :credits-loading="creditsLoading"
            :credits-loaded="creditsLoaded"
            :credits-error="creditsError"
            :cast-members="castMembers"
            :person-link="personLink"
            :on-refresh="() => loadTVCredits(true)"
          />
        </div>
      </div>
    </section>
  </template>

  <BaseDialog
    :visible="tmdbRiskModalVisible"
    title="修改 TMDB ID 风险确认"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="bg-overlay"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-4 flex items-center justify-end gap-2 px-5 pb-5"
    @close="closeTmdbRiskModal(false)"
  >
    <template #title>
      <span class="text-base font-semibold text-warn">修改 TMDB ID 风险确认</span>
    </template>

    <p class="text-sm text-muted">
      你正在修改剧集 TMDB ID：
      <span class="font-medium">{{ tmdbRiskCurrentId }}</span>
      ->
      <span class="font-medium">{{ tmdbRiskNextId }}</span>
    </p>
    <div class="mt-3 rounded-lg border border-warn-line bg-warn-soft p-3 text-xs leading-relaxed text-warn">
      <p>1) 这是高风险操作，可能导致与第三方历史引用不一致；</p>
      <p>2) 之后自动/手动同步将继续使用旧 TMDB ID 向 TMDB 拉取；</p>
      <p>3) 对外返回与页面访问将使用新的 TMDB ID。</p>
    </div>

    <template #footer>
      <button class="btn-soft" @click="closeTmdbRiskModal(false)">取消</button>
      <button class="btn-primary" data-dialog-primary @click="closeTmdbRiskModal(true)">确认继续</button>
    </template>
  </BaseDialog>

  <BaseDialog
    :visible="deleteConfirmModalVisible"
    title="删除本地数据确认"
    :busy="deleting"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="bg-overlay"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-4 flex items-center justify-end gap-2 px-5 pb-5"
    @close="closeDeleteConfirmModal"
  >
    <template #title>
      <span class="text-base font-semibold text-danger">删除本地数据确认</span>
    </template>

    <p class="text-sm text-muted">
      确认删除剧集
      <span class="font-medium">{{ detail?.name || detail?.original_name || `ID ${tvId}` }}</span>
      的本地数据吗？
    </p>
    <p class="mt-2 text-xs text-danger">删除后不可恢复。</p>

    <template #footer>
      <button class="btn-soft" :disabled="deleting" @click="closeDeleteConfirmModal">取消</button>
      <button class="btn-danger-soft" data-dialog-primary :disabled="deleting" @click="confirmDeleteCurrentTV">
        {{ deleting ? "删除中..." : "确认删除" }}
      </button>
    </template>
  </BaseDialog>

  <BaseDialog
    :visible="localDeleteConfirmModalVisible"
    :title="localDeleteConfirmTitle || '删除确认'"
    :show-close-button="false"
    max-width-class="max-w-md"
    root-class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    overlay-class="bg-overlay"
    header-class="px-5 pt-5 pb-0"
    content-class="px-5 pt-2 pb-0"
    footer-class="mt-4 flex items-center justify-end gap-2 px-5 pb-5"
    @close="closeLocalDeleteConfirmModal(false)"
  >
    <template #title>
      <span class="text-base font-semibold text-danger">{{ localDeleteConfirmTitle || "删除确认" }}</span>
    </template>

    <p class="text-sm text-muted">
      {{ localDeleteConfirmMessage }}
    </p>
    <p class="mt-2 text-xs text-danger">删除后不可恢复。</p>

    <template #footer>
      <button class="btn-soft" @click="closeLocalDeleteConfirmModal(false)">取消</button>
      <button class="btn-danger-soft" data-dialog-primary @click="closeLocalDeleteConfirmModal(true)">
        {{ localDeleteConfirmActionText }}
      </button>
    </template>
  </BaseDialog>

  <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
</template>
