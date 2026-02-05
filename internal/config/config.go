package config

import (
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var cfg *Config

func init() {
	// parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	// read config file from disk
	data, err := os.ReadFile(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("failed to read config file")
	}

	// parse yaml into config struct
	cfg = &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		logrus.WithError(err).Fatal("failed to parse config file")
	}
}

func Get() *Config {
	return cfg
}

func (r *RedisConfig) FormatSessionKey(sessionID string) string {
	// replace placeholder with actual session id
	return strings.Replace(r.SessionKeyFormat, "{session_id}", sessionID, 1)
}
