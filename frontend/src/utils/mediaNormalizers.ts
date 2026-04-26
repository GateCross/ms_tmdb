export type GenreOption = {
  id: number;
  name: string;
};

export type CastMember = {
  id: number;
  name: string;
  character: string;
  profile_path: string;
};

export type MovieEditFormData = {
  tmdb_id: string;
  title: string;
  original_title: string;
  genre_names: string[];
  tagline: string;
  release_date: string;
  status: string;
  runtime: string;
  original_language: string;
  homepage: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

export type TVEditFormData = {
  tmdb_id: string;
  name: string;
  original_name: string;
  genre_names: string[];
  type: string;
  tagline: string;
  first_air_date: string;
  status: string;
  number_of_seasons: string;
  number_of_episodes: string;
  original_language: string;
  homepage: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

function toRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function toText(value: unknown): string {
  return value == null ? "" : String(value);
}

function normalizeGenreNames(raw: unknown): string[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => toText(toRecord(item).name).trim()).filter(Boolean);
}

export function normalizeGenreOptions(raw: unknown): GenreOption[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((item, idx) => {
      const record = toRecord(item);
      return {
        id: Number(record.id) || idx + 1,
        name: toText(record.name).trim(),
      };
    })
    .filter((item) => !!item.name);
}

export function normalizeCastMembers(raw: unknown, limit = 8): CastMember[] {
  const record = toRecord(raw);
  const cast = Array.isArray(record.cast) ? record.cast : [];
  return cast
    .map((item) => {
      const recordItem = toRecord(item);
      return {
        id: Number(recordItem.id) || 0,
        name: toText(recordItem.name).trim(),
        character: toText(recordItem.character).trim(),
        profile_path: toText(recordItem.profile_path),
      };
    })
    .filter((item) => item.id > 0 && item.name)
    .slice(0, limit);
}

export function normalizeMovieEditForm(data: unknown, fallbackId: number): MovieEditFormData {
  const record = toRecord(data);
  return {
    tmdb_id: record.id != null ? String(record.id) : String(fallbackId || ""),
    title: toText(record.title),
    original_title: toText(record.original_title),
    genre_names: normalizeGenreNames(record.genres),
    tagline: toText(record.tagline),
    release_date: toText(record.release_date),
    status: toText(record.status),
    runtime: record.runtime != null ? String(record.runtime) : "",
    original_language: toText(record.original_language),
    homepage: toText(record.homepage),
    poster_path: toText(record.poster_path),
    backdrop_path: toText(record.backdrop_path),
    vote_average: record.vote_average != null ? String(record.vote_average) : "",
    popularity: record.popularity != null ? String(record.popularity) : "",
    overview: toText(record.overview),
  };
}

export function normalizeTVEditForm(data: unknown, fallbackId: number): TVEditFormData {
  const record = toRecord(data);
  return {
    tmdb_id: record.id != null ? String(record.id) : String(fallbackId || ""),
    name: toText(record.name),
    original_name: toText(record.original_name),
    genre_names: normalizeGenreNames(record.genres),
    type: toText(record.type),
    tagline: toText(record.tagline),
    first_air_date: toText(record.first_air_date),
    status: toText(record.status),
    number_of_seasons: record.number_of_seasons != null ? String(record.number_of_seasons) : "",
    number_of_episodes: record.number_of_episodes != null ? String(record.number_of_episodes) : "",
    original_language: toText(record.original_language),
    homepage: toText(record.homepage),
    poster_path: toText(record.poster_path),
    backdrop_path: toText(record.backdrop_path),
    vote_average: record.vote_average != null ? String(record.vote_average) : "",
    popularity: record.popularity != null ? String(record.popularity) : "",
    overview: toText(record.overview),
  };
}
