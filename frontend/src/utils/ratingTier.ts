// 评分分档配色：≥7 绿、≥5 黄、其余红；空值与 0 分走中性样式。
export function ratingTierClass(voteAverage: number | null | undefined): string {
  if (typeof voteAverage !== "number" || voteAverage <= 0) return "";
  if (voteAverage >= 7) return "rating-badge--high";
  if (voteAverage >= 5) return "rating-badge--mid";
  return "rating-badge--low";
}
