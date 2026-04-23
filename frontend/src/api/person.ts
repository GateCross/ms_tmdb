import http from "./http";
import { clearRequestCache, withRequestCache } from "./requestCache";

const AUXILIARY_CACHE_TTL = 10 * 60 * 1000;
const DETAIL_CACHE_TTL = 5 * 60 * 1000;

type DetailOptions = {
  force?: boolean;
};

export function getPopularPeople(page = 1, language = "zh-CN") {
  return http.get("/api/v3/person/popular", { params: { page, language } });
}

export function getPersonDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const params = append ? { language, append_to_response: append } : { language };
  const key = `person:detail:${id}:${language}:${append}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get(`/api/v3/person/${id}`, { params }),
    DETAIL_CACHE_TTL,
  );
}

export function getPersonCombinedCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const key = `person:combined_credits:${id}:${language}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get(`/api/v3/person/${id}/combined_credits`, { params: { language } }),
    AUXILIARY_CACHE_TTL,
  );
}

export function getPersonImages(id: number, options: DetailOptions = {}) {
  const key = `person:images:${id}`;
  if (options.force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get(`/api/v3/person/${id}/images`),
    AUXILIARY_CACHE_TTL,
  );
}
