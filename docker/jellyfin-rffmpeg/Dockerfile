ARG JELLYFIN_TAG=10.8.10

FROM ghcr.io/onedr0p/jellyfin:${JELLYFIN_TAG}

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY ["docker/jellyfin-rffmpeg/entrypoint.sh", "/entrypoint.sh"]
COPY ["docker/jellyfin-rffmpeg/rffmpeg.yml", "/etc/rffmpeg/rffmpeg.yml"]
COPY ["dist/rffmpeg-go_${TARGETOS}_${TARGETARCH}${TARGETVARIANT:+_7}/rffmpeg-go", "/usr/lib/rffmpeg-go/rffmpeg"]

RUN chmod +x /usr/lib/rffmpeg-go/rffmpeg && \
    ln -s /usr/lib/rffmpeg-go/rffmpeg /usr/lib/rffmpeg-go/ffmpeg && \
    ln -s /usr/lib/rffmpeg-go/rffmpeg /usr/lib/rffmpeg-go/ffprobe

RUN apt-get -qq update \
    && apt-get -qq install -y openssh-client \
    && apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false \
    && apt-get autoremove -y \
    && apt-get clean \
    && \
    rm -rf \
        /tmp/* \
        /var/lib/apt/lists/* \
        /var/tmp/

LABEL org.opencontainers.image.source="https://github.com/jellyfin/jellyfin"