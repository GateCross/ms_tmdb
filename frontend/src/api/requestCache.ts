type CacheEntry<T> = {
  data: T;
  expiresAt: number;
};

const responseCache = new Map<string, CacheEntry<unknown>>();
const inflightCache = new Map<string, Promise<unknown>>();

export async function withRequestCache<T>(
  key: string,
  loader: () => Promise<T>,
  ttlMs = 10 * 60 * 1000,
): Promise<T> {
  const now = Date.now();
  const cached = responseCache.get(key);
  if (cached && cached.expiresAt > now) {
    return cached.data as T;
  }

  const inflight = inflightCache.get(key);
  if (inflight) {
    return inflight as Promise<T>;
  }

  const request = loader()
    .then((data) => {
      responseCache.set(key, {
        data,
        expiresAt: Date.now() + ttlMs,
      });
      inflightCache.delete(key);
      return data;
    })
    .catch((error) => {
      inflightCache.delete(key);
      throw error;
    });

  inflightCache.set(key, request);
  return request;
}

export function clearRequestCache(prefix?: string) {
  if (!prefix) {
    responseCache.clear();
    inflightCache.clear();
    return;
  }

  for (const key of responseCache.keys()) {
    if (key.startsWith(prefix)) {
      responseCache.delete(key);
    }
  }

  for (const key of inflightCache.keys()) {
    if (key.startsWith(prefix)) {
      inflightCache.delete(key);
    }
  }
}
