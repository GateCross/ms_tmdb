<script setup lang="ts">
import { tmdbImg } from "@/api/tmdb";
import { ratingTierClass } from "@/utils/ratingTier";
import type { TVEpisodeForm, TVEpisodeItem, TVSeasonDetail, TVSeasonForm, TVSeasonSummary } from "./types";

defineProps<{
  seasonOptions: TVSeasonSummary[];
  selectedSeasonNumber: number | null;
  seasonLocalSaving: boolean;
  seasonDetailLoading: boolean;
  seasonDetailError?: string;
  seasonPanelVisible: boolean;
  selectedSeasonDetail: TVSeasonDetail | null;
  selectedSeasonEpisodes: TVEpisodeItem[];
  seasonEditorVisible: boolean;
  seasonEditorMode: "create" | "edit";
  seasonLocalSaved: boolean;
  seasonForm: TVSeasonForm;
  episodeCreatorVisible: boolean;
  episodeForm: TVEpisodeForm;
  editingEpisodeNumber: number | null;
  episodeEditChangedCount: number;
  episodeEditFieldClass: (field: keyof TVEpisodeForm) => string;
  formatEpisodeCode: (episodeNumber: number) => string;
  formatEpisodeRuntime: (runtime: number | null) => string;
  formatEpisodeRating: (voteAverage: number | null) => string;
  onOpenSeasonCreateEditor: () => void;
  onSelectSeason: (seasonNumber: number) => void;
  onRetrySeasonDetail?: () => void;
  onOpenSeasonEditEditor: () => void;
  onOpenEpisodeCreateEditor: () => void;
  onSaveSeasonToLocal: () => void;
  onDeleteSeasonLocalData: () => void;
  onSaveSeasonEditor: () => void;
  onCloseSeasonEditor: () => void;
  onSaveEpisodeCreate: () => void;
  onCloseEpisodeCreator: () => void;
  onStartEpisodeEdit: (episode: TVEpisodeItem) => void;
  onSaveEpisodeEdit: () => void;
  onCancelEpisodeEdit: () => void;
  onDeleteEpisode: (episodeNumber: number) => void;
}>();
</script>

<template>
  <section class="card content-auto mt-4">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <h3 class="text-sm font-semibold text-strong">季与分集管理</h3>
        <span class="text-xs text-muted">点击季卡加载分集明细，也可新增本地季</span>
      </div>
      <button
        type="button"
        class="btn-soft-xs disabled:opacity-60"
        :disabled="seasonLocalSaving || seasonDetailLoading"
        @click="onOpenSeasonCreateEditor"
      >
        新增季
      </button>
    </div>

    <div v-if="seasonOptions.length" class="season-strip">
      <button
        v-for="s in seasonOptions"
        :key="s.id || s.season_number"
        type="button"
        class="season-card"
        :class="selectedSeasonNumber === s.season_number ? 'season-card-active' : ''"
        @click="onSelectSeason(s.season_number)"
      >
        <img :src="tmdbImg(s.poster_path, 'w185')" :alt="s.name" class="cast-img" loading="lazy" />
        <p class="mt-1.5 truncate text-xs font-medium text-strong">{{ s.name }}</p>
        <p class="truncate text-xs text-muted tabular-nums">{{ s.episode_count }} 集</p>
      </button>
    </div>
    <div v-else class="season-empty">当前剧集还没有可展示的季数据，可以直接新增本地季。</div>

    <div v-if="seasonPanelVisible" class="content-auto-heavy mt-4 border-t border-line pt-4">
      <div
        v-if="seasonDetailError && !seasonDetailLoading && !selectedSeasonDetail"
        class="logs-refresh-error mb-3"
        role="status"
        aria-live="polite"
      >
        <span>{{ seasonDetailError }}</span>
        <button
          v-if="onRetrySeasonDetail"
          type="button"
          class="btn-soft-xs"
          :disabled="seasonDetailLoading"
          @click="onRetrySeasonDetail"
        >
          重试
        </button>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex flex-wrap items-center gap-2">
          <h4 class="text-sm font-semibold text-strong">
            {{
              seasonEditorVisible && seasonEditorMode === "create"
                ? "新增本地季"
                : selectedSeasonDetail?.name || "季信息与分集"
            }}
          </h4>
          <span
            v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
            class="text-xs text-muted tabular-nums"
          >
            共 {{ selectedSeasonEpisodes.length }} 集
          </span>
          <span
            v-if="seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
            class="chip-local-new"
          >
            已存本地
          </span>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
            type="button"
            class="btn-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving || seasonDetailLoading"
            @click="onOpenSeasonEditEditor"
          >
            编辑本季
          </button>
          <button
            v-if="selectedSeasonDetail && seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
            type="button"
            class="btn-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving || seasonDetailLoading"
            @click="onOpenEpisodeCreateEditor"
          >
            手动新增集
          </button>
          <button
            v-if="selectedSeasonDetail && !seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
            type="button"
            class="btn-primary-xs disabled:opacity-60"
            :disabled="seasonLocalSaving || seasonDetailLoading || !selectedSeasonDetail"
            @click="onSaveSeasonToLocal"
          >
            {{ seasonLocalSaving ? "保存中..." : "保存到本地数据库" }}
          </button>
          <button
            v-else-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
            type="button"
            class="btn-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving || seasonDetailLoading || !selectedSeasonDetail"
            @click="onSaveSeasonToLocal"
          >
            {{ seasonLocalSaving ? "覆盖中..." : "用 TMDB 覆盖本地" }}
          </button>
          <button
            v-if="selectedSeasonDetail && seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
            type="button"
            class="btn-danger-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving || seasonDetailLoading"
            @click="onDeleteSeasonLocalData"
          >
            {{ seasonLocalSaving ? "删除中..." : "删除本季" }}
          </button>
        </div>
      </div>
      <p
        v-if="selectedSeasonDetail?.overview && !(seasonEditorVisible && seasonEditorMode === 'create')"
        class="mt-2 max-w-4xl text-xs leading-relaxed text-muted"
      >
        {{ selectedSeasonDetail.overview }}
      </p>

      <div v-if="seasonEditorVisible" class="season-editor-box">
        <div class="grid gap-3 md:grid-cols-2">
          <label class="text-xs text-muted">
            季号
            <input
              v-model="seasonForm.season_number"
              class="field-control mt-1 w-full text-sm"
              :disabled="seasonEditorMode === 'edit'"
              placeholder="例如：1"
            />
          </label>
          <label class="text-xs text-muted">
            季标题
            <input v-model="seasonForm.name" class="field-control mt-1 w-full text-sm" placeholder="例如：第一季" />
          </label>
          <label class="text-xs text-muted">
            首播日期
            <input v-model="seasonForm.air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
          </label>
          <label class="text-xs text-muted">
            海报路径
            <input v-model="seasonForm.poster_path" class="field-control mt-1 w-full text-sm" placeholder="/poster.jpg" />
          </label>
          <label class="text-xs text-muted md:col-span-2">
            简介
            <textarea
              v-model="seasonForm.overview"
              rows="3"
              class="field-control mt-1 w-full text-sm"
              placeholder="请输入季简介"
            />
          </label>
        </div>
        <div class="mt-3 flex items-center gap-2">
          <button
            type="button"
            class="btn-primary-xs disabled:opacity-60"
            :disabled="seasonLocalSaving"
            @click="onSaveSeasonEditor"
          >
            {{ seasonLocalSaving ? "保存中..." : seasonEditorMode === "create" ? "创建本季" : "保存本季" }}
          </button>
          <button
            type="button"
            class="btn-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving"
            @click="onCloseSeasonEditor"
          >
            取消
          </button>
        </div>
      </div>

      <div
        v-if="episodeCreatorVisible && selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
        class="season-editor-box"
      >
        <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          <label class="text-xs text-muted">
            集号
            <input v-model="episodeForm.episode_number" class="field-control mt-1 w-full text-sm" placeholder="例如：1" />
          </label>
          <label class="text-xs text-muted">
            标题
            <input v-model="episodeForm.name" class="field-control mt-1 w-full text-sm" placeholder="请输入本集标题" />
          </label>
          <label class="text-xs text-muted">
            播出日期
            <input v-model="episodeForm.air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
          </label>
          <label class="text-xs text-muted">
            时长
            <input v-model="episodeForm.runtime" class="field-control mt-1 w-full text-sm" placeholder="分钟" />
          </label>
          <label class="text-xs text-muted">
            评分
            <input v-model="episodeForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="8.6" />
          </label>
          <label class="text-xs text-muted">
            剧照路径
            <input v-model="episodeForm.still_path" class="field-control mt-1 w-full text-sm" placeholder="/still.jpg" />
          </label>
          <label class="text-xs text-muted md:col-span-2 lg:col-span-3">
            简介
            <textarea
              v-model="episodeForm.overview"
              rows="3"
              class="field-control mt-1 w-full text-sm"
              placeholder="请输入本集简介"
            />
          </label>
        </div>
        <div class="mt-3 flex items-center gap-2">
          <button
            type="button"
            class="btn-primary-xs disabled:opacity-60"
            :disabled="seasonLocalSaving"
            @click="onSaveEpisodeCreate"
          >
            {{ seasonLocalSaving ? "保存中..." : "创建本集" }}
          </button>
          <button
            type="button"
            class="btn-soft-xs disabled:opacity-60"
            :disabled="seasonLocalSaving"
            @click="onCloseEpisodeCreator"
          >
            取消
          </button>
        </div>
      </div>

      <p v-if="seasonDetailLoading" class="mt-3 text-xs text-muted">正在加载分集明细...</p>
      <p v-else-if="!selectedSeasonDetail && !seasonEditorVisible" class="season-inline-note mt-3">
        请选择一个已有季，或点击“新增季”创建本地季。
      </p>
      <p
        v-else-if="
          selectedSeasonDetail &&
          !selectedSeasonEpisodes.length &&
          !(seasonEditorVisible && seasonEditorMode === 'create')
        "
        class="season-inline-note mt-3"
      >
        当前季暂无可展示的分集数据
      </p>

      <div v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')" class="mt-3 space-y-2.5">
        <article v-for="ep in selectedSeasonEpisodes" :key="ep.id || ep.episode_number" class="episode-card">
          <img
            :src="tmdbImg(ep.still_path, 'w342')"
            :alt="ep.name || `第${ep.episode_number}集`"
            class="episode-still"
            loading="lazy"
          />
          <div class="mt-2 min-w-0 sm:mt-0 sm:flex-1">
            <template v-if="seasonLocalSaved && editingEpisodeNumber === ep.episode_number">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="font-mono text-xs font-semibold text-brand">
                  {{ formatEpisodeCode(ep.episode_number) }}
                </span>
                <p class="text-xs" :class="episodeEditChangedCount > 0 ? 'text-warn' : 'text-muted'">
                  {{ episodeEditChangedCount > 0 ? `已修改 ${episodeEditChangedCount} 个字段` : "尚未修改字段" }}
                </p>
              </div>
              <div class="mt-2 grid gap-2 md:grid-cols-3">
                <label :class="['text-xs text-muted', episodeEditFieldClass('episode_number')]">
                  集号
                  <input
                    v-model="episodeForm.episode_number"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="例如：1"
                  />
                </label>
                <label :class="['text-xs text-muted', episodeEditFieldClass('name')]">
                  标题
                  <input
                    v-model="episodeForm.name"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="请输入本集标题"
                  />
                </label>
                <label :class="['text-xs text-muted', episodeEditFieldClass('air_date')]">
                  播出日期
                  <input
                    v-model="episodeForm.air_date"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="YYYY-MM-DD"
                  />
                </label>
                <label :class="['text-xs text-muted', episodeEditFieldClass('runtime')]">
                  时长
                  <input
                    v-model="episodeForm.runtime"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="分钟"
                  />
                </label>
                <label :class="['text-xs text-muted', episodeEditFieldClass('vote_average')]">
                  评分
                  <input
                    v-model="episodeForm.vote_average"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="8.6"
                  />
                </label>
                <label :class="['text-xs text-muted', episodeEditFieldClass('still_path')]">
                  剧照路径
                  <input
                    v-model="episodeForm.still_path"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="/still.jpg"
                  />
                </label>
                <label :class="['text-xs text-muted md:col-span-3', episodeEditFieldClass('overview')]">
                  简介
                  <textarea
                    v-model="episodeForm.overview"
                    rows="3"
                    class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                    placeholder="请输入本集简介"
                  />
                </label>
              </div>
              <div class="mt-2 flex items-center gap-2">
                <button
                  type="button"
                  class="btn-primary-xs disabled:opacity-60"
                  :disabled="seasonLocalSaving"
                  @click="onSaveEpisodeEdit"
                >
                  {{ seasonLocalSaving ? "保存中..." : "保存本集修改" }}
                </button>
                <button
                  type="button"
                  class="btn-soft-xs disabled:opacity-60"
                  :disabled="seasonLocalSaving"
                  @click="onCancelEpisodeEdit"
                >
                  取消
                </button>
              </div>
            </template>
            <template v-else>
              <div class="flex flex-wrap items-start justify-between gap-2">
                <div class="flex min-w-0 items-baseline gap-2">
                  <span class="shrink-0 font-mono text-xs font-semibold text-brand">
                    {{ formatEpisodeCode(ep.episode_number) }}
                  </span>
                  <h4 class="truncate text-sm font-semibold text-strong">{{ ep.name || `第${ep.episode_number}集` }}</h4>
                </div>
                <span class="rating-badge" :class="ratingTierClass(ep.vote_average)">
                  {{ formatEpisodeRating(ep.vote_average) }}
                </span>
              </div>
              <p class="mt-1 flex flex-wrap items-center gap-x-2 text-xs text-muted tabular-nums">
                <span>播出 {{ ep.air_date || "-" }}</span>
                <span aria-hidden="true">·</span>
                <span>片长 {{ formatEpisodeRuntime(ep.runtime) }}</span>
              </p>
              <p class="mt-1.5 text-xs leading-relaxed text-muted">
                {{ ep.overview || "暂无简介" }}
              </p>
              <div v-if="seasonLocalSaved" class="mt-2.5 flex items-center gap-2">
                <button
                  type="button"
                  class="btn-soft-xs disabled:opacity-60"
                  :disabled="seasonLocalSaving"
                  @click="onStartEpisodeEdit(ep)"
                >
                  编辑本集
                </button>
                <button
                  type="button"
                  class="btn-danger-soft-xs disabled:opacity-60"
                  :disabled="seasonLocalSaving"
                  @click="onDeleteEpisode(ep.episode_number)"
                >
                  {{ seasonLocalSaving ? "删除中..." : "删除本集" }}
                </button>
              </div>
            </template>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
