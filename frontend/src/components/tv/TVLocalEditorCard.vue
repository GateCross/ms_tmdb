<script setup lang="ts">
import { computed } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import ImageUploadField from "@/components/library/ImageUploadField.vue";
import type { GenreOption, TVEditForm } from "./types";

type SelectOption = {
  label: string;
  value: string;
  hint?: string;
};

const props = defineProps<{
  isEditing: boolean;
  saving: boolean;
  uploadingField: "" | "poster_path" | "backdrop_path";
  editForm: TVEditForm;
  genreKeyword: string;
  filteredGenreOptions: GenreOption[];
  genreOptions: GenreOption[];
  tvStatusOptions: ReadonlyArray<SelectOption>;
  tvTypeOptions: ReadonlyArray<SelectOption>;
  onSave: () => void;
  onCancel: () => void;
  onUpdateGenreKeyword: (value: string) => void;
  onUploadImage: (field: "poster_path" | "backdrop_path", event: Event) => void;
}>();

const genreKeywordModel = computed({
  get: () => props.genreKeyword,
  set: (value: string) => props.onUpdateGenreKeyword(value),
});
</script>

<template>
  <section v-if="isEditing" class="card local-editor-panel content-auto mt-4">
    <div class="local-editor-header pb-3 border-b border-line">
      <div>
        <h3 class="text-sm font-semibold text-strong">编辑剧集信息</h3>
        <p class="mt-0.5 text-xs text-muted">修改后将写入本地数据库，并标记为本地覆盖字段。</p>
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="btn-soft-xs disabled:opacity-60" :disabled="saving" @click="onCancel">取消</button>
        <button
          type="button"
          class="btn-primary-xs disabled:opacity-60"
          :disabled="saving || uploadingField !== ''"
          @click="onSave"
        >
          {{ saving ? "保存中..." : "保存到本地数据库" }}
        </button>
      </div>
    </div>

    <div class="mt-4 space-y-5">
      <div class="detail-editor-section">
        <h4 class="detail-editor-section-title">基础信息</h4>
        <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          <label class="text-xs text-muted">
            剧名
            <input v-model="editForm.name" class="field-control mt-1 w-full text-sm" placeholder="剧集标题" />
          </label>
          <label class="text-xs text-muted">
            原始剧名
            <input
              v-model="editForm.original_name"
              class="field-control mt-1 w-full text-sm"
              placeholder="Original Name"
            />
          </label>
          <label class="text-xs text-muted">
            原始语言
            <input v-model="editForm.original_language" class="field-control mt-1 w-full text-sm" placeholder="zh / en" />
          </label>
          <label class="text-xs text-muted md:col-span-2 lg:col-span-3">
            标语
            <input v-model="editForm.tagline" class="field-control mt-1 w-full text-sm" placeholder="Tagline" />
          </label>
          <div class="text-xs text-muted md:col-span-2 lg:col-span-3">
            <span>类型（多选）</span>
            <div class="field-group-box">
              <input v-model="genreKeywordModel" class="field-control-xs w-full sm:w-56" placeholder="筛选类型" />
              <div class="flex w-full flex-wrap gap-1.5">
                <label v-for="genre in filteredGenreOptions" :key="genre.id" class="field-choice-pill">
                  <input v-model="editForm.genre_names" type="checkbox" class="check-control" :value="genre.name" />
                  <span>{{ genre.name }}</span>
                </label>
                <span v-if="!genreOptions.length" class="px-1 py-1 text-xs text-muted">暂无可选类型</span>
                <span v-else-if="!filteredGenreOptions.length" class="px-1 py-1 text-xs text-muted">无匹配类型</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="detail-editor-section">
        <h4 class="detail-editor-section-title">发行、规模与评分</h4>
        <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
          <label class="text-xs text-muted">
            首播日期
            <input v-model="editForm.first_air_date" class="field-control mt-1 w-full text-sm" placeholder="YYYY-MM-DD" />
          </label>
          <label class="text-xs text-muted">
            状态
            <GlassSelect v-model="editForm.status" :options="tvStatusOptions" class="mt-1 w-full" />
          </label>
          <label class="text-xs text-muted">
            剧集类型
            <GlassSelect v-model="editForm.type" :options="tvTypeOptions" class="mt-1 w-full" />
          </label>
          <label class="text-xs text-muted">
            主页链接
            <input v-model="editForm.homepage" class="field-control mt-1 w-full text-sm" placeholder="https://..." />
          </label>
          <label class="text-xs text-muted">
            季数
            <input v-model="editForm.number_of_seasons" class="field-control mt-1 w-full text-sm" placeholder="1" />
          </label>
          <label class="text-xs text-muted">
            集数
            <input v-model="editForm.number_of_episodes" class="field-control mt-1 w-full text-sm" placeholder="12" />
          </label>
          <label class="text-xs text-muted">
            评分
            <input v-model="editForm.vote_average" class="field-control mt-1 w-full text-sm" placeholder="8.4" />
          </label>
          <label class="text-xs text-muted">
            热度
            <input v-model="editForm.popularity" class="field-control mt-1 w-full text-sm" placeholder="210.5" />
          </label>
        </div>
      </div>

      <div class="detail-editor-section">
        <h4 class="detail-editor-section-title">媒体图片与简介</h4>
        <div class="grid gap-3 md:grid-cols-2">
          <ImageUploadField
            v-model="editForm.poster_path"
            label="海报路径"
            editable
            preview-aspect="poster"
            placeholder="/poster.jpg 或本地上传"
            :uploading="uploadingField === 'poster_path'"
            @upload="(event) => onUploadImage('poster_path', event)"
          />
          <ImageUploadField
            v-model="editForm.backdrop_path"
            label="背景图路径"
            editable
            preview-aspect="backdrop"
            placeholder="/backdrop.jpg 或本地上传"
            :uploading="uploadingField === 'backdrop_path'"
            @upload="(event) => onUploadImage('backdrop_path', event)"
          />
          <label class="text-xs text-muted md:col-span-2">
            简介
            <textarea
              v-model="editForm.overview"
              rows="4"
              class="field-control mt-1 w-full text-sm"
              placeholder="请输入剧集简介"
            />
          </label>
        </div>
      </div>

      <div class="detail-editor-section">
        <div class="detail-editor-danger-box">
          <label class="block text-xs font-medium text-warn">
            TMDB ID（高风险操作）
            <input
              v-model="editForm.tmdb_id"
              class="field-control mt-1.5 w-full font-mono text-sm sm:max-w-xs"
              placeholder="例如：1399"
            />
          </label>
          <p class="mt-1.5 text-xs text-warn">
            改动后，后续同步仍使用旧 TMDB ID 拉取；对外返回与页面访问将使用新的 TMDB ID。
          </p>
        </div>
      </div>
    </div>

    <div class="mt-4 flex items-center justify-end gap-2 border-t border-line pt-4">
      <button type="button" class="btn-soft disabled:opacity-60" :disabled="saving" @click="onCancel">取消</button>
      <button
        type="button"
        class="btn-primary disabled:opacity-60"
        :disabled="saving || uploadingField !== ''"
        @click="onSave"
      >
        {{ saving ? "保存中..." : "保存到本地数据库" }}
      </button>
    </div>
  </section>
</template>
