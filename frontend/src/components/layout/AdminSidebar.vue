<script setup lang="ts">
import { Film } from "lucide-vue-next";
import { getMenuIconPaths, type AdminMenuGroup } from "./adminLayoutConfig";

defineProps<{
  activePath: string;
  groups: AdminMenuGroup[];
  open: boolean;
}>();
</script>

<template>
  <aside class="admin-sidebar" :class="{ 'admin-sidebar-open': open }" aria-label="主导航">
    <RouterLink to="/" class="admin-brand">
      <span class="admin-brand-mark" aria-hidden="true">
        <Film class="h-4 w-4" />
      </span>
      <span class="admin-brand-copy">
        <strong>媒体数据管理</strong>
      </span>
    </RouterLink>

    <nav class="admin-menu">
      <section v-for="group in groups" :key="group.section" class="admin-menu-section">
        <p class="admin-menu-title">{{ group.section }}</p>
        <RouterLink
          v-for="item in group.items"
          :key="item.path"
          :to="item.path"
          class="admin-menu-link"
          :class="{ 'admin-menu-link-active': item.path === activePath }"
        >
          <span class="admin-menu-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" focusable="false">
              <path v-for="pathData in getMenuIconPaths(item.path)" :key="pathData" :d="pathData" pathLength="24" />
            </svg>
          </span>
          <span>{{ item.title }}</span>
        </RouterLink>
      </section>
    </nav>
  </aside>
</template>
