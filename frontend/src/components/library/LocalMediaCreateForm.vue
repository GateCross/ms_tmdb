<script setup lang="ts">
import GlassSelect from "@/components/GlassSelect.vue";
import GenreCheckboxGroup from "@/components/library/GenreCheckboxGroup.vue";
import ImageUploadField from "@/components/library/ImageUploadField.vue";
import type { GenreOption } from "@/utils/mediaNormalizers";
import type { LocalMovieCreateForm, LocalTVCreateForm, MediaTab, SelectOption, UploadingKey } from "./types";

defineProps<{
  mediaType: MediaTab;
  movieGenreOptions: GenreOption[];
  tvGenreOptions: GenreOption[];
  uploadingKey: UploadingKey;
  languageOptions: readonly SelectOption[];
  movieStatusOptions: readonly SelectOption[];
  tvStatusOptions: readonly SelectOption[];
  tvTypeOptions: readonly SelectOption[];
}>();

const movieForm = defineModel<LocalMovieCreateForm>("movieForm", { required: true });
const tvForm = defineModel<LocalTVCreateForm>("tvForm", { required: true });

const emit = defineEmits<{
  upload: [mediaType: MediaTab, field: "poster_path" | "backdrop_path", event: Event];
}>();
</script>

<template>
  <div v-if="mediaType === 'movie'" class="grid gap-3 md:grid-cols-2">
    <label class="text-xs text-black/60">
      标题（必填）
      <input v-model="movieForm.title" class="field-control mt-1 w-full text-sm" placeholder="电影标题" />
    </label>
    <label class="text-xs text-black/60">
      原始标题
      <input
        v-model="movieForm.original_title"
        class="field-control mt-1 w-full text-sm"
        placeholder="Original Title"
      />
    </label>
    <label class="text-xs text-black/60">
      上映日期
      <input v-model="movieForm.release_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
    </label>
    <label class="text-xs text-black/60">
      状态
      <GlassSelect v-model="movieForm.status" :options="movieStatusOptions" class="mt-1 w-full" />
    </label>
    <label class="text-xs text-black/60">
      原始语言
      <GlassSelect v-model="movieForm.original_language" :options="languageOptions" class="mt-1 w-full" />
    </label>
    <label class="text-xs text-black/60">
      时长（分钟）
      <input v-model="movieForm.runtime" class="field-control mt-1 w-full text-sm" placeholder="120" />
    </label>
    <label class="text-xs text-black/60 md:col-span-2">
      类型（多选）
      <GenreCheckboxGroup v-model="movieForm.genre_names" :options="movieGenreOptions" />
    </label>
    <ImageUploadField
      v-model="movieForm.poster_path"
      label="海报路径"
      :uploading="uploadingKey === 'movie_poster_path'"
      @upload="(event) => emit('upload', 'movie', 'poster_path', event)"
    />
    <ImageUploadField
      v-model="movieForm.backdrop_path"
      label="背景图路径"
      :uploading="uploadingKey === 'movie_backdrop_path'"
      @upload="(event) => emit('upload', 'movie', 'backdrop_path', event)"
    />
    <label class="text-xs text-black/60">
      评分
      <input v-model="movieForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="7.8" />
    </label>
    <label class="text-xs text-black/60">
      热度
      <input v-model="movieForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="123.4" />
    </label>
    <label class="text-xs text-black/60 md:col-span-2">
      简介
      <textarea v-model="movieForm.overview" rows="3" class="field-control mt-1 w-full text-sm" placeholder="简介" />
    </label>
  </div>

  <div v-else class="grid gap-3 md:grid-cols-2">
    <label class="text-xs text-black/60">
      剧名（必填）
      <input v-model="tvForm.name" class="field-control mt-1 w-full text-sm" placeholder="剧集名称" />
    </label>
    <label class="text-xs text-black/60">
      原始剧名
      <input v-model="tvForm.original_name" class="field-control mt-1 w-full text-sm" placeholder="Original Name" />
    </label>
    <label class="text-xs text-black/60">
      首播日期
      <input v-model="tvForm.first_air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
    </label>
    <label class="text-xs text-black/60">
      状态
      <GlassSelect v-model="tvForm.status" :options="tvStatusOptions" class="mt-1 w-full" />
    </label>
    <label class="text-xs text-black/60">
      原始语言
      <GlassSelect v-model="tvForm.original_language" :options="languageOptions" class="mt-1 w-full" />
    </label>
    <label class="text-xs text-black/60">
      剧集类型
      <GlassSelect v-model="tvForm.type" :options="tvTypeOptions" class="mt-1 w-full" />
    </label>
    <label class="text-xs text-black/60">
      季数
      <input v-model="tvForm.number_of_seasons" class="field-control mt-1 w-full text-sm" placeholder="3" />
    </label>
    <label class="text-xs text-black/60">
      集数
      <input v-model="tvForm.number_of_episodes" class="field-control mt-1 w-full text-sm" placeholder="24" />
    </label>
    <label class="text-xs text-black/60 md:col-span-2">
      类型（多选）
      <GenreCheckboxGroup v-model="tvForm.genre_names" :options="tvGenreOptions" />
    </label>
    <ImageUploadField
      v-model="tvForm.poster_path"
      label="海报路径"
      :uploading="uploadingKey === 'tv_poster_path'"
      @upload="(event) => emit('upload', 'tv', 'poster_path', event)"
    />
    <ImageUploadField
      v-model="tvForm.backdrop_path"
      label="背景图路径"
      :uploading="uploadingKey === 'tv_backdrop_path'"
      @upload="(event) => emit('upload', 'tv', 'backdrop_path', event)"
    />
    <label class="text-xs text-black/60">
      评分
      <input v-model="tvForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="8.1" />
    </label>
    <label class="text-xs text-black/60">
      热度
      <input v-model="tvForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="220.5" />
    </label>
    <label class="text-xs text-black/60 md:col-span-2">
      简介
      <textarea v-model="tvForm.overview" rows="3" class="field-control mt-1 w-full text-sm" placeholder="简介" />
    </label>
  </div>
</template>
