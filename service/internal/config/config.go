package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/httpauthshim/authpublic"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

type Config struct {
	Listen        string             `yaml:"listen"`
	SiteTitle     string             `yaml:"siteTitle"`
	Auth          *authpublic.Config `yaml:"auth"`
	Database      DatabaseConfig     `yaml:"database"`
	ConfigVersion int                `yaml:"configVersion"`
}

// RequiredMigration is the sql-migrate id this binary expects to be applied.
const RequiredMigration = "14.quotes-markdown-enabled.sql"

var configDirOverride string

func SetConfigDir(dir string) {
	configDirOverride = dir
}

func GetConfigPath() string {
	if p := overrideConfigPath(); p != "" {
		return p
	}
	return firstExistingPath([]string{
		"./config.yaml",
		"./config/config.yaml",
		os.Getenv("FARIDOON_CONFIG_FILE"),
		"/config/config.yaml",
	})
}

func overrideConfigPath() string {
	if configDirOverride == "" {
		return ""
	}
	return existingPath(filepath.Join(configDirOverride, "config.yaml"))
}

func firstExistingPath(candidates []string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if found := existingPath(p); found != "" {
			return found
		}
	}
	return ""
}

func existingPath(p string) string {
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func Defaults() *Config {
	return &Config{
		ConfigVersion: 1,
		Listen:        ":8080",
		SiteTitle:     "Faridoon",
		Database: DatabaseConfig{
			Host:     envOr("DB_HOST", "mysql"),
			Port:     3306,
			Name:     firstNonEmpty(os.Getenv("DB_DATABASE"), os.Getenv("DB_NAME"), "faridoon"),
			User:     firstNonEmpty(os.Getenv("DB_USERNAME"), os.Getenv("DB_USER"), "user"),
			Password: firstNonEmpty(os.Getenv("DB_PASSWORD"), os.Getenv("DB_PASS"), ""),
		},
	}
}

func LoadConfig() *Config {
	cfg := Defaults()
	cfg.Auth = &authpublic.Config{}
	path := GetConfigPath()
	if path == "" {
		return cfg
	}
	loadConfigFile(cfg, path)
	applyEnvOverrides(cfg)
	applyConfigFallbacks(cfg)
	return cfg
}

func loadConfigFile(cfg *Config, path string) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return
	}
	_ = k.Unmarshal("", cfg)
}

func applyConfigFallbacks(cfg *Config) {
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.SiteTitle == "" {
		cfg.SiteTitle = "Faridoon"
	}
}

func applyEnvOverrides(cfg *Config) {
	setIfEnv(&cfg.Database.Host, "DB_HOST")
	setIfNonEmpty(&cfg.Database.Name, firstNonEmpty(os.Getenv("DB_DATABASE"), os.Getenv("DB_NAME")))
	setIfNonEmpty(&cfg.Database.User, firstNonEmpty(os.Getenv("DB_USERNAME"), os.Getenv("DB_USER")))
	setIfNonEmpty(&cfg.Database.Password, firstNonEmpty(os.Getenv("DB_PASSWORD"), os.Getenv("DB_PASS")))
	setIfEnv(&cfg.SiteTitle, "SITE_TITLE")
}

func setIfEnv(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func setIfNonEmpty(dst *string, v string) {
	if v != "" {
		*dst = v
	}
}

// ListenAddr returns the bind address: $PORT if set, otherwise cfg.Listen (default :8080).
// PORT may be a bare port ("8080") or a full address (":8080" / "0.0.0.0:8080").
func ListenAddr(cfg *Config) string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return normalizePort(port)
	}
	if cfg != nil && cfg.Listen != "" {
		return cfg.Listen
	}
	return ":8080"
}

func normalizePort(port string) string {
	if strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
