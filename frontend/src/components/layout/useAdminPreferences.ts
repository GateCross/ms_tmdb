import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  defaultPreferences,
  themeOptions,
  type AdminAppearance,
  type AdminPreferences,
  type AdminThemeColor,
} from "./adminLayoutConfig";

const preferenceStorageKey = "ms_tmdb_vben_admin_preferences";
const darkMedia = window.matchMedia?.("(prefers-color-scheme: dark)");

function normalizeAppearance(value: unknown): AdminAppearance {
  return value === "dark" || value === "auto" ? value : "light";
}

function normalizePreferences(raw: unknown): AdminPreferences {
  const payload = raw && typeof raw === "object" ? (raw as Partial<Record<keyof AdminPreferences, unknown>>) : {};
  // 旧版本把「深色」作为一种主题色存储，迁移为 appearance=dark + 默认主题色
  const legacyDark = payload.themeColor === "dark";
  const themeColor = themeOptions.some((item) => item.value === payload.themeColor)
    ? (payload.themeColor as AdminThemeColor)
    : defaultPreferences.themeColor;
  return {
    appearance: legacyDark ? "dark" : normalizeAppearance(payload.appearance),
    compact: typeof payload.compact === "boolean" ? payload.compact : defaultPreferences.compact,
    showTabs: typeof payload.showTabs === "boolean" ? payload.showTabs : defaultPreferences.showTabs,
    sidebarCollapsed:
      typeof payload.sidebarCollapsed === "boolean" ? payload.sidebarCollapsed : defaultPreferences.sidebarCollapsed,
    themeColor,
  };
}

export function useAdminPreferences() {
  const preferences = reactive<AdminPreferences>({ ...defaultPreferences });
  const systemPrefersDark = ref(darkMedia?.matches ?? false);
  const currentThemeOption = computed(
    () => themeOptions.find((item) => item.value === preferences.themeColor) ?? themeOptions[0],
  );
  const isDark = computed(() => {
    if (preferences.appearance === "auto") return systemPrefersDark.value;
    return preferences.appearance === "dark";
  });

  /** 仅写主题色派生变量；表面/文字/侧栏等颜色由 theme.css 的 :root + html.dark 提供 */
  const adminThemeStyle = computed<Record<string, string>>(() => {
    const accent = currentThemeOption.value.accent;

    return {
      "--accent": accent,
      "--accent-active-bg": `${accent}24`,
      "--accent-active-border": `${accent}52`,
      "--accent-hover-border": `${accent}57`,
      "--accent-ring": `${accent}1f`,
      "--accent-soft": currentThemeOption.value.accentSoft,
      "--accent-strong": currentThemeOption.value.accentStrong,
      "--field-border-focus": `${accent}ad`,
      "--surface-active": isDark.value ? `${accent}33` : `${accent}24`,
    };
  });

  function applyRootTheme(style: Record<string, string>) {
    if (typeof document === "undefined") return;

    const root = document.documentElement;
    // dark: 变体与 theme.css 的 html.dark 令牌由根元素 .dark 类驱动；挂在根上，Teleport 到 body 的弹层也能命中
    root.classList.toggle("dark", isDark.value);
    for (const [property, value] of Object.entries(style)) {
      root.style.setProperty(property, value);
    }
  }

  function loadPreferences() {
    try {
      const raw = window.localStorage.getItem(preferenceStorageKey);
      if (!raw) return;
      Object.assign(preferences, normalizePreferences(JSON.parse(raw) as unknown));
    } catch {
      Object.assign(preferences, defaultPreferences);
    }
  }

  function savePreferences() {
    window.localStorage.setItem(preferenceStorageKey, JSON.stringify(preferences));
  }

  function resetPreferences() {
    Object.assign(preferences, defaultPreferences);
  }

  function setPreference(key: keyof AdminPreferences, value: AdminPreferences[keyof AdminPreferences]) {
    Object.assign(preferences, { [key]: value });
  }

  function handleSystemColorSchemeChange(event: MediaQueryListEvent) {
    systemPrefersDark.value = event.matches;
  }

  // isDark 单独监听：.dark 类不能只依赖 adminThemeStyle 变化触发，否则 auto 模式下若样式恰好不变会漏切换
  watch([adminThemeStyle, isDark], ([style]) => applyRootTheme(style), { immediate: true });
  watch(preferences, savePreferences);
  darkMedia?.addEventListener("change", handleSystemColorSchemeChange);
  onMounted(loadPreferences);

  return {
    adminThemeStyle,
    currentThemeOption,
    preferences,
    resetPreferences,
    setPreference,
  };
}
