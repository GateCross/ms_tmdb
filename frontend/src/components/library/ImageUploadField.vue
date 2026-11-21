<script setup lang="ts">
import { Upload } from "lucide-vue-next";
import { tmdbImg } from "@/api/tmdb";

withDefaults(
  defineProps<{
    modelValue: string;
    label: string;
    uploading: boolean;
    editable?: boolean;
    placeholder?: string;
    previewAspect?: "poster" | "backdrop";
  }>(),
  {
    editable: false,
    placeholder: "上传后自动填充",
    previewAspect: "poster",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
  upload: [event: Event];
}>();
</script>

<template>
  <div class="text-xs text-muted">
    <div class="flex items-center justify-between gap-2">
      <span>{{ label }}</span>
      <label
        class="btn-soft-xs inline-flex cursor-pointer items-center gap-1"
        :class="uploading ? 'pointer-events-none opacity-60' : ''"
      >
        <Upload class="h-3.5 w-3.5" aria-hidden="true" />
        <span>{{ uploading ? "上传中..." : "上传图片" }}</span>
        <input class="sr-only" type="file" accept="image/*" :disabled="uploading" @change="emit('upload', $event)" />
      </label>
    </div>

    <div class="mt-1.5 flex items-center gap-2.5">
      <img
        v-if="modelValue.trim()"
        :src="tmdbImg(modelValue.trim(), 'w154')"
        :alt="label"
        class="shrink-0 rounded border border-line bg-raised object-cover"
        :class="previewAspect === 'backdrop' ? 'h-10 w-16' : 'h-12 w-8'"
      />
      <input
        :value="modelValue"
        :readonly="!editable"
        class="field-control w-full text-sm"
        :class="!editable ? 'opacity-80' : ''"
        :placeholder="placeholder"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
    </div>
  </div>
</template>
