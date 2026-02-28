import http from "./http";
import type { SearchResultItem } from "@/types/media";

export type SearchType = "movie" | "tv" | "person" | "multi";

type SearchResp = {
  page: number;
  total_pages?: number;
  total_results?: number;
  results: SearchResultItem[];
};

export function searchByType(type: SearchType, query: string, page = 1, language = "zh-CN") {
  return http.get<SearchResp>(`/api/v3/search/${type}`, {
    params: { query, page, language },
  });
}
