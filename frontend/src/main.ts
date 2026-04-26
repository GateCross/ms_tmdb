import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./style.css";
import "./styles/theme.css";
import "./styles/layout.css";
import "./styles/pages/home.css";
import "./styles/pages/media.css";
import "./styles/pages/settings.css";
import "./styles/pages/library.css";
import "./styles/controls.css";
import "./styles/dark-theme-overrides.css";
import "./styles/responsive.css";

createApp(App).use(router).mount("#app");
