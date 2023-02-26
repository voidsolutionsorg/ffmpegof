package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	//"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc"

	"github.com/aleksasiriski/rffmpeg-go/processor"
)

func runFfmpeg(config Config, proc *processor.Processor, cmd string, args []string) error {
	var worker conc.WaitGroup
	worker.Go(func() {
		log.Info().
			Msg(fmt.Sprintf("Starting rffmpeg as %s with args: %s", cmd, strings.Join(args[:], " ")))

		target, err := getTargetHost(config, proc)
		if err != nil {
			return err
		}

		if target.Hostname == nil || target.Hostname == "localhost" {
			ret := runLocalFfmpeg(config, proc, cmd, args)
		} else {
			ret := runRemoteFfmpeg(config, proc, cmd, args, target)
		}

		cleanup()
		if ret.returncode == 0 {
			log.Info().
				Str("returncode", ret.returncode).
				Msg("Finished rffmpeg")
		} else {
			log.Error().
				Str("returncode", ret.returncode).
				Msg("Finished rffmpeg with return code ")
		}
		exit(ret.returncode)
	})

	// handle interrupt signal
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	<-quitChannel
	cleanup()
	return nil
}

func runRemoteFfmpeg(config Config, proc *Processor, cmd string, args []string, target processor.Host) int {
	rffmpegSshCommand := generateSshCommand(config, target.Hostname)
	rffmpegFfmpegCommand := make([]string, 0)

	for _, cmd := range config.PreCommands {
		if cmd != nil {
			rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, cmd)
		}
	}

	stdin := os.Stdin
    stderr := os.Stderr

	if strings.Contains(cmd, "ffprobe") {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.FfprobeCommand)
		stdout := os.Stdout
	} else {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.FfmpegCommand)
		stdout := os.Stderr
	}
}