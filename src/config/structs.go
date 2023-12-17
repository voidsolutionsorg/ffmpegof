package config

type Program struct {
	Pid   int    `koanf:"pid"`
	Log   string `koanf:"log"`
	Debug bool   `koanf:"debug"`
}

type Directories struct {
	Persist string `koanf:"persist"`
	Owner   string `koanf:"owner"`
	Group   string `koanf:"group"`
}

type Remote struct {
	User    string   `koanf:"user"`
	Persist int      `koanf:"persist"`
	Args    []string `koanf:"args"`
}

type Commands struct {
	Ssh             string   `koanf:"ssd"`
	Pre             []string `koanf:"pre"`
	Ffmpeg          string   `koanf:"ffmpeg"`
	Ffprobe         string   `koanf:"ffprobe"`
	FallbackFfmpeg  string   `koanf:"fallback_ffmpeg"`
	FallbackFfprobe string   `koanf:"fallback_ffprobe"`
	SpecialFlags    []string `koanf:"special_flags"`
}

type Database struct {
	Type        string `koanf:"type"`
	Path        string `koanf:"path"`
	MigratorDir string `koanf:"migrator_dir"`
	Host        string `koanf:"host"`
	Port        int    `koanf:"port"`
	Name        string `koanf:"name"`
	Username    string `koanf:"username"`
	Password    string `koanf:"password"`
}

type Config struct {
	Program     Program     `koanf:"program"`
	Directories Directories `koanf:"directories"`
	Remote      Remote      `koanf:"remote"`
	Commands    Commands    `koanf:"commands"`
	Database    Database    `koanf:"database"`
}
