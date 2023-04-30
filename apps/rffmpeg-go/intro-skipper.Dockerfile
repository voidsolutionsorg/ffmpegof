ARG JELLYFIN_TAG=10.8.10

FROM ghcr.io/onedr0p/jellyfin:${JELLYFIN_TAG}

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

USER root

COPY rffmpeg-go /usr/lib/rffmpeg-go/rffmpeg

RUN ln -s /usr/lib/rffmpeg-go/rffmpeg /usr/lib/rffmpeg-go/ffmpeg && \
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

COPY --from=ghcr.io/confusedpolarbear/jellyfin-intro-skipper:${JELLYFIN_TAG} /jellyfin/jellyfin-web /usr/share/jellyfin/web

USER kah
COPY apps/rffmpeg-go/rffmpeg.yml /etc/rffmpeg/rffmpeg.yml
COPY apps/rffmpeg-go/entrypoint.sh /entrypoint.sh
CMD ["/entrypoint.sh"]

LABEL org.opencontainers.image.source="https://github.com/jellyfin/jellyfin"