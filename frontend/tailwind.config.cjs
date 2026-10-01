/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{vue,ts,tsx}"],
  // class 策略：dark: 变体由 html 根的 .dark 类驱动；挂在根上，Teleport 到 body 的弹层也能命中
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        // 语义色：全部指向 styles/theme.css 的主题令牌，随浅色/深色自动切换。
        // 新组件禁止使用 text-black/50、bg-white 等写死浅色的工具类，统一用这里的 token。
        ink: "var(--text-main)", // 主文字
        muted: "var(--text-muted)", // 次要文字
        strong: "var(--text-strong)", // 强调文字
        page: "var(--bg-main)", // 页面背景
        card: "var(--surface)", // 卡片/面板背景
        raised: "var(--surface-muted)", // 次级面板/嵌入块背景
        line: "var(--border-muted)", // 描边
        overlay: "var(--overlay-bg)", // 弹层/侧栏遮罩
        brand: {
          DEFAULT: "var(--accent)",
          soft: "var(--accent-soft)",
          strong: "var(--accent-strong)",
        },
        success: {
          DEFAULT: "var(--success-fg)",
          soft: "var(--success-bg)",
          line: "var(--success-border)",
        },
        warn: {
          DEFAULT: "var(--warn-fg)",
          soft: "var(--warn-bg)",
          line: "var(--warn-border)",
        },
        danger: {
          DEFAULT: "var(--danger-fg)",
          soft: "var(--danger-bg)",
          line: "var(--danger-border)",
        },
        info: {
          DEFAULT: "var(--info-fg)",
          soft: "var(--info-bg)",
          line: "var(--info-border)",
        },
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};
