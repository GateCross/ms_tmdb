<script setup lang="ts">
export type DataListColumn =
  | string
  | {
      label: string;
      align?: "left" | "right" | "center";
    };

withDefaults(
  defineProps<{
    columns: ReadonlyArray<DataListColumn>;
    gridClass: string;
    shellClass?: string;
    loading?: boolean;
    empty?: boolean;
    emptyText?: string;
    loadingText?: string;
  }>(),
  {
    shellClass: "",
    loading: false,
    empty: false,
    emptyText: "暂无数据",
    loadingText: "加载中...",
  },
);

function columnLabel(column: DataListColumn): string {
  return typeof column === "string" ? column : column.label;
}

function columnAlignClass(column: DataListColumn): string {
  if (typeof column === "string" || !column.align || column.align === "left") {
    return "";
  }
  return column.align === "right" ? "text-right" : "text-center";
}
</script>

<template>
  <div class="logs-list-shell" :class="[shellClass, loading && !empty ? 'data-list-refreshing' : '']">
    <div class="logs-list-head" :class="gridClass">
      <span v-for="(column, index) in columns" :key="`${columnLabel(column)}-${index}`" :class="columnAlignClass(column)">
        {{ columnLabel(column) }}
      </span>
    </div>

    <slot />

    <p v-if="loading && empty" class="logs-empty">{{ loadingText }}</p>
    <p v-else-if="empty" class="logs-empty">{{ emptyText }}</p>
  </div>
</template>
