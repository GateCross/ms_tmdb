<script setup lang="ts">
import { ref } from "vue";
import { TooltipArrow, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger } from "reka-ui";

withDefaults(
  defineProps<{
    content: string;
    side?: "top" | "right" | "bottom" | "left";
  }>(),
  { side: "top" },
);

// reka 的 Portal 不随 open 卸载，须手动门控，否则透明容器残留挡点击
const open = ref(false);
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <TooltipRoot v-model:open="open">
      <TooltipTrigger as-child>
        <slot />
      </TooltipTrigger>
      <TooltipPortal v-if="open && content">
        <TooltipContent :side="side" :side-offset="8" class="base-tooltip animate-in fade-in-0 zoom-in-95">
          {{ content }}
          <TooltipArrow class="base-tooltip-arrow" aria-hidden="true" />
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>
