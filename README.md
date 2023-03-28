# rffmpeg-go

This is a rewrite of joshuaboniface's [rffmpeg](https://github.com/joshuaboniface/rffmpeg) in Go! This wouldn't be possible without his work.

## What is this?

`rffmpeg-go` is a remote FFmpeg wrapper used to execute FFmpeg commands on a remote server via SSH. It is most useful in situations involving media servers such as Jellyfin, where one might want to perform transcoding actions with FFmpeg on a remote machine or set of machines which can better handle transcoding, take advantage of hardware acceleration, or distribute transcodes across multiple servers for load balancing.

## Quick start

### Docker

This is the recommended and easiest method, you can make your own [Dockerfile](https://github.com/aleksasiriski/jellyfin-rffmpeg/blob/main/Dockerfile) and copy the appropriate binary located at `/app/rffmpeg-go/rffmpeg-go` from this image:
```bash
ghcr.io/aleksasiriski/rffmpeg-go
```

Or you can use my prebuilt [Jellyfin docker image](https://github.com/aleksasiriski/jellyfin-rffmpeg):
```bash
ghcr.io/aleksasiriski/jellyfin-rffmpeg
```

### Binary

1. Go to [releases](https://github.com/aleksasiriski/rffmpeg-go/releases) tab and download the latest binary for your OS

1. Move the downloaded binary somewhere useful, for instance at `/usr/local/bin/rffmpeg`

1. Make soft links to the binary named `ffmpeg` and `ffprobe` in the same folder

1. Copy the [example config](https://github.com/aleksasiriski/rffmpeg-go/blob/main/rffmpeg.example.yml) to `/etc/rffmpeg/rffmpeg.yml` and change the options to your liking. **IMPORTANT: If you want to use ENV VARS to change your config, it's required to uncomment all of the options that are desired to be changed by ENV VARS. ENV VARS take precedence over the config file but are IGNORED if the config file has the options commented**

1. Point your media program to use newly available ffmpeg link for `rffmpeg-go`, for instance at `/usr/local/bin/ffmpeg`

## Hosts configuration

For remote hosts to be able to transcode files sent by `rffmpeg-go` it is required for those hosts to have access to the media files that need transcodes as well as the directory which is used to store transcoded media at **the same path** as the local host running `rffmpeg-go`.

For example, if using Jellyfin: remote hosts need access to Jellyfin's media files as well as the temporary `transcodes` directory, and both the media files must be mounted to exactly the same location as they are on the local host.

### Setup

The easiest way to setup remote transcoding hosts is to use this [docker image](https://github.com/aleksasiriski/rffmpeg-worker):
```bash
ghcr.io/aleksasiriski/rffmpeg-worker
```

In addition to that, the easiest method to share media and transcoding dir is to setup a [NFS share](https://github.com/aleksasiriski/jellyfin-rffmpeg/blob/main/docker-compose.example.yml).

### Adding
To add a target host, use the command:
```bash
rffmpeg add [-w/--weight int] [-n/--name string] <hostname/ip>
```

This command takes the optional weight flag to adjust the weight of the target host (see below) and name flag to set the server name (defaults to the hostname). A host can be added more than once under a different name.

### Removing

To remove a target host, use the command:
```bash
rffmpeg remove <name>
```

This command takes a specific target server name. Removing an in-use target host will not terminate any running processes, though it may result in undefined behaviour within rffmpeg-go. Before removing a host it is best to ensure there is nothing using it.

## Logic

### Localhost and Fallback

If one of the configured target hosts is called `localhost` or `127.0.0.1`, `rffmpeg-go` will run the `ffmpeg`/`ffprobe` commands locally without SSH. This can be useful if the local machine is also a powerful transcoding device, but you still want to offload some transcoding jobs to other machines.

In addition, `rffmpeg-go` will fall back to `localhost` automatically, even if it is not explicitly configured, should it be unable to find any working remote hosts. This helps prevent situations where `rffmpeg-go` cannot be run due to none of the remote host(s) being available.

In both cases, note that, if hardware acceleration is configured, it must be available on the local host as well, or the `ffmpeg` commands will fail. There is no easy way around this without rewriting arguments, and this is currently out-of-scope for `rffmpeg-go`. You should always use a lowest-common-denominator approach when deciding on what additional option(s) to enable, such that any configured host can run any process, or accept that fallback will not work if all remote hosts are unavailable.

The exact path to the local `ffmpeg` and `ffprobe` binaries can be overridden in the configuration, should their paths not match those of the remote system(s).

### Target Host Selection

When more than one target host is present, `rffmpeg-go` uses the following rules to select a target host. These rules are evaluated each time a new `rffmpeg-go` alias process is spawned based on the current state (actively running processes, etc.).

1. Any hosts marked `bad` are ignored.

1. All remaining hosts are iterated through in an indeterminate order. For each host:

    a. If the host is not `localhost`/`127.0.0.1`, it is tested to ensure it is reachable (responds to `ffmpeg -version` over SSH). If it is not reachable, it is marked `bad` for the duration of this processes' runtime and skipped.

    b. If the host is `idle` (has no running processes), it is immediately chosen and the iteration stops.

    c. If the host is `active` (has at least one running process), it is checked against the host with the current fewest number of processes, adjusted for host weight. If it has the fewest, it takes over this role.

1. Once all hosts have been iterated through, at least one host should have been chosen: either the first `idle` host, or the host with the fewest number of active processes. `rffmpeg-go` will then begin running against this host. If no valid target host was found, `localhost` is used (see section [Localhost and Fallback](#localhost-and-fallback) above).

### Target Host Weights and Duplicated Target Hosts

When adding a host to `rffmpeg-go`, a weight can be specified. Weights are used during the calculation of the fewest number of processes among hosts. The actual number of processes running on the host is floor divided (rounded down to the nearest divisible integer) by the weight to give a "weighted count", which is then used in the determination. This option allows one host to take on more processes than other nodes, as it will be chosen as the "least busy" host more often.

For example, consider two hosts: `host1` with weight 1, and `host2` with weight 5. `host2` would have its actual number of processes floor divided by `5`, and thus any number of processes under `5` would count as `0`, any number of processes between `5` and `10` would count as `1`, and so on, resulting in `host2` being chosen over `host1` even if it had several processes. Thus, `host2` would on average handle 5x more `ffmpeg` processes than `host1` would.

Host weighting is a fairly blunt instrument, and only becomes important when many simultaneous `ffmpeg` processes/transcodes are occurring at once across at least 2 remote hosts, and where the target hosts have significantly different performance profiles. Generally leaving all hosts at weight 1 would be sufficient for most use-cases.

Furthermore, it is possible to add a host of the same name more than once in the `rffmpeg add` command. This is functionally equivalent to setting the host with a higher weight, but may have some subtle effects on host selection beyond what weight alone can do; this is probably not worthwhile but is left in for the option.

### `bad` hosts

As mentioned above under [Target Host Selection](#target-host-selection), a host can be marked `bad` if it does not respond to an `ffmpeg -version` command in at least 1 second if it is due to be checked as a target for a new `rffmpeg-go` alias process. This can happen because a host is offline, unreachable, overloaded, or otherwise unresponsive.

Once a host is marked `bad`, it will remain so for as long as the `rffmpeg-go` process that marked it `bad` is running. This can last anywhere from a few seconds (library scan processes, image extraction) to several tens of minutes (a long video transcode). During this time, any new `rffmpeg-go` processes that start will see that the host is marked as `bad` and thus skip it for target selection. Once the marking `rffmpeg-go` process completes or is terminated, the `bad` status of that host will be cleared, allowing the next run to try it again. This strikes a balance between always retrying known-unresponsive hosts over and over (and thus delaying process startup), and ensuring that hosts will eventually be retried.

If for some reason all configured hosts are marked `bad`, fallback will be engaged; see the above section [Localhost and Fallback](#localhost-and-fallback) for details on what occurs in this situation. An explicit `localhost` host entry cannot be marked `bad`.

## FAQ

### Can `rffmpeg-go` mangle/alter FFMPEG arguments?

Explicitly *no*. `rffmpeg-go` is not designed to interact with the arguments that the media server passes to `ffmpeg`/`ffprobe` at all, nor will it. This is an explicit design decision due to the massive complexity of FFMpeg - to do this, I would need to create a mapping of just about every possible FFMpeg argument, what it means, and when to turn it on or off, which is way out of scope.

This has a number of side effects:

 * `rffmpeg-go` does not know whether hardware acceleration is turned on or not (see above caveats under [Localhost and Fallback](#localhost-and-fallback)).
 * `rffmpeg-go` does not know what media is playing or where it's outputting files to, and cannot alter these paths.
 * `rffmpeg-go` cannot turn on or off special `ffmpeg` options depending on the host selected.

Thus it is imperative that you set up your entire system correctly for `rffmpeg-go` to work.