<script setup lang="ts">
import { computed, ref, watch } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";

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

const jumpPage = ref(props.page);
const showPageSize = computed(() => props.pageSize > 0 && props.pageSizeOptions.length > 0);
const pageSizeValue = computed(() => String(props.pageSize));
const pageSizeSelectOptions = computed(() =>
  props.pageSizeOptions.map((size) => ({ label: `${size} 条`, value: String(size) })),
);

function normalizeJumpPage(value: number) {
  const next = Number.isFinite(value) ? Math.trunc(value) : props.page;
  if (next < 1) return 1;
  if (next > props.totalPages) return props.totalPages;
  return next;
}

function submitJumpPage() {
  const target = normalizeJumpPage(Number(jumpPage.value));
  jumpPage.value = target;
  if (target !== props.page) {
    emit("change-page", target);
  }
}

function changePageSize(value: string) {
  const next = Number(value);
  if (Number.isFinite(next) && next > 0 && next !== props.pageSize) {
    emit("change-page-size", Math.trunc(next));
  }
}

watch(
  () => props.page,
  (value) => {
    jumpPage.value = value;
  },
);
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
      <form class="settings-pagination-jump" @submit.prevent="submitJumpPage">
        <span>跳至</span>
        <input
          v-model.number="jumpPage"
          class="field-control-xs settings-pagination-input"
          type="number"
          min="1"
          :max="totalPages"
          :disabled="busy"
          aria-label="跳转页码"
          @blur="jumpPage = normalizeJumpPage(Number(jumpPage))"
        />
        <span>页</span>
        <button class="btn-soft px-3 py-1.5 disabled:opacity-60" type="submit" :disabled="busy">跳转</button>
      </form>
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
