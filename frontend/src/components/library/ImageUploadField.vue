<script setup lang="ts">
defineProps<{
  modelValue: string;
  label: string;
  uploading: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  upload: [event: Event];
}>();
</script>

<template>
  <label class="text-xs text-black/60">
    {{ label }}
    <input
      :value="modelValue"
      readonly
      class="field-control mt-1 w-full text-sm opacity-80"
      placeholder="上传后自动填充"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <input class="mt-2 block w-full text-xs" type="file" accept="image/*" @change="emit('upload', $event)" />
    <span v-if="uploading" class="mt-1 inline-block text-[11px] text-black/50">上传中...</span>
  </label>
</template>
