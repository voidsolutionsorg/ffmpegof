package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tminaorg/ffmpegof/src/config"
	"github.com/tminaorg/ffmpegof/src/control"
	"github.com/tminaorg/ffmpegof/src/ffmpeg"
	"github.com/tminaorg/ffmpegof/src/logger"
	"github.com/tminaorg/ffmpegof/src/migrate"
	"github.com/tminaorg/ffmpegof/src/processor"
)

func main() {
	// load config
	c := config.New()
	if err := c.Load("/etc/ffmpegof"); err != nil {
		panic(fmt.Errorf("cannot load config: %s", err.Error()))
	}

	// setup logger
	logger.Setup(c.Program.Log, c.Program.Debug)

	// setup datastore
	db, err := sql.Open(c.Database.Type, c.Database.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed opening datastore")
	}

	// setup migrator
	mg, err := migrate.New(db, c.Database.Type, c.Database.MigratorDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initialising migrator")
	}

	// setup processor
	proc, err := processor.New(processor.Config{
		Db:     db,
		DbType: c.Database.Type,
		Mg:     mg,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed initialising processor")
	}

	// check database connection
	databaseVersion, err := proc.GetVersion()
	if err != nil {
		log.Fatal().Err(err).Msg("failed getting database version")
	} else {
		log.Info().Msg(fmt.Sprintf("database in use: %s", databaseVersion))
	}

	// ffmpegof startup
	cmd := os.Args[0]
	args := os.Args[1:]
	if strings.Contains(cmd, "ffmpegof") {
		control.Run(proc)
	} else if strings.Contains(cmd, "ffmpeg") || strings.Contains(cmd, "ffprobe") {
		ffmpeg.Run(c, proc, cmd, args)
	} else {
		log.Fatal().Msg("entrypoint command must be one of three: [ffmpegof, ffmpeg, ffprobe]")
	}
}
