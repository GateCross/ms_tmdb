<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { prefetchMediaDetail } from "@/api/prefetch";
import { useRoute, useRouter } from "vue-router";
import { getPersonCombinedCredits, getPersonDetail, getPersonImages } from "@/api/person";
import { profileImg, tmdbImg } from "@/api/tmdb";
import { resolveErrorMessage } from "@/utils/errors";
import { scheduleAfterPaint } from "@/utils/schedule";

type PersonCreditItem = {
  id: number;
  media_type: string;
  title?: string;
  name?: string;
  character?: string;
  job?: string;
  poster_path?: string;
  popularity?: number;
};

type PersonImageItem = {
  file_path: string;
};

type PersonDetail = {
  name?: string;
  profile_path?: string;
  known_for_department?: string;
  birthday?: string;
  place_of_birth?: string;
  popularity?: number;
  biography?: string;
};

function toRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const error = ref("");
const detail = ref<PersonDetail | null>(null);
const topCreditItems = ref<PersonCreditItem[]>([]);
const photoProfiles = ref<PersonImageItem[]>([]);
const creditsLoading = ref(false);
const creditsLoaded = ref(false);
const creditsError = ref("");
const photosLoading = ref(false);
const photosLoaded = ref(false);
const photosError = ref("");
let detailReqSeq = 0;
let creditsReqSeq = 0;
let photosReqSeq = 0;
let cancelDeferredLoads: (() => void) | null = null;

const personId = computed(() => Number(route.params.id));
const sourceType = computed(() => {
  const value = String(route.query.fromType ?? "")
    .trim()
    .toLowerCase();
  if (value === "movie" || value === "tv") {
    return value;
  }
  return "";
});
const sourceId = computed(() => {
  const value = Number(route.query.fromId);
  return Number.isFinite(value) && value > 0 ? value : 0;
});
const topCredits = computed(() => {
  return [...topCreditItems.value].sort((a, b) => (b.popularity ?? 0) - (a.popularity ?? 0)).slice(0, 12);
});

function goBack() {
  if (sourceType.value && sourceId.value > 0) {
    void router.push(`/${sourceType.value}/${sourceId.value}`);
    return;
  }
  if (window.history.length > 1) {
    router.back();
    return;
  }
  void router.push("/");
}

function stopDeferredLoads() {
  if (cancelDeferredLoads) {
    cancelDeferredLoads();
    cancelDeferredLoads = null;
  }
}

function resetAuxState() {
  creditsReqSeq++;
  photosReqSeq++;
  topCreditItems.value = [];
  photoProfiles.value = [];
  creditsLoading.value = false;
  creditsLoaded.value = false;
  creditsError.value = "";
  photosLoading.value = false;
  photosLoaded.value = false;
  photosError.value = "";
}

function normalizeCreditItems(raw: unknown): PersonCreditItem[] {
  const cast = Array.isArray((raw as Record<string, unknown> | null)?.cast)
    ? ((raw as Record<string, unknown>).cast as unknown[])
    : [];
  return cast
    .map((item) => {
      const value = toRecord(item);
      return {
        id: Number(value.id) || 0,
        media_type: String(value.media_type ?? ""),
        title: typeof value.title === "string" ? value.title : undefined,
        name: typeof value.name === "string" ? value.name : undefined,
        character: typeof value.character === "string" ? value.character : undefined,
        job: typeof value.job === "string" ? value.job : undefined,
        poster_path: typeof value.poster_path === "string" ? value.poster_path : undefined,
        popularity: Number.isFinite(Number(value.popularity)) ? Number(value.popularity) : 0,
      };
    })
    .filter((item) => item.id > 0);
}

function normalizeImageItems(raw: unknown): PersonImageItem[] {
  const profiles = Array.isArray((raw as Record<string, unknown> | null)?.profiles)
    ? ((raw as Record<string, unknown>).profiles as unknown[])
    : [];
  return profiles
    .map((item) => ({
      file_path: String(toRecord(item).file_path ?? "").trim(),
    }))
    .filter((item) => item.file_path);
}

function prefetchCreditItem(item: PersonCreditItem) {
  const mediaType = item.media_type === "tv" ? "tv" : "movie";
  prefetchMediaDetail(mediaType, item.id);
}

function scheduleDeferredLoadsForDetail() {
  stopDeferredLoads();
  cancelDeferredLoads = scheduleAfterPaint(() => {
    void loadPersonCombinedCredits();
    void loadPersonImages();
  });
}

async function loadPersonCombinedCredits(force = false) {
  if (!personId.value || creditsLoading.value || (creditsLoaded.value && !force)) {
    return;
  }

  const requestSeq = ++creditsReqSeq;
  const targetId = personId.value;
  creditsLoading.value = true;
  creditsError.value = "";
  try {
    const resp = await getPersonCombinedCredits(targetId, "zh-CN", { force });
    if (requestSeq !== creditsReqSeq || targetId !== personId.value) {
      return;
    }
    topCreditItems.value = normalizeCreditItems(resp.data);
    creditsLoaded.value = true;
  } catch (err: unknown) {
    if (requestSeq !== creditsReqSeq || targetId !== personId.value) {
      return;
    }
    creditsError.value = resolveErrorMessage(err, "加载代表作品失败");
  } finally {
    if (requestSeq === creditsReqSeq) {
      creditsLoading.value = false;
    }
  }
}

async function loadPersonImages(force = false) {
  if (!personId.value || photosLoading.value || (photosLoaded.value && !force)) {
    return;
  }

  const requestSeq = ++photosReqSeq;
  const targetId = personId.value;
  photosLoading.value = true;
  photosError.value = "";
  try {
    const resp = await getPersonImages(targetId, { force });
    if (requestSeq !== photosReqSeq || targetId !== personId.value) {
      return;
    }
    photoProfiles.value = normalizeImageItems(resp.data).slice(0, 6);
    photosLoaded.value = true;
  } catch (err: unknown) {
    if (requestSeq !== photosReqSeq || targetId !== personId.value) {
      return;
    }
    photosError.value = resolveErrorMessage(err, "加载照片失败");
  } finally {
    if (requestSeq === photosReqSeq) {
      photosLoading.value = false;
    }
  }
}

async function loadData() {
  if (!personId.value) {
    error.value = "无效人物 ID";
    return;
  }

  const requestSeq = ++detailReqSeq;
  stopDeferredLoads();
  loading.value = true;
  error.value = "";
  resetAuxState();
  try {
    const resp = await getPersonDetail(personId.value);
    if (requestSeq !== detailReqSeq) {
      return;
    }
    detail.value = resp.data;
    scheduleDeferredLoadsForDetail();
  } catch (err: unknown) {
    if (requestSeq === detailReqSeq) {
      error.value = resolveErrorMessage(err, "加载失败");
    }
  } finally {
    if (requestSeq === detailReqSeq) {
      loading.value = false;
    }
  }
}

onMounted(loadData);
watch(personId, () => {
  void loadData();
});

onBeforeUnmount(() => {
  detailReqSeq++;
  creditsReqSeq++;
  photosReqSeq++;
  stopDeferredLoads();
});
</script>

<template>
  <p v-if="loading" class="card text-sm text-black/60">加载中...</p>
  <p v-else-if="error" class="card text-sm text-red-600">{{ error }}</p>

  <template v-else-if="detail">
    <section class="card">
      <div class="mb-4">
        <button class="btn-soft-xs px-3 py-1.5" @click="goBack">返回上一页</button>
      </div>
      <div class="detail-layout">
        <div class="detail-poster">
          <img :src="profileImg(detail.profile_path, 'w342')" :alt="detail.name" class="detail-poster-img" />
        </div>

        <div class="detail-info">
          <h1 class="text-2xl font-bold">{{ detail.name }}</h1>

          <div class="mt-3 flex flex-wrap gap-2">
            <span v-if="detail.known_for_department" class="badge">
              {{ detail.known_for_department }}
            </span>
            <span v-if="detail.birthday" class="badge">🎂 {{ detail.birthday }}</span>
            <span v-if="detail.place_of_birth" class="badge">📍 {{ detail.place_of_birth }}</span>
            <span class="badge">🔥 {{ detail.popularity?.toFixed(0) ?? "-" }}</span>
          </div>

          <p class="mt-4 text-sm leading-relaxed text-black/75">
            {{ detail.biography || "暂无简介" }}
          </p>

          <div class="content-auto mt-6">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-semibold">照片</h3>
              <button
                class="btn-soft-xs px-3 py-1 disabled:opacity-60"
                :disabled="photosLoading"
                @click="loadPersonImages(true)"
              >
                {{ photosLoading ? "加载中..." : photosLoaded ? "刷新照片" : "加载照片" }}
              </button>
            </div>
            <p v-if="photosLoading" class="text-xs text-black/55">正在加载照片...</p>
            <p v-else-if="photosError" class="text-xs text-red-600">{{ photosError }}</p>
            <p v-else-if="photosLoaded && !photoProfiles.length" class="text-xs text-black/55">暂无照片数据</p>
            <div v-else-if="photoProfiles.length" class="person-photo-strip">
              <img
                v-for="(img, idx) in photoProfiles"
                :key="idx"
                :src="tmdbImg(img.file_path, 'w185')"
                :alt="`${detail.name} photo`"
                class="person-photo-img"
                loading="lazy"
              />
            </div>
          </div>

          <div class="content-auto mt-6">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <h3 class="text-sm font-semibold">代表作品</h3>
              <button
                class="btn-soft-xs px-3 py-1 disabled:opacity-60"
                :disabled="creditsLoading"
                @click="loadPersonCombinedCredits(true)"
              >
                {{ creditsLoading ? "加载中..." : creditsLoaded ? "刷新作品" : "加载作品" }}
              </button>
            </div>
            <p v-if="creditsLoading" class="text-xs text-black/55">正在加载代表作品...</p>
            <p v-else-if="creditsError" class="text-xs text-red-600">{{ creditsError }}</p>
            <p v-else-if="creditsLoaded && !topCredits.length" class="text-xs text-black/55">暂无代表作品数据</p>
            <div v-else-if="topCredits.length" class="cast-grid">
              <div v-for="c in topCredits" :key="c.id + (c.media_type || '')" class="cast-card">
                <RouterLink
                  :to="`/${c.media_type === 'tv' ? 'tv' : 'movie'}/${c.id}`"
                  @mouseenter="prefetchCreditItem(c)"
                  @focus="prefetchCreditItem(c)"
                  @touchstart.passive="prefetchCreditItem(c)"
                >
                  <img :src="tmdbImg(c.poster_path, 'w185')" :alt="c.title || c.name" class="cast-img" loading="lazy" />
                </RouterLink>
                <p class="mt-1 truncate text-xs font-medium">{{ c.title || c.name }}</p>
                <p class="truncate text-xs text-black/50">{{ c.character ?? c.job ?? "" }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </template>
</template>
