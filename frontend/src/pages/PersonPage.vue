<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { getPersonDetail } from "@/api/person";
import { profileImg, tmdbImg } from "@/api/tmdb";

const route = useRoute();
const loading = ref(false);
const error = ref("");
const detail = ref<any>(null);

const personId = computed(() => Number(route.params.id));

async function loadData() {
  if (!personId.value) {
    error.value = "无效人物 ID";
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const resp = await getPersonDetail(personId.value);
    detail.value = resp.data;
  } catch (err: any) {
    error.value = err.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

// 合并电影和电视出演，按热度排序取前12
function topCredits(d: any) {
  const mc = d?.combined_credits?.cast ?? [];
  return mc.sort((a: any, b: any) => (b.popularity ?? 0) - (a.popularity ?? 0)).slice(0, 12);
}

onMounted(loadData);
watch(personId, () => {
  void loadData();
});
</script>

<template>
  <p v-if="loading" class="card text-sm text-black/60">加载中...</p>
  <p v-else-if="error" class="card text-sm text-red-600">{{ error }}</p>

  <template v-else-if="detail">
    <section class="card">
      <div class="detail-layout">
        <!-- 头像 -->
        <div class="detail-poster">
          <img
            :src="profileImg(detail.profile_path, 'w342')"
            :alt="detail.name"
            class="w-full rounded-xl shadow-soft"
          />
        </div>

        <!-- 个人信息 -->
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

          <!-- 照片墙 -->
          <div v-if="detail.images?.profiles?.length" class="mt-6">
            <h3 class="mb-2 text-sm font-semibold">照片</h3>
            <div class="flex gap-2 overflow-x-auto pb-2">
              <img
                v-for="(img, idx) in detail.images.profiles.slice(0, 6)"
                :key="idx"
                :src="tmdbImg(img.file_path, 'w185')"
                :alt="`${detail.name} photo`"
                class="h-32 w-auto flex-shrink-0 rounded-lg object-cover"
                loading="lazy"
              />
            </div>
          </div>

          <!-- 代表作品 -->
          <div v-if="topCredits(detail).length" class="mt-6">
            <h3 class="mb-2 text-sm font-semibold">代表作品</h3>
            <div class="cast-grid">
              <div v-for="c in topCredits(detail)" :key="c.id + (c.media_type || '')" class="cast-card">
                <RouterLink :to="`/${c.media_type === 'tv' ? 'tv' : 'movie'}/${c.id}`">
                  <img
                    :src="tmdbImg(c.poster_path, 'w185')"
                    :alt="c.title || c.name"
                    class="cast-img"
                    loading="lazy"
                  />
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
