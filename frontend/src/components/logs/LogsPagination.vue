<script setup lang="ts">
import { computed } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";

type PageItem = number | "ellipsis";

const props = withDefaults(
  defineProps<{
    total: number;
    page: number;
    totalPages: number;
    busy: boolean;
    small?: boolean;
    pageSize?: number;
    pageSizeOptions?: number[];
    showStats?: boolean;
  }>(),
  {
    small: false,
    pageSize: 0,
    pageSizeOptions: () => [10, 20, 50, 100],
    showStats: true,
  },
);

const emit = defineEmits<{
  "change-page": [page: number];
  "change-page-size": [pageSize: number];
}>();

const showPageSize = computed(() => props.pageSize > 0 && props.pageSizeOptions.length > 0);
const pageSizeValue = computed(() => String(props.pageSize));
const pageSizeSelectOptions = computed(() =>
  props.pageSizeOptions.map((size) => ({ label: `${size} 条`, value: String(size) })),
);

// 页码窗口：总数少直接全出，多时保留首尾页 + 当前页邻域，间隙折叠为省略号
const pageItems = computed<PageItem[]>(() => {
  const total = Math.max(1, props.totalPages);
  if (total <= 7) {
    return Array.from({ length: total }, (_, index) => index + 1);
  }
  const pages = [
    ...new Set([1, total, props.page - 1, props.page, props.page + 1].filter((page) => page >= 1 && page <= total)),
  ].sort((a, b) => a - b);
  return pages.flatMap((page, index) => (index && page - pages[index - 1] > 1 ? ["ellipsis" as const, page] : [page]));
});

function changePageSize(value: string) {
  const next = Number(value);
  if (Number.isFinite(next) && next > 0 && next !== props.pageSize) {
    emit("change-page-size", Math.trunc(next));
  }
}
</script>

<template>
  <div class="settings-pagination-row" :class="small ? 'settings-pagination-row-sm' : ''">
    <p v-if="showStats">共 {{ total }} 条记录 · 第 {{ page }}/{{ totalPages }} 页</p>
    <div class="settings-pagination-actions">
      <label v-if="showPageSize" class="settings-pagination-size">
        每页
        <GlassSelect
          :model-value="pageSizeValue"
          :options="pageSizeSelectOptions"
          :disabled="busy"
          class="settings-pagination-size-select"
          @change="changePageSize"
        />
      </label>
      <button
        class="btn-soft px-3 py-1.5 disabled:opacity-60"
        :disabled="busy || page <= 1"
        @click="emit('change-page', page - 1)"
      >
        上一页
      </button>
      <nav v-if="pageItems.length > 1" class="settings-pagination-nav" aria-label="分页">
        <template v-for="(item, index) in pageItems" :key="`${item}-${index}`">
          <span v-if="item === 'ellipsis'" class="settings-pagination-ellipsis" aria-hidden="true">…</span>
          <button
            v-else
            class="settings-pagination-page"
            :class="item === page ? 'settings-pagination-page-active' : ''"
            type="button"
            :disabled="busy"
            :aria-current="item === page ? 'page' : undefined"
            @click="emit('change-page', item)"
          >
            {{ item }}
          </button>
        </template>
      </nav>
      <button
        class="btn-soft px-3 py-1.5 disabled:opacity-60"
        :disabled="busy || page >= totalPages"
        @click="emit('change-page', page + 1)"
      >
        下一页
      </button>
    </div>
  </div>
</template>
