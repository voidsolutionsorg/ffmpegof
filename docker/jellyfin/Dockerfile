ARG JELLYFIN_VERSION=10.9.11

FROM docker.io/jellyfin/jellyfin:${JELLYFIN_VERSION}

ENV HEALTHCHECK_URL=http://localhost:8096/health \
    FFMPEGOF_PROGRAM_LOG=/config/log \
    FFMPEGOF_DIRECTORIES_OWNER=root \
    FFMPEGOF_DIRECTORIES_GROUP=root \
    FFMPEGOF_REMOTE_USER=root \
    FFMPEGOF_DATABASE_PATH=/config/ffmpegof/db

COPY ffmpegof /usr/local/bin/ffmpegof
RUN ln -s /usr/local/bin/ffmpegof /usr/local/bin/ffmpeg \
    && ln -s /usr/local/bin/ffmpegof /usr/local/bin/ffprobe

RUN mkdir -p /run/shm

EXPOSE 8096
VOLUME /config
ENTRYPOINT ["./jellyfin/jellyfin", \
    "--datadir", "/config", \
    "--cachedir", "/config/cache", \
    "--ffmpeg", "/usr/local/bin/ffmpeg"]

HEALTHCHECK --interval=30s --timeout=30s --start-period=10s --retries=3 \
    CMD curl -Lk -fsS "${HEALTHCHECK_URL}" || exit 1

LABEL org.opencontainers.image.source="https://github.com/tminaorg/jellyfin-ffmpegof"
