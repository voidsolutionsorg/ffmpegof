package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func DateString() string {
	return time.Now().Format("20060102")
}

func Setup(logDirPath string, debug bool) {
	// Generate logfile name
	datetime := DateString()
	logFilePath := path.Join(logDirPath, fmt.Sprintf("ffmpegof_%v.log", datetime))

	// Setup logger
	logger := log.Output(io.MultiWriter(zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out:        os.Stderr,
	}, zerolog.ConsoleWriter{
		TimeFormat: time.Stamp,
		Out: &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    5,
			MaxAge:     14,
			MaxBackups: 5,
		},
		NoColor: true,
	}))

	// Setup verbosity
	if debug {
		log.Logger = logger.Level(zerolog.DebugLevel)
	} else {
		log.Logger = logger.Level(zerolog.InfoLevel)
	}
}
