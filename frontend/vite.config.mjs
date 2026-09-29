import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const packageJson = JSON.parse(readFileSync(path.resolve(__dirname, "package.json"), "utf-8"));
const appVersion = process.env.VITE_APP_VERSION || packageJson.version || "";

export default defineConfig({
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    // 监听 0.0.0.0，允许局域网设备通过本机 IP 访问开发服务
    host: true,
    port: 5173,
    proxy: {
      // 后端原生兼容多前缀，开发代理只做原样转发。
      "/api": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
      "/v3": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
      "/3": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
      "/uploads": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
    },
  },
});
