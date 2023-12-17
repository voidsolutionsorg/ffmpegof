package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
)

func (c *Config) Load(dataDirPath string) error {
	// Use "." as the key path delimiter. This can be "/" or any character.
	k := koanf.New(".")

	// Load default values using the structs provider.
	// We provide a struct along with the struct tag `koanf` to the
	// provider.
	if err := k.Load(structs.Provider(c, "koanf"), nil); err != nil {
		log.Panic().Err(err).Msg("config.Load(): failed loading default values")
	}

	// Load YAML config
	yamlPath := path.Join(dataDirPath, "ffmpegof.yaml")
	if _, err := os.Stat(yamlPath); err != nil {
		log.Trace().Msgf("config.Load(): no yaml config present at path: %v, looking for .yml", yamlPath)
		yamlPath = path.Join(dataDirPath, "ffmpegof.yml")
		if _, errr := os.Stat(yamlPath); errr != nil {
			log.Trace().Msgf("config.Load(): no yaml config present at path: %v", yamlPath)
		} else if errr := k.Load(file.Provider(yamlPath), yaml.Parser()); errr != nil {
			return fmt.Errorf("config.Load(): error loading yaml config")
		}
	} else if err := k.Load(file.Provider(yamlPath), yaml.Parser()); err != nil {
		return fmt.Errorf("config.Load(): error loading yaml config")
	}

	// Load ENV config
	if err := k.Load(env.Provider("FFMPEGOF_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "FFMPEGOF_")), "_", ".", -1)
	}), nil); err != nil {
		return fmt.Errorf("config.Load(): error loading env config")
	}

	// Unmarshal config into struct
	if err := k.Unmarshal("", c); err != nil {
		return fmt.Errorf("config.Load(): failed unmarshaling koanf config")
	}

	// Set database config
	switch c.Database.Type {
	case "sqlite":
		{
			err := os.MkdirAll(path.Dir(c.Database.Path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed creating database directory: %w", err)
			}
			dbpath, err := filepath.Abs(c.Database.Path)
			if err != nil {
				return fmt.Errorf("failed loading sqlite file: %w", err)
			}
			c.Database.Path = dbpath + "?_foreign_keys=on"
			c.Database.MigratorDir = "migrations/sqlite"
		}
	case "postgres":
		{
			c.Database.Path = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Database.Host, c.Database.Port, c.Database.Username, c.Database.Password, c.Database.Name)
			c.Database.MigratorDir = "migrations/postgres"
		}
	default:
		return fmt.Errorf("database type isn't supported")
	}

	// Set program PID
	c.Program.Pid = os.Getpid()

	return nil
}
