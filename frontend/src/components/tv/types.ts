import type { AdminCompareFieldDetail } from "@/api/admin";

export type GenreOption = {
  id: number;
  name: string;
};

export type TVEditForm = {
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

export type RemoteDiffNotice = {
  remoteSummary: string;
  localOverrideSummary: string;
  remoteFields: string[];
  localOverrideFields: string[];
  remoteDetails: AdminCompareFieldDetail[];
  localOverrideDetails: AdminCompareFieldDetail[];
};

export type RemoteDiffDecision = "unknown" | "has_diff_pending" | "keep_local" | "overwritten" | "no_diff";

export type TVSeasonSummary = {
  id: number;
  season_number: number;
  name: string;
  poster_path: string;
  episode_count: number;
};

export type TVEpisodeItem = {
  id: number;
  episode_number: number;
  name: string;
  air_date: string;
  runtime: number | null;
  vote_average: number | null;
  overview: string;
  still_path: string;
};

export type TVSeasonDetail = {
  id: number;
  season_number: number;
  name: string;
  air_date: string;
  overview: string;
  poster_path: string;
  episodes: TVEpisodeItem[];
};

export type TVSeasonForm = {
  season_number: string;
  name: string;
  air_date: string;
  poster_path: string;
  overview: string;
};

export type TVEpisodeForm = {
  episode_number: string;
  name: string;
  air_date: string;
  runtime: string;
  vote_average: string;
  still_path: string;
  overview: string;
};

export type TVCastMember = {
  id: number;
  name: string;
  character: string;
  profile_path: string;
};
