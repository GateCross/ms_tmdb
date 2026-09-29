<script setup lang="ts">
import { computed, defineAsyncComponent, watch } from "vue";
import { tmdbImg } from "@/api/tmdb";
import LoadState from "@/components/common/LoadState.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import MovieConfirmDialogs from "@/components/movie/MovieConfirmDialogs.vue";
import { formatStatusLabel, movieStatusOptions } from "@/constants/mediaStatus";
import { useMovieDetail } from "@/composables/useMovieDetail";
import { useToastNotice } from "@/composables/useToastNotice";
import { ratingTierClass } from "@/utils/ratingTier";

const MovieRemoteDiffCard = defineAsyncComponent(() => import("@/components/movie/MovieRemoteDiffCard.vue"));
const MovieLocalEditor = defineAsyncComponent(() => import("@/components/movie/MovieLocalEditor.vue"));
const MovieCastSection = defineAsyncComponent(() => import("@/components/movie/MovieCastSection.vue"));

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
  saveMovieChanges,
  loadMovieCredits,
  loadData,
  confirmDeleteCurrentMovie,
} = useMovieDetail();

const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

const deleteConfirmTitle = computed(() => {
  return detail.value?.title || detail.value?.original_title || `ID ${movieId.value}`;
});

watch(saveMessage, (message) => {
  if (message.trim()) {
    showToastNotice(message);
    saveMessage.value = "";
  }
});
</script>

<template>
  <LoadState
    v-if="!detail"
    class="card"
    :loading="loading"
    loading-text="电影详情加载中..."
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
        :alt="detail.title || detail.original_title"
        class="hero-banner-media"
      />
      <div class="absolute left-4 top-4 z-10">
        <button class="detail-back-btn" @click="goBack">返回上一页</button>
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
          <img :src="tmdbImg(detail.poster_path, 'w780')" :alt="detail.title" class="detail-poster-img" />
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
            <span class="rating-badge" :class="ratingTierClass(detail.vote_average)">
              {{ detail.vote_average == null ? "-" : `${detail.vote_average.toFixed(1)} 分` }}
            </span>
            <span class="badge">上映 {{ detail.release_date ?? "-" }}</span>
            <span v-if="detail.runtime" class="badge">片长 {{ detail.runtime }} 分钟</span>
            <span class="badge">{{ formatStatusLabel(detail.status) }}</span>
          </div>

          <!-- 类型标签 -->
          <div v-if="detail.genres?.length" class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="g in detail.genres" :key="g.id" class="genre-pill">
              {{ g.name }}
            </span>
          </div>

          <p class="mt-4 text-sm leading-relaxed text-black/75">
            {{ detail.overview || "暂无简介" }}
          </p>

          <MovieRemoteDiffCard
            :target-id="movieId"
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

          <MovieLocalEditor
            :is-editing="isEditing"
            :deleting="deleting"
            :saving="saving"
            :edit-form="editForm"
            :genre-keyword="genreKeyword"
            :filtered-genre-options="filteredGenreOptions"
            :genre-options="genreOptions"
            :movie-status-options="movieStatusOptions"
            :on-delete="deleteCurrentMovie"
            :on-enter-edit="enterEditMode"
            :on-save="saveMovieChanges"
            :on-cancel="cancelEditMode"
            :on-update-genre-keyword="updateGenreKeyword"
          />

          <MovieCastSection
            :credits-loading="creditsLoading"
            :credits-loaded="creditsLoaded"
            :credits-error="creditsError"
            :cast-members="castMembers"
            :person-link="personLink"
            :on-refresh="() => loadMovieCredits(true)"
          />
        </div>
      </div>
    </section>
  </template>

  <MovieConfirmDialogs
    :tmdb-risk-modal-visible="tmdbRiskModalVisible"
    :tmdb-risk-current-id="tmdbRiskCurrentId"
    :tmdb-risk-next-id="tmdbRiskNextId"
    :delete-confirm-modal-visible="deleteConfirmModalVisible"
    :deleting="deleting"
    :movie-title="deleteConfirmTitle"
    :on-close-tmdb-risk="closeTmdbRiskModal"
    :on-close-delete-confirm="closeDeleteConfirmModal"
    :on-confirm-delete="confirmDeleteCurrentMovie"
  />

  <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
</template>
