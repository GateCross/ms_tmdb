<script setup lang="ts">
const navItems = [
  { to: "/", label: "首页" },
  { to: "/search", label: "搜索" },
  { to: "/library", label: "库" },
  { to: "/proxy-settings", label: "代理设置" },
];
</script>

<template>
  <div class="page-shell">
    <header class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <p class="text-xs uppercase tracking-[0.25em] text-pine/80">MS-TMDB</p>
        <h1 class="text-2xl font-bold text-ink md:text-3xl">媒体代理控制台</h1>
      </div>
      <nav class="flex items-center gap-2 rounded-full bg-white/70 p-1 shadow-soft">
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
    </header>

    <RouterView />
  </div>
</template>
