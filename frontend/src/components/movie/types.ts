import type { AdminCompareFieldDetail } from "@/api/admin";
import type { CastMember, GenreOption, MovieEditFormData } from "@/utils/mediaNormalizers";

export type { GenreOption };

export type MovieCastMember = CastMember;
export type MovieEditForm = MovieEditFormData;

export type MovieDetail = {
  id?: number;
  sync_tmdb_id?: number;
  title?: string;
  original_title?: string;
  tagline?: string;
  poster_path?: string;
  backdrop_path?: string;
  vote_average?: number;
  release_date?: string;
  runtime?: number;
  status?: string;
  overview?: string;
  genres?: GenreOption[];
};

export type RemoteDiffNotice = {
  remoteFields: string[];
  localOverrideFields: string[];
  remoteDetails: AdminCompareFieldDetail[];
  localOverrideDetails: AdminCompareFieldDetail[];
};

export type RemoteDiffDecision = "unknown" | "has_diff_pending" | "keep_local" | "overwritten" | "no_diff";
