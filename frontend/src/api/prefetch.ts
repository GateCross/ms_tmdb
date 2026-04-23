import { getMovieDetail } from "./movie";
import { getPersonDetail } from "./person";
import { getTVDetail } from "./tv";

export type PrefetchMediaType = "movie" | "tv" | "person";

export function prefetchMediaDetail(mediaType: PrefetchMediaType, id: number) {
  if (!Number.isFinite(id) || id <= 0) {
    return;
  }

  const task = mediaType === "movie"
    ? getMovieDetail(id)
    : mediaType === "tv"
      ? getTVDetail(id)
      : getPersonDetail(id);

  void task.catch(() => {
    // Prefetch failures should not affect navigation or UI state.
  });
}
