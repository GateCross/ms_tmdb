export interface AdminMenuItem {
  order: number;
  path: string;
  section: string;
  title: string;
}

export interface AdminMenuGroup {
  items: AdminMenuItem[];
  section: string;
}

export interface AdminTab {
  fullPath: string;
  path: string;
  section: string;
  title: string;
}

export type AdminThemeColor = "teal" | "blue" | "green" | "amber" | "rose";
export type AdminAppearance = "light" | "dark" | "auto";

export interface AdminPreferences {
  appearance: AdminAppearance;
  compact: boolean;
  showTabs: boolean;
  sidebarCollapsed: boolean;
  themeColor: AdminThemeColor;
}

/**
 * 主题预设只描述「主题色」差异；深浅两套表面/文字/状态色全部收敛在 styles/theme.css
 * 的 :root 与 html.dark 令牌里，这里不再携带整套颜色值。
 */
export interface AdminThemeOption {
  /** accent 须为 6 位 hex：派生变量在尾部直接拼接两位 hex alpha */
  accent: string;
  accentSoft: string;
  accentStrong: string;
  label: string;
  value: AdminThemeColor;
}

export const defaultPreferences: AdminPreferences = {
  appearance: "light",
  compact: false,
  showTabs: true,
  sidebarCollapsed: false,
  themeColor: "blue",
};

export const themeOptions: AdminThemeOption[] = [
  {
    accent: "#0ea5a4",
    accentSoft: "#90cea1",
    accentStrong: "#0d7c8a",
    label: "青绿",
    value: "teal",
  },
  {
    accent: "#1677ff",
    accentSoft: "#69b1ff",
    accentStrong: "#0958d9",
    label: "Vben 蓝",
    value: "blue",
  },
  {
    accent: "#16a34a",
    accentSoft: "#86efac",
    accentStrong: "#15803d",
    label: "绿色",
    value: "green",
  },
  {
    accent: "#d97706",
    accentSoft: "#fcd34d",
    accentStrong: "#b45309",
    label: "琥珀",
    value: "amber",
  },
  {
    accent: "#e11d48",
    accentSoft: "#fda4af",
    accentStrong: "#be123c",
    label: "玫红",
    value: "rose",
  },
];

const fallbackMenuIconPaths = ["M4 5h16", "M4 12h16", "M4 19h16"];
const menuIconPaths: Record<string, string[]> = {
  "/": ["M3 10.5 12 3l9 7.5", "M5 10v10h14V10", "M9 20v-6h6v6"],
  "/library": [
    "M4 6c0-1.1 3.6-2 8-2s8 .9 8 2-3.6 2-8 2-8-.9-8-2Z",
    "M4 6v6c0 1.1 3.6 2 8 2s8-.9 8-2V6",
    "M4 12v6c0 1.1 3.6 2 8 2s8-.9 8-2v-6",
  ],
  "/logs": ["M7 4h10", "M7 9h10", "M7 14h6", "M5 20h14a2 2 0 0 0 2-2V3H3v15a2 2 0 0 0 2 2Z"],
  "/system-settings": [
    "M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z",
    "M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.07a2 2 0 0 1-2.83 2.83l-.07-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 1.55V21a2 2 0 0 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.88.34l-.07.06a2 2 0 1 1-2.83-2.83l.06-.07A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.55-1H3a2 2 0 0 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.88l-.06-.07a2 2 0 1 1 2.83-2.83l.07.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.55V3a2 2 0 0 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.88-.34l.07-.06a2 2 0 1 1 2.83 2.83l-.06.07A1.7 1.7 0 0 0 19.4 9c.36.63.99 1 1.55 1H21a2 2 0 0 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1Z",
  ],
};

export function getMenuIconPaths(path: string) {
  return menuIconPaths[path] ?? fallbackMenuIconPaths;
}

export function themeSwatchStyle(option: AdminThemeOption) {
  return {
    background: `linear-gradient(135deg, ${option.accent} 0%, ${option.accentStrong} 100%)`,
  };
}
