<script setup lang="ts">
import { tmdbImg } from "@/api/tmdb";
import type { TVEpisodeForm, TVEpisodeItem, TVSeasonDetail, TVSeasonForm, TVSeasonSummary } from "./types";

defineProps<{
  seasonOptions: TVSeasonSummary[];
  selectedSeasonNumber: number | null;
  seasonLocalSaving: boolean;
  seasonDetailLoading: boolean;
  seasonPanelVisible: boolean;
  selectedSeasonDetail: TVSeasonDetail | null;
  selectedSeasonEpisodes: TVEpisodeItem[];
  seasonEditorVisible: boolean;
  seasonEditorMode: "create" | "edit";
  seasonLocalSaved: boolean;
  seasonLocalMessage: string;
  seasonFormError: string;
  episodeFormError: string;
  seasonForm: TVSeasonForm;
  episodeCreatorVisible: boolean;
  episodeForm: TVEpisodeForm;
  seasonDetailError: string;
  editingEpisodeNumber: number | null;
  episodeEditChangedCount: number;
  episodeEditFieldClass: (field: keyof TVEpisodeForm) => string;
  formatEpisodeCode: (episodeNumber: number) => string;
  formatEpisodeRuntime: (runtime: number | null) => string;
  formatEpisodeRating: (voteAverage: number | null) => string;
  onOpenSeasonCreateEditor: () => void;
  onSelectSeason: (seasonNumber: number) => void;
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
  <div class="content-auto mt-6">
    <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
      <div>
        <h3 class="text-sm font-semibold">季管理</h3>
        <p class="mt-1 text-xs text-black/50">点击季卡按需加载分集明细，也可手动新增本地季</p>
      </div>
      <button
        type="button"
        class="btn-soft-xs px-3 py-1 disabled:opacity-60"
        :disabled="seasonLocalSaving || seasonDetailLoading"
        @click="onOpenSeasonCreateEditor"
      >
        新增季
      </button>
    </div>

    <div v-if="seasonOptions.length" class="cast-grid">
      <button
        v-for="s in seasonOptions"
        :key="s.id || s.season_number"
        type="button"
        class="cast-card season-card"
        :class="selectedSeasonNumber === s.season_number ? 'season-card-active' : ''"
        @click="onSelectSeason(s.season_number)"
      >
        <img :src="tmdbImg(s.poster_path, 'w185')" :alt="s.name" class="cast-img" loading="lazy" />
        <p class="mt-1 truncate text-xs font-medium">{{ s.name }}</p>
        <p class="truncate text-xs text-black/50">{{ s.episode_count }} 集</p>
      </button>
    </div>
    <div v-else class="season-empty">当前剧集还没有可展示的季数据，可以直接新增本地季。</div>
  </div>

  <div v-if="seasonPanelVisible" class="panel-glass content-auto-heavy mt-4 rounded-xl p-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h3 class="text-sm font-semibold">
        {{
          seasonEditorVisible && seasonEditorMode === "create"
            ? "新增本地季"
            : selectedSeasonDetail?.name || "季信息与分集"
        }}
      </h3>
      <div class="flex flex-wrap items-center gap-2">
        <span
          v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
          class="text-xs text-black/55"
        >
          共 {{ selectedSeasonEpisodes.length }} 集
        </span>
        <button
          v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
          type="button"
          class="btn-soft-xs px-3 py-1 disabled:opacity-60"
          :disabled="seasonLocalSaving || seasonDetailLoading"
          @click="onOpenSeasonEditEditor"
        >
          编辑本季
        </button>
        <button
          v-if="selectedSeasonDetail && seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
          type="button"
          class="btn-soft-xs px-3 py-1 disabled:opacity-60"
          :disabled="seasonLocalSaving || seasonDetailLoading"
          @click="onOpenEpisodeCreateEditor"
        >
          手动新增集
        </button>
        <button
          v-if="selectedSeasonDetail && !seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
          type="button"
          class="btn-soft-xs px-3 py-1 disabled:opacity-60"
          :disabled="seasonLocalSaving || seasonDetailLoading || !selectedSeasonDetail"
          @click="onSaveSeasonToLocal"
        >
          {{ seasonLocalSaving ? "保存中..." : "保存到本地数据库" }}
        </button>
        <button
          v-else-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')"
          type="button"
          class="detail-alert-action disabled:opacity-60"
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
        <span v-if="seasonLocalSaved" class="text-xs text-black/50"> 可编辑单集，也可手动新增或删除本地季 </span>
      </div>
    </div>
    <p
      v-if="selectedSeasonDetail?.overview && !(seasonEditorVisible && seasonEditorMode === 'create')"
      class="mt-2 text-xs leading-relaxed text-black/60"
    >
      {{ selectedSeasonDetail.overview }}
    </p>
    <p
      v-if="seasonLocalSaved && !(seasonEditorVisible && seasonEditorMode === 'create')"
      class="mt-1 text-xs text-green-700"
    >
      当前季已保存到本地数据库
    </p>
    <p v-if="seasonLocalMessage" class="mt-1 text-xs text-green-700">
      {{ seasonLocalMessage }}
    </p>
    <p v-if="seasonFormError" class="mt-1 text-xs text-red-600">
      {{ seasonFormError }}
    </p>
    <p v-if="episodeFormError" class="mt-1 text-xs text-red-600">
      {{ episodeFormError }}
    </p>

    <div v-if="seasonEditorVisible" class="season-editor-box">
      <div class="grid gap-3 md:grid-cols-2">
        <label class="text-xs text-black/60">
          季号
          <input
            v-model="seasonForm.season_number"
            class="field-control mt-1 w-full text-sm"
            :disabled="seasonEditorMode === 'edit'"
            placeholder="例如：1"
          />
        </label>
        <label class="text-xs text-black/60">
          季标题
          <input v-model="seasonForm.name" class="field-control mt-1 w-full text-sm" placeholder="例如：第一季" />
        </label>
        <label class="text-xs text-black/60">
          首播日期
          <input v-model="seasonForm.air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
        </label>
        <label class="text-xs text-black/60">
          海报路径
          <input v-model="seasonForm.poster_path" class="field-control mt-1 w-full text-sm" placeholder="/poster.jpg" />
        </label>
        <label class="text-xs text-black/60 md:col-span-2">
          简介
          <textarea
            v-model="seasonForm.overview"
            rows="3"
            class="field-control mt-1 w-full text-sm"
            placeholder="请输入季简介"
          />
        </label>
      </div>
      <div class="mt-3 flex items-center gap-3">
        <button
          type="button"
          class="btn-primary disabled:opacity-60"
          :disabled="seasonLocalSaving"
          @click="onSaveSeasonEditor"
        >
          {{ seasonLocalSaving ? "保存中..." : seasonEditorMode === "create" ? "创建本季" : "保存本季" }}
        </button>
        <button
          type="button"
          class="btn-soft disabled:opacity-60"
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
      <div class="grid gap-3 md:grid-cols-2">
        <label class="text-xs text-black/60">
          集号
          <input v-model="episodeForm.episode_number" class="field-control mt-1 w-full text-sm" placeholder="例如：1" />
        </label>
        <label class="text-xs text-black/60">
          标题
          <input v-model="episodeForm.name" class="field-control mt-1 w-full text-sm" placeholder="请输入本集标题" />
        </label>
        <label class="text-xs text-black/60">
          播出日期
          <input v-model="episodeForm.air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
        </label>
        <label class="text-xs text-black/60">
          时长
          <input v-model="episodeForm.runtime" class="field-control mt-1 w-full text-sm" placeholder="分钟" />
        </label>
        <label class="text-xs text-black/60">
          评分
          <input v-model="episodeForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="8.6" />
        </label>
        <label class="text-xs text-black/60">
          剧照路径
          <input v-model="episodeForm.still_path" class="field-control mt-1 w-full text-sm" placeholder="/still.jpg" />
        </label>
        <label class="text-xs text-black/60 md:col-span-2">
          简介
          <textarea
            v-model="episodeForm.overview"
            rows="3"
            class="field-control mt-1 w-full text-sm"
            placeholder="请输入本集简介"
          />
        </label>
      </div>
      <div class="mt-3 flex items-center gap-3">
        <button
          type="button"
          class="btn-primary disabled:opacity-60"
          :disabled="seasonLocalSaving"
          @click="onSaveEpisodeCreate"
        >
          {{ seasonLocalSaving ? "保存中..." : "创建本集" }}
        </button>
        <button
          type="button"
          class="btn-soft disabled:opacity-60"
          :disabled="seasonLocalSaving"
          @click="onCloseEpisodeCreator"
        >
          取消
        </button>
      </div>
    </div>

    <p v-if="seasonDetailLoading" class="mt-3 text-xs text-black/60">正在加载分集明细...</p>
    <p v-else-if="seasonDetailError" class="mt-3 text-xs text-red-600">{{ seasonDetailError }}</p>
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

    <div v-if="selectedSeasonDetail && !(seasonEditorVisible && seasonEditorMode === 'create')" class="mt-3 space-y-3">
      <article v-for="ep in selectedSeasonEpisodes" :key="ep.id || ep.episode_number" class="episode-card">
        <img
          :src="tmdbImg(ep.still_path, 'w342')"
          :alt="ep.name || `第${ep.episode_number}集`"
          class="episode-still"
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
            <p class="mt-2 text-xs" :class="episodeEditChangedCount > 0 ? 'text-amber-700' : 'text-black/55'">
              {{ episodeEditChangedCount > 0 ? `已修改 ${episodeEditChangedCount} 个字段` : "尚未修改字段" }}
            </p>
            <div class="mt-2 grid gap-2 md:grid-cols-3">
              <label :class="['text-xs text-black/60', episodeEditFieldClass('episode_number')]">
                集号
                <input
                  v-model="episodeForm.episode_number"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="例如：1"
                />
              </label>
              <label :class="['text-xs text-black/60', episodeEditFieldClass('name')]">
                标题
                <input
                  v-model="episodeForm.name"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="请输入本集标题"
                />
              </label>
              <label :class="['text-xs text-black/60', episodeEditFieldClass('air_date')]">
                播出日期
                <input
                  v-model="episodeForm.air_date"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="YYYY-MM-DD"
                />
              </label>
              <label :class="['text-xs text-black/60', episodeEditFieldClass('runtime')]">
                时长
                <input
                  v-model="episodeForm.runtime"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="分钟"
                />
              </label>
              <label :class="['text-xs text-black/60', episodeEditFieldClass('vote_average')]">
                评分
                <input
                  v-model="episodeForm.vote_average"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="8.6"
                />
              </label>
              <label :class="['text-xs text-black/60', episodeEditFieldClass('still_path')]">
                剧照路径
                <input
                  v-model="episodeForm.still_path"
                  class="field-control mt-1 w-full px-2.5 py-1.5 text-sm"
                  placeholder="/still.jpg"
                />
              </label>
              <label :class="['text-xs text-black/60 md:col-span-3', episodeEditFieldClass('overview')]">
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
            <h4 class="mt-2 truncate text-sm font-semibold">{{ ep.name || `第${ep.episode_number}集` }}</h4>
            <p class="mt-1 text-xs leading-relaxed text-black/65">
              {{ ep.overview || "暂无简介" }}
            </p>
            <div v-if="seasonLocalSaved" class="mt-2 flex items-center gap-2">
              <button
                type="button"
                class="btn-soft-xs px-3 py-1 disabled:opacity-60"
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
</template>
