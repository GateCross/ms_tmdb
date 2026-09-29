import { computed, onMounted, reactive, watch } from "vue";
import {
  defaultPreferences,
  sidebarOptions,
  themeOptions,
  type AdminPreferences,
  type AdminSidebarColor,
  type AdminThemeColor,
} from "./adminLayoutConfig";

const preferenceStorageKey = "ms_tmdb_vben_admin_preferences";

function normalizePreferences(raw: unknown): AdminPreferences {
  const payload = raw && typeof raw === "object" ? (raw as Partial<Record<keyof AdminPreferences, unknown>>) : {};
  const themeColor = themeOptions.some((item) => item.value === payload.themeColor)
    ? (payload.themeColor as AdminThemeColor)
    : defaultPreferences.themeColor;
  const sidebarColor = sidebarOptions.some((item) => item.value === payload.sidebarColor)
    ? (payload.sidebarColor as AdminSidebarColor)
    : defaultPreferences.sidebarColor;
  return {
    compact: typeof payload.compact === "boolean" ? payload.compact : defaultPreferences.compact,
    showTabs: typeof payload.showTabs === "boolean" ? payload.showTabs : defaultPreferences.showTabs,
    sidebarCollapsed:
      typeof payload.sidebarCollapsed === "boolean" ? payload.sidebarCollapsed : defaultPreferences.sidebarCollapsed,
    sidebarColor,
    themeColor,
  };
}

export function useAdminPreferences() {
  const preferences = reactive<AdminPreferences>({ ...defaultPreferences });
  const currentThemeOption = computed(
    () => themeOptions.find((item) => item.value === preferences.themeColor) ?? themeOptions[0],
  );
  const currentSidebarOption = computed(
    () => sidebarOptions.find((item) => item.value === preferences.sidebarColor) ?? sidebarOptions[0],
  );
  const adminThemeStyle = computed<Record<string, string>>(() => {
    const isDarkTheme = currentThemeOption.value.colorScheme === "dark";
    const accent = currentThemeOption.value.accent;

    return {
      "--bg-main": currentThemeOption.value.bgMain,
      "--border-muted": currentThemeOption.value.borderMuted,
      "--accent": accent,
      "--accent-active-bg": `${accent}24`,
      "--accent-active-border": `${accent}52`,
      "--accent-hover-border": `${accent}57`,
      "--accent-ring": `${accent}1f`,
      "--accent-soft": currentThemeOption.value.accentSoft,
      "--accent-strong": currentThemeOption.value.accentStrong,
      "--control-bg": isDarkTheme ? "rgba(255, 255, 255, 0.06)" : "#f2f3f5",
      "--control-hover-bg": isDarkTheme ? "rgba(255, 255, 255, 0.1)" : "#f2f3f5",
      "--control-text": isDarkTheme ? currentThemeOption.value.textMain : "#1f2329",
      "--field-border-focus": currentThemeOption.value.fieldBorderFocus,
      "--field-bg": currentThemeOption.value.fieldBg,
      "--field-border": currentThemeOption.value.fieldBorder,
      "--glass-bg": currentThemeOption.value.glassBg,
      "--glass-bg-strong": currentThemeOption.value.glassBgStrong,
      "--glass-border": currentThemeOption.value.glassBorder,
      "--glass-shadow": currentThemeOption.value.glassShadow,
      "--glass-shadow-soft": currentThemeOption.value.glassShadowSoft,
      "--menu-bg": isDarkTheme ? currentThemeOption.value.surfaceMuted : "#ffffff",
      "--menu-hover-bg": isDarkTheme ? "rgba(255, 255, 255, 0.08)" : "#f2f4f7",
      "--overlay-bg": "rgba(0, 0, 0, 0.42)",
      "--scrollbar-thumb": isDarkTheme ? "rgba(148, 163, 184, 0.36)" : "rgba(100, 116, 139, 0.38)",
      "--scrollbar-thumb-hover": isDarkTheme ? "rgba(148, 163, 184, 0.56)" : "rgba(71, 85, 105, 0.58)",
      "--scrollbar-track": "transparent",
      "--sidebar-active-bg": currentSidebarOption.value.activeBg,
      "--sidebar-active-icon-bg": currentSidebarOption.value.activeIconBg,
      "--sidebar-active-icon-text": currentSidebarOption.value.activeIconText,
      "--sidebar-active-text": currentSidebarOption.value.activeText,
      "--sidebar-bg": currentSidebarOption.value.bg,
      "--sidebar-border": currentSidebarOption.value.border,
      "--sidebar-hover-bg": currentSidebarOption.value.hoverBg,
      "--sidebar-hover-icon-bg": currentSidebarOption.value.hoverIconBg,
      "--sidebar-hover-icon-text": currentSidebarOption.value.hoverIconText,
      "--sidebar-hover-text": currentSidebarOption.value.hoverText,
      "--sidebar-icon-bg": currentSidebarOption.value.iconBg,
      "--sidebar-muted": currentSidebarOption.value.muted,
      "--sidebar-text": currentSidebarOption.value.text,
      "--surface": currentThemeOption.value.surface,
      "--surface-active": isDarkTheme ? `${accent}33` : `${accent}24`,
      "--surface-elevated": isDarkTheme ? currentThemeOption.value.surfaceMuted : currentThemeOption.value.surface,
      "--surface-hover": isDarkTheme ? "rgba(255, 255, 255, 0.06)" : "#f7f8fa",
      "--surface-muted": currentThemeOption.value.surfaceMuted,
      "--surface-strong": currentThemeOption.value.surfaceStrong,
      "--text-main": currentThemeOption.value.textMain,
      "--text-muted": currentThemeOption.value.textMuted,
      "--text-secondary": currentThemeOption.value.textMuted,
      "--text-strong": isDarkTheme ? currentThemeOption.value.textMain : "#1f2329",
      "--topbar-bg": currentThemeOption.value.topbarBg,
      "--primary": "212 100% 54%",
      "color-scheme": currentThemeOption.value.colorScheme,
    };
  });

  function applyRootTheme(style: Record<string, string>) {
    if (typeof document === "undefined") return;

    const root = document.documentElement;
    // dark: 变体由根元素 .dark 类驱动；挂在根上，Teleport 到 body 的弹层也能命中
    root.classList.toggle("dark", currentThemeOption.value.colorScheme === "dark");
    root.setAttribute("data-theme", currentThemeOption.value.dataTheme);
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

  watch(adminThemeStyle, applyRootTheme, { immediate: true });
  watch(preferences, savePreferences);
  onMounted(loadPreferences);

  return {
    adminThemeStyle,
    currentSidebarOption,
    currentThemeOption,
    preferences,
    resetPreferences,
    setPreference,
  };
}
