FROM nginx:1.30-alpine

ARG TARGETARCH

LABEL org.opencontainers.image.title="ms_tmdb"
LABEL org.opencontainers.image.description="TMDB proxy and local enhancement platform runtime image"

WORKDIR /app

# Frontend dist (provided by workflow artifact download)
COPY build-artifacts/frontend/dist /usr/share/nginx/html

# Backend binary and config (provided by workflow artifact download + repo files)
COPY build-artifacts/backend/${TARGETARCH}/tmdb /app/tmdb
COPY backend/etc /app/etc

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/entrypoint.sh /entrypoint.sh

RUN chmod +x /app/tmdb /entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
