<script setup lang="ts">
import type { GenreOption } from "@/utils/mediaNormalizers";

const props = defineProps<{
  modelValue: string[];
  options: GenreOption[];
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string[]];
}>();

function toggleGenre(name: string, checked: boolean) {
  if (checked) {
    emit("update:modelValue", props.modelValue.includes(name) ? props.modelValue : [...props.modelValue, name]);
    return;
  }
  emit(
    "update:modelValue",
    props.modelValue.filter((item) => item !== name),
  );
}
</script>

<template>
  <div class="mt-1 max-h-32 overflow-y-auto rounded-lg border border-white/70 bg-white/55 p-2 backdrop-blur">
    <label v-for="genre in options" :key="genre.id" class="mr-3 inline-flex items-center gap-1.5 py-1 text-xs">
      <input
        type="checkbox"
        class="check-control"
        :value="genre.name"
        :checked="modelValue.includes(genre.name)"
        @change="toggleGenre(genre.name, ($event.target as HTMLInputElement).checked)"
      />
      <span>{{ genre.name }}</span>
    </label>
    <span v-if="!options.length" class="text-xs text-black/50">暂无可选类型</span>
  </div>
</template>
