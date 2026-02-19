package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func Load(profile string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("application")
	v.AddConfigPath("configs")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if profile != "" {
		v.SetConfigName(fmt.Sprintf("application-%s", profile))
		v.MergeInConfig() // ignore error jika file tak ada
	}

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func ActiveProfile() string { return os.Getenv("APP_PROFILE") } // e.g., dev
