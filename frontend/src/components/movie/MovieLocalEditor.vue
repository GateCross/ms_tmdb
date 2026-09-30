<script setup lang="ts">
import { computed } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import type { GenreOption, MovieEditForm } from "./types";

type SelectOption = {
  label: string;
  value: string;
  hint?: string;
};

const props = defineProps<{
  isEditing: boolean;
  deleting: boolean;
  saving: boolean;
  editForm: MovieEditForm;
  genreKeyword: string;
  filteredGenreOptions: GenreOption[];
  genreOptions: GenreOption[];
  movieStatusOptions: ReadonlyArray<SelectOption>;
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
  <div class="panel-glass local-editor-panel content-auto mt-6 rounded-xl p-4">
    <div class="local-editor-header">
      <h3 class="text-sm font-semibold">本地信息编辑</h3>
      <div class="flex items-center gap-2">
        <button
          class="btn-danger-soft-xs disabled:opacity-60"
          :disabled="deleting || saving"
          @click="onDelete"
        >
          {{ deleting ? "删除中..." : "删除本地数据" }}
        </button>
        <button v-if="!isEditing" class="btn-soft-xs" @click="onEnterEdit">编辑</button>
      </div>
    </div>

    <p v-if="!isEditing" class="mt-2 text-xs text-muted">
      当前为查看模式，点击“编辑”后可修改并保存到本地数据库。
    </p>

    <div v-else class="mt-3">
      <div class="grid gap-3 md:grid-cols-2">
        <label class="text-xs text-muted">
          TMDB ID
          <input v-model="editForm.tmdb_id" class="field-control mt-1 w-full text-sm" placeholder="例如：550" />
          <p class="mt-1 text-[11px] text-warn">
            高风险：改动后，后续同步仍使用旧 TMDB ID 拉取；对外返回与访问使用新 TMDB ID。
          </p>
        </label>
        <label class="text-xs text-muted">
          片名
          <input v-model="editForm.title" class="field-control mt-1 w-full text-sm" placeholder="电影标题" />
        </label>
        <label class="text-xs text-muted">
          原始片名
          <input
            v-model="editForm.original_title"
            class="field-control mt-1 w-full text-sm"
            placeholder="Original Title"
          />
        </label>
        <label class="text-xs text-muted md:col-span-2">
          类型（多选）
          <div class="field-group-box">
            <input v-model="genreKeywordModel" class="field-control-xs w-full" placeholder="筛选类型" />
            <label v-for="genre in filteredGenreOptions" :key="genre.id" class="field-choice-pill">
              <input v-model="editForm.genre_names" type="checkbox" class="check-control" :value="genre.name" />
              <span>{{ genre.name }}</span>
            </label>
            <span v-if="!genreOptions.length" class="px-1 py-1 text-xs text-muted"> 暂无可选类型 </span>
            <span v-else-if="!filteredGenreOptions.length" class="px-1 py-1 text-xs text-muted">
              无匹配类型
            </span>
          </div>
        </label>
        <label class="text-xs text-muted">
          上映日期
          <input
            v-model="editForm.release_date"
            class="field-control mt-1 w-full text-sm"
            placeholder="YYYY-MM-DD"
          />
        </label>
        <label class="text-xs text-muted">
          状态
          <GlassSelect v-model="editForm.status" :options="movieStatusOptions" class="mt-1 w-full" />
        </label>
        <label class="text-xs text-muted">
          标语
          <input v-model="editForm.tagline" class="field-control mt-1 w-full text-sm" placeholder="Tagline" />
        </label>
        <label class="text-xs text-muted">
          时长(分钟)
          <input v-model="editForm.runtime" class="field-control mt-1 w-full text-sm" placeholder="Runtime" />
        </label>
        <label class="text-xs text-muted">
          原始语言
          <input
            v-model="editForm.original_language"
            class="field-control mt-1 w-full text-sm"
            placeholder="zh / en"
          />
        </label>
        <label class="text-xs text-muted">
          主页链接
          <input
            v-model="editForm.homepage"
            class="field-control mt-1 w-full text-sm"
            placeholder="https://..."
          />
        </label>
        <label class="text-xs text-muted">
          海报路径
          <input
            v-model="editForm.poster_path"
            class="field-control mt-1 w-full text-sm"
            placeholder="/poster.jpg"
          />
        </label>
        <label class="text-xs text-muted">
          背景图路径
          <input
            v-model="editForm.backdrop_path"
            class="field-control mt-1 w-full text-sm"
            placeholder="/backdrop.jpg"
          />
        </label>
        <label class="text-xs text-muted">
          评分
          <input v-model="editForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="7.8" />
        </label>
        <label class="text-xs text-muted">
          热度
          <input v-model="editForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="123.45" />
        </label>
        <label class="text-xs text-muted md:col-span-2">
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
        <button class="btn-primary disabled:opacity-60" :disabled="saving" @click="onSave">
          {{ saving ? "保存中..." : "保存到本地数据库" }}
        </button>
        <button class="btn-soft disabled:opacity-60" :disabled="saving" @click="onCancel">取消</button>
      </div>
    </div>
  </div>
</template>
