<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { AdminSyncMode, AdminSyncPayload, AdminSyncResp } from "@/api/admin";
import { syncMovie, syncPerson, syncTV } from "@/api/admin";

const props = defineProps<{
  mediaType: "movie" | "tv" | "person";
  targetId: number;
  allowedModes?: AdminSyncMode[];
  embedded?: boolean;
  presetChangedFields?: string[];
}>();

const emit = defineEmits<{
  (event: "synced"): void;
}>();

const syncMode = ref<AdminSyncMode>("update_unmodified");
const syncing = ref(false);
const diffChecking = ref(false);
const syncError = ref("");
const syncMessage = ref("");
const changedFields = ref<string[]>([]);
const selectedOverwriteFields = ref<string[]>([]);

const modeOptions: Array<{ label: string; value: AdminSyncMode; hint: string }> = [
  { label: "仅更新未变更字段", value: "update_unmodified", hint: "保留本地已修改字段，只更新其它字段" },
  { label: "全量覆盖数据库", value: "overwrite_all", hint: "使用 TMDB 最新数据覆盖本地全部字段" },
  { label: "按变化字段覆盖", value: "selective", hint: "先检测变化字段，再选择要覆盖的本地字段" },
];

const canSync = computed(() => Number.isFinite(props.targetId) && props.targetId > 0);
const usingPresetChangedFields = computed(() => {
  return Array.isArray(props.presetChangedFields) && props.presetChangedFields.length > 0;
});
const visibleModeOptions = computed(() => {
  const allowed = props.allowedModes;
  if (!Array.isArray(allowed) || allowed.length === 0) {
    return modeOptions;
  }
  return modeOptions.filter((option) => allowed.includes(option.value));
});
const panelClass = computed(() => {
  if (props.embedded) {
    return "mt-3 rounded-lg border border-amber-200 bg-gray-50 p-3";
  }
  return "mt-6 rounded-xl border border-black/10 bg-white/70 p-4";
});

function normalizePresetChangedFields(): string[] {
  if (!Array.isArray(props.presetChangedFields)) {
    return [];
  }
  const unique = new Set<string>();
  props.presetChangedFields.forEach((field) => {
    const name = String(field ?? "").trim();
    if (!name) {
      return;
    }
    unique.add(name);
  });
  return [...unique];
}

function applyPresetChangedFields() {
  const fields = normalizePresetChangedFields();
  changedFields.value = fields;
  selectedOverwriteFields.value = [...fields];
}

function resolveFieldLabel(field: string) {
  const map: Record<string, string> = {
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
    homepage: "主页",
    poster_path: "海报路径",
    backdrop_path: "背景图路径",
    vote_average: "评分",
    popularity: "热度",
    genre_names: "类型",
    genres: "类型",
    number_of_seasons: "季数",
    number_of_episodes: "集数",
    type: "剧集类型",
    biography: "人物简介",
    profile_path: "头像路径",
    birthday: "生日",
    place_of_birth: "出生地",
    known_for_department: "擅长领域",
  };
  return map[field] ?? field;
}

function resolveErrorMessage(err: unknown, fallback: string) {
  if (err && typeof err === "object" && "message" in err) {
    const message = (err as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }
  return fallback;
}

async function executeSync(payload: AdminSyncPayload) {
  const targetId = Number(props.targetId);
  if (!Number.isFinite(targetId) || targetId <= 0) {
    throw new Error("无效目标 ID");
  }
  if (props.mediaType === "movie") {
    return syncMovie(targetId, payload);
  }
  if (props.mediaType === "tv") {
    return syncTV(targetId, payload);
  }
  return syncPerson(targetId, payload);
}

async function loadChangedFields() {
  if (usingPresetChangedFields.value) {
    applyPresetChangedFields();
    syncMessage.value = `检测到 ${changedFields.value.length} 个有变化字段`;
    return;
  }
  if (!canSync.value || diffChecking.value) return;
  diffChecking.value = true;
  syncError.value = "";
  try {
    const resp = await executeSync({ mode: "preview" });
    const data = resp.data as AdminSyncResp;
    changedFields.value = Array.isArray(data.changed_fields) ? data.changed_fields : [];
    selectedOverwriteFields.value = [...changedFields.value];
    syncMessage.value = data.message || `检测到 ${changedFields.value.length} 个变化字段`;
  } catch (err: unknown) {
    syncError.value = resolveErrorMessage(err, "检测变化字段失败");
  } finally {
    diffChecking.value = false;
  }
}

async function applySync() {
  if (!canSync.value || syncing.value) return;
  syncing.value = true;
  syncError.value = "";
  syncMessage.value = "";
  try {
    if (syncMode.value === "selective" && changedFields.value.length === 0 && !usingPresetChangedFields.value) {
      await loadChangedFields();
    }

    const payload: AdminSyncPayload = { mode: syncMode.value };
    if (syncMode.value === "selective") {
      payload.overwrite_fields = selectedOverwriteFields.value;
    }
    const resp = await executeSync(payload);
    const data = resp.data as AdminSyncResp;
    changedFields.value = Array.isArray(data.changed_fields) ? data.changed_fields : [];
    selectedOverwriteFields.value = changedFields.value.filter((field) => !data.overwritten_fields?.includes(field));
    syncMessage.value = data.message || "同步完成";
    emit("synced");
  } catch (err: unknown) {
    syncError.value = resolveErrorMessage(err, "同步失败");
  } finally {
    syncing.value = false;
  }
}

watch(syncMode, (mode) => {
  syncError.value = "";
  if (mode !== "selective") {
    changedFields.value = [];
    selectedOverwriteFields.value = [];
    return;
  }
  if (usingPresetChangedFields.value) {
    applyPresetChangedFields();
  }
});

watch(
  () => props.presetChangedFields,
  () => {
    if (syncMode.value === "selective" && usingPresetChangedFields.value) {
      applyPresetChangedFields();
    }
  },
  { immediate: true, deep: true },
);

watch(
  visibleModeOptions,
  (options) => {
    if (!options.some((option) => option.value === syncMode.value)) {
      syncMode.value = options[0]?.value ?? "update_unmodified";
    }
  },
  { immediate: true },
);
</script>

<template>
  <div :class="panelClass">
    <h3 v-if="!props.embedded" class="text-sm font-semibold">数据库同步</h3>
    <p class="text-xs text-black/60" :class="{ 'mt-1': !props.embedded }">
      直接在详情页执行数据重拉取，不再需要进入管理页。
    </p>

    <div class="mt-3 grid gap-2">
      <label v-for="option in visibleModeOptions" :key="option.value" class="sync-option-card">
        <div class="flex items-center gap-2">
          <input v-model="syncMode" type="radio" class="radio-control" :value="option.value" />
          <span class="font-medium">{{ option.label }}</span>
        </div>
        <p class="mt-1 pl-5 text-xs text-black/55">{{ option.hint }}</p>
      </label>
    </div>

    <div v-if="syncMode === 'selective'" class="sync-selective-box">
      <div class="flex items-center gap-2">
        <button
          v-if="!usingPresetChangedFields"
          class="btn-soft-xs disabled:opacity-60"
          :disabled="diffChecking || syncing || !canSync"
          @click="loadChangedFields"
        >
          {{ diffChecking ? "检测中..." : "检测变化字段" }}
        </button>
        <span class="text-xs text-black/60">共 {{ changedFields.length }} 项</span>
      </div>
      <p v-if="usingPresetChangedFields" class="mt-2 text-xs text-black/55">
        已使用上方远程差异字段列表，可直接选择覆盖项。
      </p>

      <div class="mt-2 flex flex-wrap gap-2">
        <label v-for="field in changedFields" :key="field" class="field-choice-pill">
          <input v-model="selectedOverwriteFields" type="checkbox" class="check-control" :value="field" />
          <span>{{ resolveFieldLabel(field) }}</span>
        </label>
        <span v-if="!changedFields.length" class="text-xs text-black/50"> 暂未检测到变化字段 </span>
      </div>
    </div>

    <div class="mt-3 flex items-center gap-3">
      <button
        class="btn-primary disabled:opacity-60"
        :disabled="syncing || diffChecking || !canSync"
        @click="applySync"
      >
        {{ syncing ? "同步中..." : "执行同步" }}
      </button>
    </div>

    <div class="mt-2">
      <span v-if="syncMessage" class="text-xs text-green-700">{{ syncMessage }}</span>
      <span v-if="syncError" class="text-xs text-red-600">{{ syncError }}</span>
    </div>
  </div>
</template>
