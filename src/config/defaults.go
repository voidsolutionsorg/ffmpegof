package config

func New() *Config {
	return &Config{
		Program: Program{
			Log:   "/var/log/jellyfin",
			Debug: false,
		},
		Directories: Directories{
			Persist: "/run/shm",
			Owner:   "jellyfin",
			Group:   "jellyfin",
		},
		Remote: Remote{
			User:    "jellyfin",
			Persist: 300,
			Args: []string{
				"-i",
				"/var/lib/ffmpegof/.ssh/id_ed25519",
			},
		},
		Commands: Commands{
			Ssh:             "/usr/bin/ssh",
			Pre:             []string{},
			Ffmpeg:          "/usr/lib/jellyfin-ffmpeg/ffmpeg",
			Ffprobe:         "/usr/lib/jellyfin-ffmpeg/ffprobe",
			FallbackFfmpeg:  "/usr/lib/jellyfin-ffmpeg/ffmpeg",
			FallbackFfprobe: "/usr/lib/jellyfin-ffmpeg/ffprobe",
			SpecialFlags: []string{
				"-version",
				"-encoders",
				"-decoders",
				"-hwaccels",
				"-filters",
				"-h",
				"-muxers",
				"-fp_format",
			},
		},
		Database: Database{
			Type:     "sqlite",
			Path:     "/var/lib/ffmpegof/db",
			Host:     "localhost",
			Port:     5432,
			Name:     "ffmpegof",
			Username: "postgres",
			Password: "",
		},
	}
}
