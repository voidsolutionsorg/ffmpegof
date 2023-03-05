package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/aleksasiriski/rffmpeg-go/migrate"
	"github.com/aleksasiriski/rffmpeg-go/processor"
)

func main() {
	// config
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Cannot load config: %w", err))
	}

	// logger
	logger := log.Output(io.MultiWriter(zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out:        os.Stderr,
	}, zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out: &lumberjack.Logger{
			Filename:   config.Program.Log,
			MaxSize:    5,
			MaxAge:     14,
			MaxBackups: 5,
		},
		NoColor: true,
	}))

	if config.Program.Debug {
		log.Logger = logger.Level(zerolog.DebugLevel)
	} else {
		log.Logger = logger.Level(zerolog.InfoLevel)
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

	// database connection
	databaseVersion, err := proc.GetVersion()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed getting database version:")
	} else {
		log.Info().
			Msg(fmt.Sprintf("Database in use: %s", databaseVersion))
	}

	// rffmpeg-go
	cmd := os.Args[0]
	args := os.Args[1:]
	if cmd == "ffmpeg" || cmd == "ffprobe" {
		runFfmpeg(config, proc, cmd, args)
	} else {
		runControl(config, proc)
	}
}
