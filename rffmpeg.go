package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/aleksasiriski/rffmpeg-go/migrate"
	"github.com/aleksasiriski/rffmpeg-go/processor"
)

var (
	// CLI
	cli struct {
		// flags
		Config    string `type:"path" default:"${config_dir}" env:"RFFMPEG_AUTOSCALER_CONFIG" help:"Config file path"`
		Log       string `type:"path" default:"${log_file}" env:"RFFMPEG_AUTOSCALER_LOG" help:"Log file path"`
		Verbosity int    `type:"counter" default:"0" short:"v" env:"RFFMPEG_AUTOSCALER_VERBOSITY" help:"Log level verbosity"`
	}
)

func main() {
	// parse cli
	ctx := kong.Parse(&cli,
		kong.Name("rffmpeg-go"),
		kong.Description("Remote ffmpeg"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.Vars{
			"config_dir": "/config",
			"log_file":   "/config/log/rffmpeg.log",
		},
	)

	if err := ctx.Validate(); err != nil {
		fmt.Println("Failed parsing cli:", err)
		os.Exit(1)
	}

	// logger
	logger := log.Output(io.MultiWriter(zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out:        os.Stderr,
	}, zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out: &lumberjack.Logger{
			Filename:   cli.Log,
			MaxSize:    5,
			MaxAge:     14,
			MaxBackups: 5,
		},
		NoColor: true,
	}))

	switch {
	case cli.Verbosity == 1:
		log.Logger = logger.Level(zerolog.DebugLevel)
	case cli.Verbosity > 1:
		log.Logger = logger.Level(zerolog.TraceLevel)
	default:
		log.Logger = logger.Level(zerolog.InfoLevel)
	}

	// config
	config, err := LoadConfig(cli.Config)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Cannot load config:")
	}

	// datastore
	db, err := sql.Open(config.Database.Type, config.Database.Path)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed opening datastore:")
	}
	if config.Database.Type == "sqlite" {
		db.SetMaxOpenConns(1)
	}

	// migrator
	mg, err := migrate.New(db, config.Database.Type, config.Database.MigratorDir)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed initialising migrator:")
	}

	// processor
	proc, err := processor.New(processor.Config{
		Db:     db,
		DbType: config.Database.Type,
		Mg:     mg,
	})
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed initialising processor:")
	}

	// display initialised banner
	log.Info().
		Str("Migrator", "success").
		Str("Processor", "success").
		Msg("Initialised")

	// rffmpeg-go
	var helper conc.WaitGroup
	var worker conc.WaitGroup
	helper.Go(func() {
		for {
			worker.Go(func() {
				//Something(config, proc, client)
			})
			time.Sleep(time.Minute * 5)
		}
	})

	// handle interrupt signal
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	worker.Wait()

	// testing
	fmt.Println(proc)
}
