# jellyfin-ffmpegof

## Unofficial docker image made by including [ffmpegof](https://github.com/tminaorg/ffmpegof) in official Jellyfin docker image.

### Note:

- This image uses `/config/cache` for cache dir by default, instead of the official `/cache`. This allows to use a single volume for stateful data, which can save costs when using Kubernetes in the cloud.
- The public ssh key is located inside the container at `/config/ffmpegof/.ssh/id_ed25519.pub`
- The known_hosts file is located inside the container at `/config/ffmpegof/.ssh/known_hosts`

## Setup

### Database

#### SQLite

SQLite is already provided and configured, you can start by adding workers.

#### Postgresql

If you want to use this container as a stateless app (currently not possible because Jellyfin itself isn't stateless) set the required fields in `/config/ffmpegof/ffmpegof.yaml` or equivalent env vars that take precedence for Postgresql:

| Name                       | Default value |                    Description |
| :------------------------- | :-----------: | -----------------------------: |
| FFMPEGOF_DATABASE_TYPE     |    sqlite     | Must be 'sqlite' or 'postgres` |
| FFMPEGOF_DATABASE_HOST     |   localhost   |         Postgres database host |
| FFMPEGOF_DATABASE_PORT     |     5432      |         Postgres database port |
| FFMPEGOF_DATABASE_NAME     |   ffmpegof    |         Postgres database name |
| FFMPEGOF_DATABASE_USERNAME |   postgres    |     Postgres database username |
| FFMPEGOF_DATABASE_PASSWORD |      ""       |     Postgres database password |

### Workers

Workers must have access to Jellyfin's `/config/cache`, `/config/transcodes` and `/config/data/subtitles` directories. It's recommended to setup [NFS share](https://github.com/tminaorg/ffmpegof/blob/main/docker-compose.example.yml) for this.

For a worker docker image you can use [this](https://github.com/tminaorg/rffmpeg-worker).

#### Adding new workers

Generate a new ssh key for Jellyfin:

```bash
docker compose exec -it jellyfin ssh-keygen -t ed25519 -f /config/ffmpegof/.ssh/id_ed25519 -q -N ""
```

**Absolutely make sure that `/config/ffmpegof/.ssh` has 700 or 600 chmod-ed permissions, otherwise the remote ssh commands will fail with UNPROTECTED PRIVATE KEY FILE! error**

Copy the public ssh key to the worker if it supports password login:

```bash
docker compose exec -it jellyfin ssh-copy-id -i /config/ffmpegof/.ssh/id_ed25519.pub root@<worker_ip_address>
```

If the worker doesn't support password login, you will have to copy the public key manually.

**If you're using my rffmpeg-worker image, then the copied keys are stateless. Either add a volume to `/etc/authorized_keys` or make a file binding for the required ssh keys**

Add the worker to ffmpegof db:

```bash
docker compose exec -it jellyfin ffmpegof add [--weight 1] [--name first_worker] <worker_ip_address>
```

Check the status of ffmpegof:

```bash
docker compose exec -it jellyfin ffmpegof status
```

### Hardware Acceleration

Enable it normally in the Jellyfin admin panel.

If you want to use Hardware Acceleration all of the workers **must** support the same tech (VAAPI, NVENC, etc.).

**Note**: If the Jellyfin host doesn't support that same Hardware Accel tech then it **can't** be used as a failover, but if you have available workers it will still transcode without problems.

## Kubernetes

On Kubernetes you can use [OpenEBS](https://github.com/openebs/dynamic-nfs-provisioner) to create RWX from RWO volume or [Longhorn](https://longhorn.io) RWX volumes (NFSv4) and mount said paths to Jellyfin host and workers (must be exactly the same mount points!).

Here's a [Helm chart repo with instuctions](https://github.com/tminaorg/jellyfin-kubernetes)
