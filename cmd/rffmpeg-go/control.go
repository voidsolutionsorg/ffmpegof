package main

import (
	"fmt"
	"os"
	//"strings"

	"github.com/aleksasiriski/rffmpeg-go/processor"
	//"github.com/rs/zerolog/log"
	//"github.com/sourcegraph/conc"
	"github.com/alecthomas/kong"
)

var (
	// CLI
	cli struct {
		// flags
		Config    string `type:"path" default:"${config_dir}" env:"RFFMPEG_AUTOSCALER_CONFIG" help:"Config file path"`
		Verbosity int    `type:"counter" default:"0" short:"v" env:"RFFMPEG_AUTOSCALER_VERBOSITY" help:"Log level verbosity"`
	}
)

func runControl(config Config, proc *processor.Processor, args []string) {
	// parse cli
	ctx := kong.Parse(&cli,
		kong.Name("rffmpeg"),
		kong.Description("Remote ffmpeg"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.Vars{
			"config_dir": "/config",
		},
	)

	if err := ctx.Validate(); err != nil {
		fmt.Println("Failed parsing cli:", err)
		os.Exit(1)
	}
}
