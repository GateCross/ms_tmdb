# MS-TMDB

基于 `go-zero + PostgreSQL + Vue 3` 的 TMDB 代理与本地增强平台。

项目目标：
- 统一代理 TMDB v3 API（`/api/v3/*`）。
- 对电影/剧集/人物详情做本地 Read-Through 缓存。
- 提供管理接口与前端页面，支持本地编辑覆盖与库检索。

## 功能概览

- TMDB API 代理：电影、剧集、人物、搜索、发现、趋势、类型等。
- 本地缓存与回源：详情请求优先走本地，过期后自动回源更新。
- 管理接口：手动同步、编辑本地数据、清理本地覆盖、统计。
- 前端能力：
  - 详情页查看/编辑模式切换（电影、剧集）。
  - 本地字段保存（含类型多选）。
  - Library 卡片/表格双视图与双模式模糊搜索。

## 项目结构

```text
ms_tmdb/
├─ backend/      # go-zero 后端
├─ frontend/     # Vue 3 前端
├─ docker/       # 运行镜像与部署编排
└─ docs/         # 方案、接口、迭代文档
```

## 快速开始

### 1. 配置后端数据库与 TMDB

```bash
cp backend/etc/tmdb.example.yaml backend/etc/tmdb.yaml
# 再按需修改数据库地址、账号和 Tmdb.ApiKey
```

### 2. 开发环境（命令运行）

#### 启动后端

```bash
cd backend
go run tmdb.go -f etc/tmdb.yaml
```

默认监听：`http://localhost:8888`

#### 启动前端

```bash
cd frontend
pnpm install
pnpm dev
```

默认访问：`http://localhost:5173`

### 3. 生产环境（Docker Compose）

将 `docker` 目录下的 `docker-compose.yml` 和 `tmdb.yaml` 复制到服务器同一目录（可直接复制文件内容创建）。

先修改 `tmdb.yaml` 中的数据库连接与 `Tmdb.ApiKey`，再在该目录启动：

```bash
docker compose up -d
```

默认访问：`http://localhost:8080`

## 常用入口

- 前端页面：
  - `/` 首页
  - `/library` 本地库
  - `/movie/:id` 电影详情
  - `/tv/:id` 剧集详情
- 后端接口：
  - `/api/v3/*` TMDB 代理接口
  - `/api/admin/*` 本地管理接口

## 其他程序调用（API）

可通过 HTTP 直接调用本服务，适用于脚本、后端服务、工作流平台（如 n8n）等。

- 默认服务地址：`http://localhost:8888`
- 返回：`application/json`
- 鉴权：当前默认无需额外 Token（如后续接入鉴权，以部署配置为准）

### 1. 读取详情（带本地读穿缓存）

```bash
curl "http://localhost:8888/api/v3/movie/550?language=zh-CN&append_to_response=credits,images"
```

说明：
- 无需传 `api_key`，服务端会自动附加 TMDB Key。
- `/api/v3/movie/{id}`、`/api/v3/tv/{id}`、`/api/v3/person/{id}` 都支持同类调用。

### 2. 对比远程与本地差异（推荐先调用）

```bash
curl "http://localhost:8888/api/admin/compare/movie/550"
```

典型响应：

```json
{
  "has_diff": true,
  "diff_fields": ["vote_average", "vote_count"],
  "message": "检测到远程数据差异"
}
```

### 3. 执行同步（按模式）

```bash
# 仅预览变化，不落库
curl -X POST "http://localhost:8888/api/admin/sync/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"mode\":\"preview\"}"

# 全量覆盖本地
curl -X POST "http://localhost:8888/api/admin/sync/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"mode\":\"overwrite_all\"}"

# 仅更新未被本地修改的字段
curl -X POST "http://localhost:8888/api/admin/sync/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"mode\":\"update_unmodified\"}"

# 选择性覆盖指定字段
curl -X POST "http://localhost:8888/api/admin/sync/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"mode\":\"selective\",\"overwrite_fields\":[\"vote_average\",\"vote_count\"]}"
```

支持的资源路径：
- 电影：`/api/admin/sync/movie/{id}`
- 剧集：`/api/admin/sync/tv/{id}`
- 人物：`/api/admin/sync/person/{id}`

### 4. 更新本地覆盖字段

```bash
curl -X PUT "http://localhost:8888/api/admin/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"搏击俱乐部\",\"vote_average\":8.8,\"genre_names\":[\"剧情\",\"惊悚\"]}"
```

### 5. 获取本地库列表

```bash
curl "http://localhost:8888/api/admin/movies?page=1&page_size=20&keyword=star&search_mode=contains"
curl "http://localhost:8888/api/admin/tv-series?page=1&page_size=20"
```

### 6. Python 调用示例

```python
import requests

base = "http://localhost:8888"
movie_id = 550

# 1) 差异检测
cmp_resp = requests.get(f"{base}/api/admin/compare/movie/{movie_id}", timeout=10).json()

# 2) 有差异再按策略同步
if cmp_resp.get("has_diff"):
    sync_resp = requests.post(
        f"{base}/api/admin/sync/movie/{movie_id}",
        json={"mode": "update_unmodified"},
        timeout=20,
    ).json()
    print(sync_resp.get("message"))
```

### 7. 第三方接入推荐流程

适用于定时任务、后端服务、工作流引擎（n8n/Node-RED）：

1. 先调用对比接口判断是否有差异  
   `GET /api/admin/compare/movie/{id}` 或 `GET /api/admin/compare/tv/{id}`
2. `has_diff=false` 时跳过同步，直接结束或记录“无需更新”
3. `has_diff=true` 时按策略选择同步模式  
   - 保守更新：`mode=update_unmodified`  
   - 全量覆盖：`mode=overwrite_all`  
   - 字段级覆盖：`mode=selective` + `overwrite_fields`
4. 调用同步接口执行更新  
   `POST /api/admin/sync/movie/{id}` 或 `POST /api/admin/sync/tv/{id}`
5. 最后读取详情或列表做结果确认  
   `GET /api/v3/movie/{id}`、`GET /api/v3/tv/{id}`、`GET /api/admin/movies`

## 配置说明

- 后端配置文件：`backend/etc/tmdb.yaml`
- 后端配置模板：`backend/etc/tmdb.example.yaml`
- Docker 运行配置：`docker/tmdb.yaml`
- 关键配置项：
  - `Postgres.Host / Postgres.Port / Postgres.User / Postgres.Password / Postgres.DBName / Postgres.SSLMode`
  - `Tmdb.ApiKey`
  - `Tmdb.BaseURL`
  - `Tmdb.DefaultLanguage`

建议将 `Tmdb.ApiKey` 替换为你自己的密钥后再部署。

## 文档

- `docs/api-reference.md`
