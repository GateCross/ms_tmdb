export type MediaGenre = {
  id?: number;
  name: string;
};

export type MediaSummary = {
  id?: number;
  tmdb_id: number;
  title?: string;
  original_title?: string;
  name?: string;
  original_name?: string;
  poster_path?: string;
  backdrop_path?: string;
  vote_average?: number;
  popularity?: number;
  release_date?: string;
  first_air_date?: string;
  number_of_seasons?: number;
  number_of_episodes?: number;
  is_modified?: boolean;
  media_type?: "movie" | "tv" | "person" | "multi" | string;
  profile_path?: string;
  overview?: string;
  genre_names?: string[];
};

export type SearchResultItem = {
  id: number;
  media_type?: "movie" | "tv" | "person" | "multi" | string;
  title?: string;
  original_title?: string;
  name?: string;
  original_name?: string;
  poster_path?: string;
  profile_path?: string;
  vote_average?: number;
  release_date?: string;
  first_air_date?: string;
  overview?: string;
};

export type ApiErrorLike = {
  message?: string;
};
