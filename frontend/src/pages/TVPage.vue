<script setup lang="ts">
import { defineAsyncComponent } from "vue";
import { tmdbImg } from "@/api/tmdb";
import { formatStatusLabel, formatTvTypeLabel, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";
import { useTVDetailPage } from "@/composables/useTVDetailPage";

const TVRemoteDiffCard = defineAsyncComponent(() => import("@/components/tv/TVRemoteDiffCard.vue"));
const TVLocalEditorCard = defineAsyncComponent(() => import("@/components/tv/TVLocalEditorCard.vue"));
const TVSeasonManager = defineAsyncComponent(() => import("@/components/tv/TVSeasonManager.vue"));
const TVCastSection = defineAsyncComponent(() => import("@/components/tv/TVCastSection.vue"));
const {
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
} = useTVDetailPage();
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
          <img :src="tmdbImg(detail.poster_path, 'w342')" :alt="detail.name" class="detail-poster-img" />
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
            <span v-for="g in detail.genres" :key="g.id" class="genre-pill">
              {{ g.name }}
            </span>
          </div>

          <p class="mt-4 text-sm leading-relaxed text-black/75">
            {{ detail.overview || "暂无简介" }}
          </p>

          <TVRemoteDiffCard
            :target-id="tvId"
            :checking-remote-diff="checkingRemoteDiff"
            :remote-diff-notice="remoteDiffNotice"
            :remote-diff-message="remoteDiffMessage"
            :remote-diff-error="remoteDiffError"
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
            :save-message="saveMessage"
            :save-error="saveError"
            :delete-error="deleteError"
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
            :season-panel-visible="seasonPanelVisible"
            :selected-season-detail="selectedSeasonDetail"
            :selected-season-episodes="selectedSeasonEpisodes"
            :season-editor-visible="seasonEditorVisible"
            :season-editor-mode="seasonEditorMode"
            :season-local-saved="seasonLocalSaved"
            :season-local-message="seasonLocalMessage"
            :season-form-error="seasonFormError"
            :episode-form-error="episodeFormError"
            :season-form="seasonForm"
            :episode-creator-visible="episodeCreatorVisible"
            :episode-form="episodeForm"
            :season-detail-error="seasonDetailError"
            :editing-episode-number="editingEpisodeNumber"
            :episode-edit-changed-count="episodeEditChangedCount"
            :episode-edit-field-class="episodeEditFieldClass"
            :format-episode-code="formatEpisodeCode"
            :format-episode-runtime="formatEpisodeRuntime"
            :format-episode-rating="formatEpisodeRating"
            :on-open-season-create-editor="openSeasonCreateEditor"
            :on-select-season="selectSeason"
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
            :on-prefetch-person="prefetchPerson"
          />
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

  <div
    v-if="localDeleteConfirmModalVisible"
    class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/45 p-4"
    @click.self="closeLocalDeleteConfirmModal(false)"
  >
    <section class="panel-glass w-full max-w-md rounded-2xl p-5">
      <h3 class="text-base font-semibold text-red-700">{{ localDeleteConfirmTitle || "删除确认" }}</h3>
      <p class="mt-2 text-sm text-black/75">
        {{ localDeleteConfirmMessage }}
      </p>
      <p class="mt-2 text-xs text-red-700">删除后不可恢复。</p>

      <div class="mt-4 flex items-center justify-end gap-2">
        <button class="btn-soft" @click="closeLocalDeleteConfirmModal(false)">取消</button>
        <button class="btn-danger-soft" @click="closeLocalDeleteConfirmModal(true)">
          {{ localDeleteConfirmActionText }}
        </button>
      </div>
    </section>
  </div>
</template>
