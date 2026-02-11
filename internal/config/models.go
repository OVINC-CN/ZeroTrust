package config

import "time"

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type RedisConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Password         string `yaml:"password"`
	DB               int    `yaml:"db"`
	SessionKeyFormat string `yaml:"session_key_format"`
}

type ResourceConfig struct {
	ServiceName string            `yaml:"service_name"`
	Attributes  map[string]string `yaml:"attributes"`
}

type OTelConfig struct {
	Enabled  bool           `yaml:"enabled"`
	Endpoint string         `yaml:"endpoint"`
	Insecure bool           `yaml:"insecure"`
	Resource ResourceConfig `yaml:"resource"`
}

type AuthConfig struct {
	ClientIPHeader     string   `yaml:"client_ip_header"`
	SessionCookieName  string   `yaml:"session_cookie_name"`
	LoginUrl           string   `yaml:"login_url"`
	LoginRedirectParam string   `yaml:"login_redirect_param"`
	TraceIDHeader      string   `yaml:"trace_id_header"`
	VerifyMethods      []string `yaml:"verify_methods"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
	OTel   OTelConfig   `yaml:"otel"`
	Auth   AuthConfig   `yaml:"auth"`
}
