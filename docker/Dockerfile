FROM sc4h/alpine-s6overlay:v2-3.15

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ENV \
  PATH="/app/rffmpeg-go:${PATH}"

# Binary
COPY ["docker/rffmpeg-go.yml.example", "/etc/rffmpeg/rffmpeg.yml"]

# Binary
COPY ["dist/rffmpeg-go_${TARGETOS}_${TARGETARCH}${TARGETVARIANT:+_7}/rffmpeg-go", "/app/rffmpeg-go/rffmpeg-go"]

# Volume
VOLUME ["/config"]