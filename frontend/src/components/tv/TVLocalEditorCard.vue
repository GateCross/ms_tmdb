<script setup lang="ts">
import { computed } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import type { GenreOption, TVEditForm } from "./types";

type SelectOption = {
  label: string;
  value: string;
  hint?: string;
};

const props = defineProps<{
  isEditing: boolean;
  deleting: boolean;
  saving: boolean;
  editForm: TVEditForm;
  genreKeyword: string;
  filteredGenreOptions: GenreOption[];
  genreOptions: GenreOption[];
  saveMessage: string;
  saveError: string;
  deleteError: string;
  tvStatusOptions: ReadonlyArray<SelectOption>;
  tvTypeOptions: ReadonlyArray<SelectOption>;
  onDelete: () => void;
  onEnterEdit: () => void;
  onSave: () => void;
  onCancel: () => void;
  onUpdateGenreKeyword: (value: string) => void;
}>();

const genreKeywordModel = computed({
  get: () => props.genreKeyword,
  set: (value: string) => props.onUpdateGenreKeyword(value),
});
</script>

<template>
  <div class="panel-glass content-auto mt-6 rounded-xl p-4">
    <div class="flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold">本地信息编辑</h3>
      <div class="flex items-center gap-2">
        <button
          class="btn-danger-soft-xs disabled:opacity-60"
          :disabled="deleting || saving"
          @click="onDelete"
        >
          {{ deleting ? "删除中..." : "删除本地数据" }}
        </button>
        <button
          v-if="!isEditing"
          class="btn-soft-xs"
          @click="onEnterEdit"
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
          <div class="mt-1 flex flex-wrap gap-2 rounded-lg border border-white/70 bg-white/55 p-2">
            <input
              v-model="genreKeywordModel"
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
          @click="onSave"
        >
          {{ saving ? "保存中..." : "保存到本地数据库" }}
        </button>
        <button
          class="btn-soft disabled:opacity-60"
          :disabled="saving"
          @click="onCancel"
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
</template>
