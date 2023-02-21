package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type Jellyfin struct {
	Host   string `mapstructure:"HOST"`
	SshKey string `mapstructure:"SSH_KEY"`
	Jobs   int    `mapstructure:"JOBS"`
	Weight int    `mapstructure:"WEIGHT"`
}

type Hetzner struct {
	Token          string `mapstructure:"TOKEN"`
	Server         string `mapstructure:"SERVER"`
	Image          string `mapstructure:"IMAGE"`
	SshKey         string `mapstructure:"SSH_KEY"`
	Network        string `mapstructure:"NETWORK"`
	Firewall       string `mapstructure:"FIREWALL"`
	PlacementGroup string `mapstructure:"PLACEMENT_GROUP"`
	Location       string `mapstructure:"LOCATION"`
	CloudInit      string `mapstructure:"CLOUD_INIT"`
}

type Database struct {
	Type        string `mapstructure:"TYPE"`
	Path        string `mapstructure:"PATH"`
	MigratorDir string `mapstructure:"MIGRATOR_DIR"`
	Host        string `mapstructure:"HOST"`
	Port        int    `mapstructure:"PORT"`
	Name        string `mapstructure:"NAME"`
	Username    string `mapstructure:"USERNAME"`
	Password    string `mapstructure:"PASSWORD"`
}

type Media struct {
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}

type Config struct {
	Jellyfin Jellyfin `mapstructure:"JELLYFIN"`
	Hetzner  Hetzner  `mapstructure:"HETZNER"`
	Database Database `mapstructure:"DATABASE"`
	Media    Media    `mapstructure:"MEDIA"`
}

func LoadConfig(path string) (Config, error) {
	config := Config{
		Jellyfin: Jellyfin{
			SshKey: "/config/rffmpeg/.ssh/id_ed25519.pub",
			Jobs:   2,
			Weight: 1,
		},
		Hetzner: Hetzner{
			Server:         "cpx21",
			Image:          "docker-ce",
			SshKey:         "root@jellyfin",
			Network:        "rffmpeg-workers",
			Firewall:       "rffmpeg-workers",
			PlacementGroup: "rffmpeg-workers",
			Location:       "nbg1",
			CloudInit:      "#cloud-config\nruncmd:\n- systemctl disable --now ssh.service\n- echo 'JELLYFIN_LAN_ONLY_IP=%s' | tee -a /root/.env\n- echo 'MEDIA_USERNAME=%s' | tee -a /root/.env\n- echo 'MEDIA_PASSWORD=%s' | tee -a /root/.env\n- wget https://raw.githubusercontent.com/aleksasiriski/rffmpeg-worker/main/docker-compose.example.yml -O /root/docker-compose.yml\n- cd /root && docker compose pull && docker compose up -d\n",
		},
		Database: Database{
			Type:        "sqlite",
			Path:        "/config/rffmpeg/rffmpeg.db",
			MigratorDir: "migrations/sqlite",
			Host:        "localhost",
			Port:        5432,
			Name:        "rffmpeg",
			Username:    "postgres",
		},
	}

	viper.AddConfigPath(path)
	viper.SetConfigName("rffmpeg-autoscaler")
	viper.SetConfigType("yaml")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("Failed parsing config: %w", err)
		}
	}
	err = viper.Unmarshal(&config)

	if config.Jellyfin.Host == "" {
		return config, fmt.Errorf("Jellyfin host is not specified!")
	}
	if config.Hetzner.Token == "" {
		return config, fmt.Errorf("Hetzner token is not specified!")
	}
	if config.Database.Type != "sqlite" && config.Database.Type != "postgres" {
		return config, fmt.Errorf("Database type must be sqlite or postgres!")
	}

	sshkeypath, err := filepath.Abs(config.Jellyfin.SshKey)
	if err != nil {
		return config, fmt.Errorf("Failed loading ssh key file: %w", err)
	}
	dbpath, err := filepath.Abs(config.Database.Path)
	if err != nil {
		return config, fmt.Errorf("Failed loading sqlite file: %w", err)
	}
	config.Jellyfin.SshKey = sshkeypath
	config.Database.Path = dbpath

	switch config.Database.Type {
	case "sqlite":
		{
			config.Database.MigratorDir = "migrations/sqlite"
			err := os.MkdirAll(filepath.Dir(config.Database.Path), os.ModePerm)
			if err != nil {
				return config, fmt.Errorf("Failed creating database directory: %w", err)
			}
			config.Database.Path = config.Database.Path + "?_foreign_keys=on"
		}
	case "postgres":
		{
			config.Database.MigratorDir = "migrations/postgres"
			config.Database.Path = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Name)
		}
	}
	config.Hetzner.CloudInit = fmt.Sprintf(config.Hetzner.CloudInit, config.Jellyfin.Host, config.Media.Username, config.Media.Password)

	return config, err
}
