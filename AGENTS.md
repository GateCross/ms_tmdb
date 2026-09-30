# MS-TMDB 项目规范（AI 编码必读）

媒体数据管理后台：TMDB 代理 + 本地缓存 + 管理界面。`backend/` 为 Go（go-zero 风格，module `ms_tmdb`），`frontend/` 为 Vue 3 前端。后端接口以 `docs/api-reference.md` 为准（`/api/v3` TMDB 代理，`/api/admin` 管理接口，`/uploads` 静态图片）。

改任何代码前先读完本文件；与本文件冲突的写法一律以本文件为准。

## 一、技术栈（frontend/，不得随意引入新依赖）

| 类别 | 选型 | 备注 |
| --- | --- | --- |
| 框架 | Vue 3.5 + TypeScript 5.8（`strict: true`） | 全部 `<script setup lang="ts">`，无 Options API |
| 构建 | Vite 6 + @vitejs/plugin-vue | dev 监听 `0.0.0.0:5173`（局域网访问，勿改回 localhost） |
| 路由 | Vue Router 4（history 模式） | 所有页面路由懒加载 `() => import(...)` |
| UI 原语 | radix-vue ^1.9 | 仅用于弹层/下拉等无头原语，无组件库皮肤 |
| 图标 | lucide-vue-next | 具名按需导入，尺寸用 Tailwind class（如 `h-4 w-4`） |
| 样式 | TailwindCSS 3.4（PostCSS 方式）+ tailwindcss-animate | `darkMode: 'class'` 已启用：`.dark` 类由 `useAdminPreferences.applyRootTheme` 按界面模式同步到 `html` 根（挂在根上才能覆盖 Teleport 到 body 的弹层）；深浅两套颜色令牌集中在 `styles/theme.css` 的 `:root`（浅色默认）与 `html.dark` |
| 请求 | axios 1.8 单例 | 见 `src/api/http.ts` |
| 工具 | clsx + tailwind-merge → `cn()`（`src/lib/utils.ts`） | 合并 class 一律用 `cn()` |
| 包管理 | pnpm 11.5.1（`packageManager` 钉死，CI 同版本） | 不要用 npm/yarn |

明确**没有**的东西，不要引入：状态库（Pinia/Vuex）、表单库（vee-validate 等）、UI 组件库（Element/Naive/Ant）、运行时校验（Zod）、测试框架、chart 库。需要新依赖先说明理由。

## 二、命令

```bash
pnpm -C frontend install     # 安装依赖（CI 用 --frozen-lockfile）
pnpm -C frontend dev         # 开发，0.0.0.0:5173，代理 /api /v3 /3 /uploads → localhost:8888
pnpm -C frontend typecheck   # vue-tsc --noEmit，交付前必须通过
pnpm -C frontend lint        # eslint
pnpm -C frontend format:check
pnpm -C frontend build       # 先 vue-tsc 再 vite build
```

没有测试框架，验证手段 = typecheck + lint + 人工过一遍相关页面。

## 三、目录与分层

```
frontend/src/
  api/        # 唯一的请求层，按资源域一域一文件
  composables/# useXxx 页面级逻辑（可含响应式状态）
  components/ # 按业务域分子目录：common/ layout/ library/ logs/ movie/ tv/
  pages/      # XxxPage.vue，只做组装，业务逻辑放 composables
  router/     # 路由表 + 全局加载/标题
  constants/  # 中文 label/英文 value 的选项常量（mediaStatus.ts）
  types/      # 跨域共享领域类型（media.ts）
  utils/      # 纯函数
  lib/        # cn()
  styles/     # 手写 CSS（见第五节）
```

数据流固定为：`pages` → `composables/useXxx` → `api/*` → `http.ts`。组件不直接 import axios。

## 四、编码细则

### API 层（src/api/）
- 统一走 `http.ts` 的 axios 单例（baseURL `/`、timeout 15s）。响应拦截器负责把错误转成中文提示并弹全局 Toast；`RequestOptions.showErrorToast: false` 用于预取/辅助请求静默。
- 函数命名：`getXxx`（详情）、`listXxx`（分页列表）、`createXxx/updateXxx/deleteXxx`、`syncXxx`、`compareXxxRemote`、`clearXxxCache`。
- 读接口需要缓存时用 `withRequestCache(key, loader, ttl)`（`requestCache.ts`：LRU 200 + inflight 去重 + 代次防旧请求回填）；detail 类支持 `force` 跳缓存（先 `clearRequestCache(key)` 再请求）。
- **写操作成功后必须调用对应 `clearMovieCache/clearTVCache/clearPersonCache(id)`** 失效读缓存（参考 `useLibraryList.confirmDeleteItem`）。
- 后端响应字段一律 snake_case，前端类型照抄不改驼峰；类型定义集中在各 api 文件（Admin 系）或 `types/media.ts`（媒体域）；组件域类型放 `components/<域>/types.ts`。
- 后端给的数据不可信：列表/表单数据进 UI 前先过 `utils/mediaNormalizers.ts` 的 `normalizeXxx` 归一化。

### Composables（src/composables/）
- 命名 `useXxx`，具名导出；组件内一次性状态放这里，跨页共享的全局态用模块级单例 `reactive/ref`（如 `useGlobalToast`、`useGlobalPageLoading`——项目不用 Pinia）。
- `useGlobalToast` = 全局单例 Toast（拦截器也在用）；`useToastNotice` = 组件内局部 toast，必须在 `onBeforeUnmount` 清理定时器。
- 异步加载必须带竞态保护：模块内 `let reqSeq = 0`，响应回来 `if (seq !== reqSeq) return`（参考 `useLibraryList.loadData`）。
- 搜索输入必须防抖（350ms，watch keywordInput 模式）；`utils/schedule.ts` 的 `scheduleAfterPaint` 用于避开首帧的延迟任务。
- URL 状态（tab/页码/关键词等）通过 query 同步：读用 `utils/routeQuery.ts` 的 `readQueryString`，写用 `router.replace` + `isSameQuery` 防抖动。
- 区分首载失败 `loadError`（显示空态）与刷新失败 `refreshError`（保留旧数据 + 顶部重试条），错误文案用 `resolveErrorMessage(err, fallback)`；catch 里不重复弹错，写 `/* handled by global toast */`。

### 组件（src/components/）
- props 用类型式 `defineProps<{...}>` + `withDefaults`；emits 用类型式 `defineEmits<{ close: [] }>`；v-model 用 computed get/set（见 `GlassSelect.vue`）并同发 `update:modelValue` + `change`。
- class 拼接一律 `cn()`；外部 attrs 转发时用 `useAttrs` + `defineOptions({ inheritAttrs: false })`（参考 GlassSelect）。
- **弹层一律用 radix-vue 封装的基建，禁止手写 focus trap / scroll lock**（git bd24ab7 已迁移并删除手写实现）：
  - 对话框：`components/common/BaseDialog.vue`（皮肤壳 `ModalShell.vue`，variant `glass`/`vben`）
  - 抽屉：`BaseDrawer.vue`；下拉：`GlassSelect.vue`；局部 toast：`ToastNotice.vue`
  - radix 的 `Portal` 不随 open 卸载，必须 `<Portal v-if="open">` 手动门控，否则透明全屏容器残留挡点击。
  - busy（提交中）必须阻止 Escape/点遮罩关闭；初始焦点用 `[data-dialog-primary]` 标记在容器内查找。
- **Select 选项 value 禁止空字符串**（曾导致日志页崩溃），用 `"all"` 等哨兵值。
- z-index 层级约定：弹窗 overlay/root `z-[1300]`、Select 弹层 `z-[1400]`。
- 列表加载/空态用 `common/LoadState.vue`（loading/empty/slot 三态）；表格型列表壳复用 `common/DataListShell.vue`（本地库表格与日志列表已统一于此，新列表先考虑复用）。
- 图标用 lucide-vue-next 具名导入；小图标也可内联 SVG（如删除按钮）。
- movie/tv 下成对出现的组件（MovieLocalEditor vs TVLocalEditorCard、MovieRemoteDiffCard vs TVRemoteDiffCard）是刻意的对称拆分，各自配 `types.ts`；改一边时检查另一边是否需要同步。

### 样式（src/styles/）
- 设计令牌单一事实来源：`styles/theme.css` 的 `:root`（浅色默认值）+ `html.dark`（深色覆盖），包含表面/文字/状态色全套变量；主题色（`--accent` 系 5 个派生变量 + `--surface-active`/`--field-border-focus`）由 `useAdminPreferences` 按主题色预设写入 `html` 内联样式，`index.html` 头部有同逻辑的内联引导脚本（首帧前应用，防刷新闪屏，两处逻辑必须同步改）。
- 界面模式三态：`appearance: light / dark / auto`（跟随系统 `prefers-color-scheme`）；旧存储里 `themeColor: "dark"` 由 `normalizePreferences` 迁移为 `appearance: "dark"`。
- **新组件禁止写死浅色/深色颜色**：Tailwind 一律用语义色 token（`text-ink`/`text-muted`/`bg-card`/`bg-raised`/`border-line`/`text-success`/`bg-warn-soft`/`border-danger-line` 等，见 tailwind.config.cjs，全部指向 CSS 变量）；手写 CSS 一律引用 `--xxx` 变量；确需按明暗区分的样式用 `html.dark` 选择器就地写在同文件（参考 controls.css 的 global-toast 块），禁止新建按主题前缀覆盖的独立文件。
- `style.css` 只有 `@tailwind` 指令；CSS 引入顺序在 `main.ts` 固定（theme → layout → pages/* → controls → responsive），顺序即覆盖优先级，新增文件要考虑插入位置。
- 页面样式一页一文件：`styles/pages/home.css` / `media.css`（详情页）/ `library.css` / `settings.css`；通用控件类（`.btn-primary`、`.btn-soft`、`.card`、`.panel-glass`、`.field-control` 等）在 `styles/controls.css`，先复用再造。
- 组件内样式：优先 Tailwind 工具类；确需 scoped CSS 可用 `@apply`（参考 LoadState.vue）。
- Tailwind 语义色（tailwind.config.cjs）：`ink/muted/strong/page/card/raised/line` + `brand/success/warn/danger/info`（各带 `soft`/`line` 变体），全部指向 theme.css 令牌。
- 响应式：`styles/responsive.css` + Tailwind 断点。**海报网格列数断点与 `useLibraryList.currentGridColumnCount()` 必须一一对应**（640/768/1024/1280/1536/1920/2560 → 2/3/4/5/6/8/10/12 列，卡片视图每页行数固定 2），改任何一边要同步另一边。

### 路由（src/router/index.ts）
- 新页面：懒加载 + `meta` 填全（`title` 中文、`section` 分组、`order` 排序；侧栏隐藏用 `hideMenu`，详情页用 `activeMenu` 指向所属菜单）。侧栏菜单由路由 meta 自动生成，不要手写菜单数组。
- 全局加载条与 `document.title` 由 `beforeEach/afterEach` 统一处理，页面内不要重复实现。

### 其他
- 语言参数默认 `zh-CN`；TMDB 图片一律经 `api/tmdb.ts` 的 `tmdbImg/profileImg` 生成（空路径返回内置占位图），不要手拼 image.tmdb.org URL。
- 全局常量 `__APP_VERSION__`（vite define 注入），eslint globals 已注册。
- 代码注释用中文、直述克制，只解释代码看不出来的约束（本仓库现状即如此）。

## 五、历史踩坑红线（改坏过一次，别再犯）

1. dev server `host: true`（0.0.0.0）是需求，不是笔误。
2. 弹层/下拉必须走 radix-vue；`DialogPortal` 必须 `v-if="open"` 门控。
3. 下拉 value 不能是 `""`。
4. 搜索必须防抖；预取失败必须静默（`showErrorToast: false`），且不影响导航。
5. 本地新建/保存的 `genres` 由后端按 TMDB 官方 ID 还原，前端只传 `genre_names` 名称数组，不要自造数字 ID。
6. 全局错误提示只在拦截器弹一次；页面内不要堆内联错误弹窗（79366fe 已清理过一轮）。
7. 列表优先复用 `DataListShell`/`LoadState`/`LogsPagination`，别再各写一套分页和列表壳。

## 六、提交规范

Commit message 用中文，前缀：`feat:` / `fix:` / `refactor:` / `perf:` / `build:` / `chore:`，一行说清改动；重要取舍写进 body（参考 git log 中 bd24ab7、e650b71 的写法）。

## 七、改动自检清单

- [ ] `pnpm -C frontend typecheck` 通过
- [ ] `pnpm -C frontend lint` 无 error
- [ ] 新弹层/下拉用的是 BaseDialog/ModalShell/GlassSelect 而非手写
- [ ] 写操作后调用了对应 `clearXxxCache`
- [ ] 异步有竞态保护、搜索有防抖、错误文案中文
- [ ] 改了网格列数时同步了 `currentGridColumnCount` 与 CSS 断点
