<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";

const navItems = [
  { to: "/", label: "首页" },
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
  <div class="app-shell" data-theme="msdark">
    <header class="app-header">
      <div class="page-shell app-header-inner">
        <div>
          <h1 class="app-brand-title">TMDB管理台</h1>
        </div>
        <nav class="app-nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="app-nav-link"
            active-class="app-nav-link-active"
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
