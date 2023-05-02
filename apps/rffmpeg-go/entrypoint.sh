#!/usr/bin/env bash

#shellcheck disable=SC1091
test -f "/scripts/umask.sh" && source "/scripts/umask.sh"

#shellcheck disable=SC2086
exec \
    /usr/bin/jellyfin \
        --ffmpeg="/usr/local/bin/ffmpeg" \
        --webdir="/usr/share/jellyfin/web" \
        --datadir="/config" \
        --cachedir="/config/cache" \
        "$@"