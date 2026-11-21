const MEDIA_FIELD_LABEL_MAP: Record<string, string> = {
  title: "片名",
  original_title: "原始片名",
  name: "名称",
  original_name: "原始名称",
  overview: "简介",
  tagline: "标语",
  release_date: "上映日期",
  first_air_date: "首播日期",
  status: "状态",
  runtime: "时长",
  homepage: "主页链接",
  original_language: "原始语言",
  poster_path: "海报路径",
  backdrop_path: "背景图路径",
  vote_average: "评分",
  popularity: "热度",
  genre_names: "类型",
  genres: "类型",
  number_of_seasons: "季数",
  number_of_episodes: "集数",
  type: "剧集类型",
  seasons: "季信息",
  biography: "人物简介",
  profile_path: "头像路径",
  birthday: "生日",
  place_of_birth: "出生地",
  known_for_department: "擅长领域",
};

/**
 * 将后端/TMDB 媒体字段名映射为中文显示名称
 */
export function resolveMediaFieldLabel(field: string): string {
  const key = String(field ?? "").trim();
  if (!key) return "";
  return MEDIA_FIELD_LABEL_MAP[key] ?? key;
}

/**
 * 将媒体字段名列表格式化为中文摘要文案
 */
export function summarizeMediaFields(fields: string[], maxPreview = 6): string {
  if (!fields.length) return "无";
  const labels = fields.slice(0, maxPreview).map((field) => resolveMediaFieldLabel(field));
  if (fields.length > maxPreview) {
    return `${labels.join("、")} 等 ${fields.length} 项`;
  }
  return `${labels.join("、")}（共 ${fields.length} 项）`;
}
