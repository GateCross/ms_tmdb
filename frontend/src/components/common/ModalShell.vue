<script setup lang="ts">
withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    maxWidthClass?: string;
    contentClass?: string;
  }>(),
  {
    maxWidthClass: "max-w-5xl",
    contentClass: "max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6",
  },
);

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-[1000] flex items-center justify-center p-3 sm:p-6">
    <div class="absolute inset-0 bg-black/60 backdrop-blur-[2px]" @click="emit('close')" />
    <section :class="['panel-glass relative z-10 w-full overflow-hidden rounded-2xl', maxWidthClass]">
      <div
        class="sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/10 bg-black/35 px-4 py-3 backdrop-blur sm:px-6"
      >
        <h3 class="text-sm font-semibold">{{ title }}</h3>
        <button class="btn-soft px-3 py-1.5 text-xs" @click="emit('close')">关闭</button>
      </div>

      <div :class="contentClass">
        <slot />
      </div>
    </section>
  </div>
</template>
