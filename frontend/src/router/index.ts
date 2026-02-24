import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: () => import("@/pages/HomePage.vue") },
    { path: "/search", component: () => import("@/pages/SearchPage.vue") },
    { path: "/movie/:id", component: () => import("@/pages/MoviePage.vue") },
    { path: "/tv/:id", component: () => import("@/pages/TVPage.vue") },
    { path: "/person/:id", component: () => import("@/pages/PersonPage.vue") },
    { path: "/library", component: () => import("@/pages/LibraryPage.vue") },
    { path: "/system-settings", component: () => import("@/pages/SystemSettingsPage.vue") },
    { path: "/proxy-settings", redirect: "/system-settings" },
    { path: "/auto-sync-settings", redirect: "/system-settings" },
    { path: "/:pathMatch(.*)*", component: () => import("@/pages/NotFoundPage.vue") },
  ],
});

export default router;
