export type StatusOption = {
  label: string;
  value: string;
};

export const movieStatusOptions: StatusOption[] = [
  { label: "已上映", value: "Released" },
  { label: "后期制作", value: "Post Production" },
  { label: "制作中", value: "In Production" },
  { label: "计划中", value: "Planned" },
  { label: "已取消", value: "Canceled" },
  { label: "传闻中", value: "Rumored" },
];

export const tvStatusOptions: StatusOption[] = [
  { label: "连载中", value: "Returning Series" },
  { label: "已完结", value: "Ended" },
  { label: "已取消", value: "Canceled" },
  { label: "制作中", value: "In Production" },
  { label: "计划中", value: "Planned" },
  { label: "试播集", value: "Pilot" },
];

export const tvTypeOptions: StatusOption[] = [
  { label: "剧情剧", value: "Scripted" },
  { label: "迷你剧", value: "Miniseries" },
  { label: "真人秀", value: "Reality" },
  { label: "新闻节目", value: "News" },
  { label: "访谈节目", value: "Talk Show" },
  { label: "纪录片", value: "Documentary" },
];

const statusLabelMap = new Map<string, string>(
  [...movieStatusOptions, ...tvStatusOptions].map((item) => [item.value, item.label]),
);

export function formatStatusLabel(status: string | null | undefined): string {
  const raw = String(status ?? "").trim();
  if (!raw) return "未知";
  return statusLabelMap.get(raw) ?? raw;
}

const typeLabelMap = new Map<string, string>(tvTypeOptions.map((item) => [item.value, item.label]));

export function formatTvTypeLabel(tvType: string | null | undefined): string {
  const raw = String(tvType ?? "").trim();
  if (!raw) return "未知";
  return typeLabelMap.get(raw) ?? raw;
}
