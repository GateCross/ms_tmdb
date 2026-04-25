# MS-TMDB API 参考文档（当前实现）

更新时间：2026-02-24

## 基本信息

- 开发环境地址：`http://localhost:8888`
- Docker Compose 地址：`http://localhost:8080`
- TMDB 代理前缀：`/api/v3`
- 管理接口前缀：`/api/admin`
- 响应格式：`application/json`

说明：

1. 客户端无需传 `api_key`，服务端会自动附加 TMDB Key。
2. 常用查询参数（如 `language`、`page`、`region`、`append_to_response`）会按接口定义透传。
3. 实际可用路由以 `backend/api/tmdb.api` 为准。

## 一、TMDB 代理接口（`/api/v3`）

### 1.1 带本地读穿缓存的详情接口

- `GET /api/v3/movie/{id}`
- `GET /api/v3/tv/{id}`
- `GET /api/v3/person/{id}`

行为：

1. 本地命中且未过期时优先返回本地缓存。
2. 未命中或过期时回源 TMDB，成功后写回本地。
3. 回源失败且存在本地缓存时，可返回本地缓存作为降级结果。

### 1.2 其他已接入代理接口（节选）

电影：

- `/api/v3/movie/{id}/credits`
- `/api/v3/movie/{id}/images`
- `/api/v3/movie/{id}/videos`
- `/api/v3/movie/popular`
- `/api/v3/movie/top_rated`

电视剧：

- `/api/v3/tv/{id}/credits`
- `/api/v3/tv/{id}/images`
- `/api/v3/tv/{id}/videos`
- `/api/v3/tv/popular`
- `/api/v3/tv/top_rated`
- `/api/v3/tv/{series_id}/season/{season_number}`
- `/api/v3/tv/{series_id}/season/{season_number}/episode/{episode_number}`

其中：

- `GET /api/v3/tv/{series_id}/season/{season_number}` 的响应包含 `episodes` 数组，可直接用于“分集明细”展示与第三方系统拉取。

人物：

- `/api/v3/person/{id}/movie_credits`
- `/api/v3/person/{id}/tv_credits`
- `/api/v3/person/popular`

搜索与发现：

- `/api/v3/search/movie`
- `/api/v3/search/tv`
- `/api/v3/search/person`
- `/api/v3/search/multi`
- `/api/v3/discover/movie`
- `/api/v3/discover/tv`
- `/api/v3/trending/{media_type}/{time_window}`

其他：

- `/api/v3/genre/movie/list`
- `/api/v3/genre/tv/list`
- `/api/v3/configuration`
- `/api/v3/find/{external_id}`

## 二、管理接口（`/api/admin`）

### 2.1 本地覆盖编辑

- `PUT /api/admin/movie/{id}`
- `PUT /api/admin/tv/{id}`
- `GET /api/admin/tv/{id}/season/{season_number}/local`（查询某季是否已保存本地）
- `POST /api/admin/tv/{id}/season/{season_number}/local`（拉取 TMDB 季明细并保存/覆盖本地）
- `PUT /api/admin/tv/{id}/season/{season_number}/local`（更新本地季明细，`payload` 为季详情对象）

### 2.2 同步 TMDB 数据

- `POST /api/admin/sync/movie/{id}`
- `POST /api/admin/sync/tv/{id}`
- `POST /api/admin/sync/person/{id}`

请求体字段：

- `mode`：可选，常用值 `preview` / `overwrite_all` / `update_unmodified` / `selective`
- `overwrite_fields`：仅在 `selective` 模式下使用

### 2.3 清理本地覆盖

- `DELETE /api/admin/movie/{id}/local`
- `DELETE /api/admin/tv/{id}/local`

### 2.4 列表

- `GET /api/admin/movies?page=1&page_size=20&keyword=&search_mode=contains`
- `GET /api/admin/tv-series?page=1&page_size=20&keyword=&search_mode=contains`

说明：当 `keyword` 为纯数字时，列表接口会额外按 `tmdb_id` 精确匹配。

### 2.5 代理设置

- `GET /api/admin/proxy`
- `PUT /api/admin/proxy`

### 2.6 远程差异检测

- `GET /api/admin/compare/movie/{id}`
- `GET /api/admin/compare/tv/{id}`
- `GET /api/admin/compare/person/{id}`

## 三、调用示例

读取电影详情：

```bash
curl "http://localhost:8888/api/v3/movie/550?language=zh-CN&append_to_response=credits,images"
```

对比后按策略同步：

```bash
curl "http://localhost:8888/api/admin/compare/movie/550"

curl -X POST "http://localhost:8888/api/admin/sync/movie/550" \
  -H "Content-Type: application/json" \
  -d "{\"mode\":\"update_unmodified\"}"
```

读取剧集某季分集明细（第三方调用）：

```bash
curl "http://localhost:8888/api/v3/tv/279446/season/1?language=zh-CN&append_to_response=credits,images,videos"
```

保存并覆盖本地季明细：

```bash
curl -X POST "http://localhost:8888/api/admin/tv/279446/season/1/local?language=zh-CN"
```

## 四、维护建议

接口新增或变更后，优先同步：

1. `backend/api/tmdb.api`
2. `docs/api-reference.md`
3. `README.md`（仅保留面向使用者的最小说明）
