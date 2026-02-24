<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";

const navItems = [
  { to: "/", label: "首页" },
  { to: "/search", label: "搜索" },
  { to: "/library", label: "库" },
  { to: "/system-settings", label: "系统设置" },
];

const showBackToTop = ref(false);

function handleScroll() {
  showBackToTop.value = window.scrollY > 360;
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  handleScroll();
  window.addEventListener("scroll", handleScroll, { passive: true });
});

onBeforeUnmount(() => {
  window.removeEventListener("scroll", handleScroll);
});
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="page-shell app-header-inner">
        <div>
          <p class="text-xs uppercase tracking-[0.25em] text-pine/80">MS-TMDB</p>
          <h1 class="text-2xl font-bold text-ink md:text-3xl">媒体代理控制台</h1>
        </div>
        <nav class="flex items-center gap-2 rounded-full bg-white/80 p-1 shadow-soft">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="rounded-full px-4 py-2 text-sm text-ink transition hover:bg-sand/70"
            active-class="bg-pine text-white hover:bg-pine"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
      </div>
    </header>

    <main class="page-shell app-content">
      <RouterView />
    </main>

    <button
      v-if="showBackToTop"
      class="back-top-btn"
      @click="scrollToTop"
    >
      返回顶部
    </button>
  </div>
</template>
