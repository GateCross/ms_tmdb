/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{vue,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#171717",
        haze: "#f8f6f2",
        sand: "#e7dcc8",
        coral: "#f47c64",
        pine: "#275a53",
      },
      boxShadow: {
        soft: "0 20px 45px rgba(23, 23, 23, 0.12)",
      },
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    prefix: "dui-",
    themes: [
      {
        msdark: {
          primary: "#01b4e4",
          secondary: "#90cea1",
          accent: "#f5c542",
          neutral: "#172033",
          "base-100": "#101a2d",
          "base-200": "#172033",
          "base-300": "#22304a",
          info: "#38bdf8",
          success: "#22c55e",
          warning: "#f59e0b",
          error: "#ef4444",
          "--rounded-box": "1rem",
          "--rounded-btn": "9999px",
          "--rounded-badge": "9999px",
          "--animation-btn": "0.18s",
          "--animation-input": "0.18s",
          "--btn-focus-scale": "0.98",
          "--border-btn": "1px",
          "--tab-border": "1px",
          "--tab-radius": "9999px",
        },
      },
      "night",
      "dim",
    ],
  },
};
