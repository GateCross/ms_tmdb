export type MediaTab = "movie" | "tv";

export type UploadingKey = "" | "movie_poster_path" | "movie_backdrop_path" | "tv_poster_path" | "tv_backdrop_path";

export type SelectOption = {
  label: string;
  value: string;
};

export type LocalMovieCreateForm = {
  title: string;
  original_title: string;
  genre_names: string[];
  release_date: string;
  status: string;
  runtime: string;
  original_language: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};

export type LocalTVCreateForm = {
  name: string;
  original_name: string;
  genre_names: string[];
  first_air_date: string;
  status: string;
  type: string;
  number_of_seasons: string;
  number_of_episodes: string;
  original_language: string;
  poster_path: string;
  backdrop_path: string;
  vote_average: string;
  popularity: string;
  overview: string;
};
