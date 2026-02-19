package config

type App struct {
	Name string `mapstructure:"name"`
}

type Logging struct {
	Dir        string `mapstructure:"dir"`
	Filename   string `mapstructure:"filename"`
	Level      string `mapstructure:"level"`
	MaxSizeMB  int    `mapstructure:"maxSizeMB"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAgeDays int    `mapstructure:"maxAgeDays"`
	Compress   bool   `mapstructure:"compress"`
}

type Server struct {
	Port int `mapstructure:"port"`
	CORS struct {
		Enabled             bool     `mapstructure:"enabled"`
		Origins             []string `mapstructure:"origins"`
		Methods             []string `mapstructure:"methods"`
		Headers             []string `mapstructure:"headers"`
		ExposeHeaders       []string `mapstructure:"exposeHeaders"`
		Credentials         bool     `mapstructure:"credentials"`
		MaxAgeSeconds       int      `mapstructure:"maxAgeSeconds"`
		AllowWildcard       bool     `mapstructure:"allowWildcard"`
		AllowPrivateNetwork bool     `mapstructure:"allowPrivateNetwork"`
	} `mapstructure:"cors"`
}

type DB struct {
	DSN              string `mapstructure:"dsn"`
	MaxOpenConns     int    `mapstructure:"maxOpenConns"`
	MaxIdleConns     int    `mapstructure:"maxIdleConns"`
	ConnMaxLifetimeS int    `mapstructure:"connMaxLifetimeSec"`
}

type Security struct {
	Enabled bool `mapstructure:"enabled"`
	JWT     struct {
		Issuer   string `mapstructure:"issuer"`
		Audience string `mapstructure:"audience"`
		Secret   string `mapstructure:"secret"`
		TTLMin   int    `mapstructure:"ttlMinutes"`
	} `mapstructure:"jwt"`
}

// ✅ Tambahan ini yang hilang
type Metrics struct {
	Prometheus struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"prometheus"`
}

type Config struct {
	App      App      `mapstructure:"app"`
	Server   Server   `mapstructure:"server"`
	DB       DB       `mapstructure:"db"`
	Security Security `mapstructure:"security"`
	Metrics  Metrics  `mapstructure:"metrics"` // <-- penting
	Logging  Logging  `mapstructure:"logging"`
}
