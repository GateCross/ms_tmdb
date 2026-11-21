<script setup lang="ts">
import { computed, defineAsyncComponent, watch } from "vue";
import { ArrowLeft, Copy, Pencil, Trash2 } from "lucide-vue-next";
import { tmdbImg } from "@/api/tmdb";
import LoadState from "@/components/common/LoadState.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import MovieConfirmDialogs from "@/components/movie/MovieConfirmDialogs.vue";
import { formatStatusLabel, movieStatusOptions } from "@/constants/mediaStatus";
import { useMovieDetail } from "@/composables/useMovieDetail";
import { useToastNotice } from "@/composables/useToastNotice";
import { ratingTierClass } from "@/utils/ratingTier";
import { copyText } from "@/utils/clipboard";

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
} = useMovieDetail();

const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

const deleteConfirmTitle = computed(() => {
  return detail.value?.title || detail.value?.original_title || `ID ${movieId.value}`;
});

const displayTitle = computed(() => detail.value?.title || detail.value?.original_title || "");
const displayOriginalTitle = computed(() => {
  const original = (detail.value?.original_title || "").trim();
  return original && original !== displayTitle.value ? original : "";
});

async function copyTmdbId(id: number) {
  if (!id) return;
  if (await copyText(String(id))) {
    showToastNotice("已复制 TMDB ID");
    return;
  }
  showToastNotice("复制失败，请手动复制", "info");
}

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

    <!-- 融合头部概要卡 -->
    <section class="detail-hero-card">
      <div v-if="detail.backdrop_path" class="detail-hero-backdrop">
        <img
          :src="tmdbImg(detail.backdrop_path, 'original')"
          :alt="displayTitle"
          class="detail-hero-backdrop-img"
        />
        <div class="detail-hero-backdrop-fade" aria-hidden="true"></div>
        <div class="absolute left-4 top-4 z-10">
          <button type="button" class="detail-back-btn detail-back-btn-overlay" @click="goBack">
            <ArrowLeft class="h-3.5 w-3.5" aria-hidden="true" />
            <span>返回上一页</span>
          </button>
        </div>
      </div>

      <div class="detail-hero-body">
        <div v-if="!detail.backdrop_path" class="mb-4">
          <button type="button" class="detail-back-btn" @click="goBack">
            <ArrowLeft class="h-3.5 w-3.5" aria-hidden="true" />
            <span>返回上一页</span>
          </button>
        </div>

        <div class="detail-summary-layout">
          <div class="detail-poster" :class="{ 'detail-poster-overlap': Boolean(detail.backdrop_path) }">
            <img :src="tmdbImg(detail.poster_path, 'w780')" :alt="displayTitle" class="detail-poster-img" />
          </div>

          <div class="detail-info">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h1 class="text-2xl font-semibold text-strong md:text-3xl">{{ displayTitle }}</h1>
                <p v-if="displayOriginalTitle" class="mt-0.5 text-sm text-muted">
                  {{ displayOriginalTitle }}
                </p>
                <p v-if="detail.tagline" class="mt-1 text-xs text-muted">
                  {{ detail.tagline }}
                </p>
              </div>

              <div class="flex shrink-0 flex-wrap items-center gap-2">
                <button
                  type="button"
                  class="btn-soft-xs inline-flex items-center gap-1"
                  @click="isEditing ? cancelEditMode() : enterEditMode()"
                >
                  <Pencil class="h-3.5 w-3.5" aria-hidden="true" />
                  <span>{{ isEditing ? "收起编辑" : "编辑信息" }}</span>
                </button>
                <button
                  type="button"
                  class="btn-danger-soft-xs inline-flex items-center gap-1 disabled:opacity-60"
                  :disabled="deleting || saving"
                  @click="deleteCurrentMovie"
                >
                  <Trash2 class="h-3.5 w-3.5" aria-hidden="true" />
                  <span>{{ deleting ? "删除中..." : "删除" }}</span>
                </button>
              </div>
            </div>

            <div class="detail-meta-row">
              <span class="rating-badge" :class="ratingTierClass(detail.vote_average)">
                {{ detail.vote_average == null ? "-" : `${detail.vote_average.toFixed(1)} 分` }}
              </span>
              <span class="badge">{{ formatStatusLabel(detail.status) }}</span>
              <span class="detail-meta-sep" aria-hidden="true">·</span>
              <span>上映 {{ detail.release_date || "-" }}</span>
              <template v-if="detail.runtime">
                <span class="detail-meta-sep" aria-hidden="true">·</span>
                <span>{{ detail.runtime }} 分钟</span>
              </template>
              <span class="detail-meta-sep" aria-hidden="true">·</span>
              <button
                type="button"
                class="detail-copy-btn"
                title="点击复制 TMDB ID"
                @click="copyTmdbId(currentTmdbId)"
              >
                <span>TMDB {{ currentTmdbId }}</span>
                <Copy class="h-3 w-3" aria-hidden="true" />
              </button>
              <span v-if="hasRewrittenTmdbId" class="font-mono text-xs text-muted">
                (原始 {{ originalTmdbId }})
              </span>
            </div>

            <div v-if="detail.genres?.length" class="mt-3 flex flex-wrap gap-1.5">
              <span v-for="g in detail.genres" :key="g.id" class="genre-pill">
                {{ g.name }}
              </span>
            </div>

            <p class="detail-overview">
              {{ detail.overview || "暂无简介" }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- 全宽分节：远程差异 / 编辑 / 演员 -->
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
      :saving="saving"
      :uploading-field="uploadingField"
      :edit-form="editForm"
      :genre-keyword="genreKeyword"
      :filtered-genre-options="filteredGenreOptions"
      :genre-options="genreOptions"
      :movie-status-options="movieStatusOptions"
      :on-save="saveMovieChanges"
      :on-cancel="cancelEditMode"
      :on-update-genre-keyword="updateGenreKeyword"
      :on-upload-image="uploadEditImage"
    />

    <MovieCastSection
      :credits-loading="creditsLoading"
      :credits-loaded="creditsLoaded"
      :credits-error="creditsError"
      :cast-members="castMembers"
      :person-link="personLink"
      :on-refresh="() => loadMovieCredits(true)"
    />
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
