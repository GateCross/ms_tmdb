<script setup lang="ts">
import DataListShell from "@/components/common/DataListShell.vue";
import type { LibraryListItem } from "@/components/library/types";
import { tmdbImg } from "@/api/tmdb";
import { ratingTierClass } from "@/utils/ratingTier";

defineProps<{
  items: LibraryListItem[];
  deletingId: number | null;
  canDeleteItem: (item: LibraryListItem) => boolean;
  scheduleItemDetail: (item: LibraryListItem) => void;
  cancelItemDetail: (item: LibraryListItem) => void;
  touchItemDetail: (item: LibraryListItem) => void;
  openItemDetail: (item: LibraryListItem) => void;
  requestDeleteItem: (item: LibraryListItem) => void;
}>();

const columns = ["名称", "TMDB ID", "评分", "日期", "类型", "状态", "操作"];

// 类型列最多展示两个 chip，其余以 +N 汇总，完整列表由单元格 title 提示
function genreChips(item: LibraryListItem): string[] {
  return Array.isArray(item.genre_names) ? item.genre_names.slice(0, 2) : [];
}

function genreOverflowCount(item: LibraryListItem): number {
  return Array.isArray(item.genre_names) ? Math.max(0, item.genre_names.length - 2) : 0;
}
</script>

<template>
  <DataListShell
    shell-class="library-data-list"
    grid-class="logs-grid-library"
    :columns="columns"
    :empty="items.length === 0"
    empty-text="暂无本地数据，可以尝试切换分类、重置搜索，或新建一条本地记录。"
  >
    <article
      v-for="item in items"
      :key="item.tmdb_id"
      class="logs-row logs-grid-library"
    >
      <div class="library-cell-title">
        <img
          class="library-row-poster"
          :src="tmdbImg(item.poster_path, 'w92')"
          alt=""
          loading="lazy"
          draggable="false"
        />
        <div class="logs-main">
          <strong class="library-list-title" :title="item.title || item.name">
            {{ item.title || item.name || "-" }}
          </strong>
          <span
            class="library-list-subtitle"
            :title="item.original_title || item.original_name || ''"
          >
            {{ item.original_title || item.original_name || "-" }}
          </span>
        </div>
      </div>

      <span class="library-list-id">{{ item.tmdb_id }}</span>

      <div>
        <span class="rating-badge" :class="ratingTierClass(item.vote_average)">
          {{ (item.vote_average ?? 0).toFixed(1) }}
        </span>
      </div>

      <span class="logs-duration library-list-date">
        {{ item.release_date || item.first_air_date || "-" }}
      </span>

      <div
        class="logs-body-cell library-genre-cell"
        :title="
          Array.isArray(item.genre_names) && item.genre_names.length
            ? item.genre_names.join(' / ')
            : '-'
        "
      >
        <template v-if="genreChips(item).length">
          <span
            v-for="genre in genreChips(item)"
            :key="genre"
            class="library-genre-chip"
          >
            {{ genre }}
          </span>
          <span
            v-if="genreOverflowCount(item) > 0"
            class="library-genre-chip library-genre-chip-more"
          >
            +{{ genreOverflowCount(item) }}
          </span>
        </template>
        <span v-else class="library-list-status-muted">-</span>
      </div>

      <div class="logs-source">
        <span v-if="item.tmdb_id < 0" class="chip-local-new">本地新建</span>
        <span v-else-if="item.is_modified" class="chip-modified">已修改</span>
        <span v-else class="chip-neutral">未修改</span>
      </div>

      <div class="library-table-actions flex items-center gap-2">
        <button
          class="table-action-btn table-action-btn-soft"
          type="button"
          data-tooltip="查看详情"
          aria-label="查看详情"
          @pointerenter="scheduleItemDetail(item)"
          @pointerleave="cancelItemDetail(item)"
          @focus="scheduleItemDetail(item)"
          @blur="cancelItemDetail(item)"
          @touchstart.passive="touchItemDetail(item)"
          @click="openItemDetail(item)"
        >
          <svg
            viewBox="0 0 24 24"
            class="h-4 w-4 fill-none stroke-current"
            stroke-width="1.8"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"
            />
            <circle cx="12" cy="12" r="2.6" />
          </svg>
        </button>
        <button
          v-if="canDeleteItem(item)"
          class="table-action-btn table-action-btn-danger"
          type="button"
          :data-tooltip="deletingId === item.tmdb_id ? '删除中...' : '删除'"
          :aria-label="deletingId === item.tmdb_id ? '删除中' : '删除'"
          :disabled="deletingId === item.tmdb_id"
          @click="requestDeleteItem(item)"
        >
          <span v-if="deletingId === item.tmdb_id" class="text-[11px]">...</span>
          <svg
            v-else
            viewBox="0 0 24 24"
            class="h-4 w-4 fill-none stroke-current"
            stroke-width="1.8"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"
            />
          </svg>
        </button>
      </div>
    </article>
  </DataListShell>
</template>
