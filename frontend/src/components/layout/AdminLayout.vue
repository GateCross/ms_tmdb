<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AdminPreferencesDrawer from "@/components/layout/AdminPreferencesDrawer.vue";
import AdminSidebar from "@/components/layout/AdminSidebar.vue";
import AdminTabs from "@/components/layout/AdminTabs.vue";
import AdminTopbar from "@/components/layout/AdminTopbar.vue";
import { globalPageLoading } from "@/composables/useGlobalPageLoading";
import { buildSearchQuery, readQueryString } from "@/utils/routeSearch";
import { themeOptions, type AdminMenuGroup, type AdminMenuItem } from "./adminLayoutConfig";
import { useAdminPreferences } from "./useAdminPreferences";
import { useAdminTabs } from "./useAdminTabs";

const route = useRoute();
const router = useRouter();
const sidebarOpen = ref(false);
const preferencesOpen = ref(false);
const showBackToTop = ref(false);
const topbarSearchQuery = ref("");

const { adminThemeStyle, currentThemeOption, preferences, resetPreferences, setPreference } = useAdminPreferences();

const menuItems = computed<AdminMenuItem[]>(() =>
  router
    .getRoutes()
    .filter((item) => item.meta.title && !item.meta.hideMenu && !item.redirect)
    .map((item) => ({
      order: Number(item.meta.order ?? 100),
      path: item.path,
      section: String(item.meta.section ?? "系统"),
      title: String(item.meta.menuTitle ?? item.meta.title),
    }))
    .sort((left, right) => left.order - right.order),
);

const menuGroups = computed<AdminMenuGroup[]>(() => {
  const groups = new Map<string, AdminMenuItem[]>();
  for (const item of menuItems.value) {
    const groupItems = groups.get(item.section) ?? [];
    groupItems.push(item);
    groups.set(item.section, groupItems);
  }
  return Array.from(groups.entries()).map(([section, items]) => ({ section, items }));
});

const activeMenuPath = computed(() => String(route.meta.activeMenu ?? route.path));
const currentMenu = computed(() => menuItems.value.find((item) => item.path === activeMenuPath.value));
const hasRouteTitle = computed(() => typeof route.meta.title === "string" && route.meta.title.length > 0);
const currentSection = computed(() =>
  currentMenu.value?.section ?? String(route.meta.section ?? (route.path === "/" ? "工作台" : "系统")),
);
const currentTitle = computed(() => String(route.meta.title ?? currentMenu.value?.title ?? "首页"));
/* 日志页、本地库表格视图使用固定工作区；卡片视图保持整页自然滚动 */
const isFixedWorkspacePage = computed(
  () => route.path === "/logs" || (route.path === "/library" && route.query.view === "table"),
);

const { closeAllTabs, closeLeftTabs, closeOtherTabs, closeRightTabs, closeTab, openedTabs } = useAdminTabs({
  currentSection,
  currentTitle,
  hasRouteTitle,
  onRouteChange: () => {
    sidebarOpen.value = false;
  },
  route,
  router,
});

function handleScroll() {
  showBackToTop.value = window.scrollY > 360;
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function reloadPage() {
  router.go(0);
}

function toggleSidebar() {
  if (window.innerWidth < 1024) {
    sidebarOpen.value = !sidebarOpen.value;
    return;
  }
  preferences.sidebarCollapsed = !preferences.sidebarCollapsed;
}

async function submitTopbarSearch() {
  const trimmedQuery = topbarSearchQuery.value.trim();
  if (!trimmedQuery) {
    await router.push("/");
    return;
  }

  await router.push({
    path: "/",
    query: buildSearchQuery("multi", trimmedQuery),
  });
}

watch(
  () => route.query.q,
  (value) => {
    topbarSearchQuery.value = readQueryString(value);
  },
  { immediate: true },
);

onMounted(() => {
  handleScroll();
  window.addEventListener("scroll", handleScroll, { passive: true });
});

onBeforeUnmount(() => {
  window.removeEventListener("scroll", handleScroll);
});
</script>

<template>
  <div
    class="admin-app"
    :class="{ 'admin-app-sidebar-collapsed': preferences.sidebarCollapsed }"
    :style="adminThemeStyle"
  >
    <AdminSidebar :active-path="activeMenuPath" :groups="menuGroups" :open="sidebarOpen" />

    <button
      v-if="sidebarOpen"
      class="admin-sidebar-mask"
      type="button"
      aria-label="关闭导航"
      @click="sidebarOpen = false"
    ></button>

    <section class="admin-workspace" :class="{ 'admin-workspace-fixed': isFixedWorkspacePage }">
      <AdminTopbar
        v-model:search-query="topbarSearchQuery"
        :current-section="currentSection"
        :current-title="currentTitle"
        :sidebar-open="sidebarOpen"
        @open-preferences="preferencesOpen = true"
        @reload-page="reloadPage"
        @submit-search="submitTopbarSearch"
        @toggle-sidebar="toggleSidebar"
      />

      <AdminTabs
        v-if="preferences.showTabs"
        :current-full-path="route.fullPath"
        :tabs="openedTabs"
        @close="closeTab"
        @close-all="closeAllTabs"
        @close-left="closeLeftTabs"
        @close-other="closeOtherTabs"
        @close-right="closeRightTabs"
      />

      <main
        id="admin-main"
        tabindex="-1"
        class="page-shell admin-content"
        :class="{ 'admin-content-compact': preferences.compact, 'admin-content-fixed': isFixedWorkspacePage }"
      >
        <section class="vben-page">
          <div class="vben-page-content">
            <RouterView />
          </div>
        </section>
      </main>
    </section>

    <AdminPreferencesDrawer
      :current-theme-option="currentThemeOption"
      :preferences="preferences"
      :theme-options="themeOptions"
      :visible="preferencesOpen"
      @close="preferencesOpen = false"
      @reset="resetPreferences"
      @update-preference="setPreference"
    />

    <button v-if="showBackToTop" class="back-top-btn" type="button" aria-label="返回顶部" @click="scrollToTop"></button>

    <Transition name="admin-global-loading-fade">
      <div v-if="globalPageLoading" class="admin-global-loading" role="status" aria-live="polite" aria-label="页面加载中">
        <div class="admin-global-loading-loader" aria-hidden="true"></div>
      </div>
    </Transition>
  </div>
</template>
