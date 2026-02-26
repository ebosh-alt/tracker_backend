package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	// ConfigPathEnv переменная окружения для явного пути к конфигу.
	ConfigPathEnv = "CONFIG_PATH"
	// ConfigEnvEnv переменная окружения с именем окружения (dev|prod).
	ConfigEnvEnv = "CONFIG_ENV"
	// DefaultConfigEnv окружение по умолчанию.
	DefaultConfigEnv = "dev"
	// ConfigFilePattern шаблон имени файла конфигурации.
	ConfigFilePattern = "config.%s.yml"
)

type Config struct {
	App      AppConfig      `yaml:"App" mapstructure:"App"`
	Server   ServerConfig   `yaml:"Server" mapstructure:"Server"`
	Database DatabaseConfig `yaml:"Database" mapstructure:"Database"`
	Telegram TelegramConfig `yaml:"Telegram" mapstructure:"Telegram"`
	Auth     AuthConfig     `yaml:"Auth" mapstructure:"Auth"`
}

type AppConfig struct {
	Env                string   `yaml:"env" mapstructure:"env"`
	LogLevel           string   `yaml:"logLevel" mapstructure:"logLevel"`
	CorsAllowedOrigins []string `yaml:"corsAllowedOrigins" mapstructure:"corsAllowedOrigins"`
}

type ServerConfig struct {
	AppVersion     string   `yaml:"appVersion" mapstructure:"appVersion"`
	Host           string   `yaml:"host" mapstructure:"host"`
	Port           string   `yaml:"port" mapstructure:"port"`
	TrustedProxies []string `yaml:"trustedProxies" mapstructure:"trustedProxies"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"DBName" mapstructure:"DBName"`
	SSLMode  string `yaml:"sslMode" mapstructure:"sslMode"`
	PgDriver string `yaml:"pgDriver" mapstructure:"pgDriver"`
	URL      string `yaml:"url" mapstructure:"url"`
}

type TelegramConfig struct {
	BotToken      string `yaml:"botToken" mapstructure:"botToken"`
	WebhookSecret string `yaml:"webhookSecret" mapstructure:"webhookSecret"`
}

type AuthConfig struct {
	AccessSecret  string `yaml:"accessSecret" mapstructure:"accessSecret"`
	RefreshSecret string `yaml:"refreshSecret" mapstructure:"refreshSecret"`
}

// New загружает конфиг из yml и валидирует его.
// Источник выбирается так:
// 1) если задан CONFIG_PATH — читается только этот файл,
// 2) иначе читается internal/infra/config/config.<CONFIG_ENV>.yml,
// где CONFIG_ENV: dev|prod (по умолчанию dev).
func New() (*Config, error) {
	return load(true)
}

// NewForMigrate загружает конфиг для migrate runner.
// Проверяет только поля, необходимые для подключения к БД.
func NewForMigrate() (*Config, error) {
	return load(false)
}

func load(requireRuntimeSecrets bool) (*Config, error) {
	var cfg Config
	v := viper.New()

	configPath, resolvedEnv, err := resolveConfigPath()
	if err != nil {
		return &cfg, err
	}
	v.SetConfigFile(configPath)
	v.SetConfigType("yml")

	if err := v.ReadInConfig(); err != nil {
		return &cfg, err
	}
	if err := v.UnmarshalExact(&cfg); err != nil {
		return &cfg, err
	}

	cfg.normalize(resolvedEnv)
	if err := cfg.validate(requireRuntimeSecrets); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

// Addr возвращает адрес сервера в формате host:port.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", strings.TrimSpace(s.Host), strings.TrimSpace(s.Port))
}

// DSN возвращает строку подключения к Postgres.
func (d DatabaseConfig) DSN() string {
	if raw := strings.TrimSpace(d.URL); raw != "" {
		return raw
	}

	user := url.QueryEscape(strings.TrimSpace(d.User))
	password := url.QueryEscape(strings.TrimSpace(d.Password))
	host := strings.TrimSpace(d.Host)
	port := strings.TrimSpace(d.Port)
	dbName := strings.TrimSpace(d.DBName)
	sslMode := strings.TrimSpace(d.SSLMode)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbName,
		sslMode,
	)
}

func resolveConfigPath() (string, string, error) {
	configEnv, err := resolveConfigEnv()
	if err != nil {
		return "", "", err
	}

	if fromEnv := strings.TrimSpace(os.Getenv(ConfigPathEnv)); fromEnv != "" {
		if st, err := os.Stat(fromEnv); err == nil && !st.IsDir() {
			return fromEnv, configEnv, nil
		}
		return "", "", fmt.Errorf("%s points to missing file: %s", ConfigPathEnv, fromEnv)
	}

	path := filepath.Join("internal", "infra", "config", fmt.Sprintf(ConfigFilePattern, configEnv))
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		return path, configEnv, nil
	}

	return "", "", fmt.Errorf(
		"config file not found: %s (set %s to override path)",
		path,
		ConfigPathEnv,
	)
}

func resolveConfigEnv() (string, error) {
	configEnv := strings.ToLower(strings.TrimSpace(os.Getenv(ConfigEnvEnv)))
	if configEnv == "" {
		configEnv = DefaultConfigEnv
	}
	switch configEnv {
	case "dev", "prod":
	default:
		return "", fmt.Errorf("%s must be one of: dev, prod", ConfigEnvEnv)
	}

	return configEnv, nil
}

func (c *Config) normalize(resolvedEnv string) {
	if strings.TrimSpace(c.App.Env) == "" {
		c.App.Env = resolvedEnv
	}
	if strings.TrimSpace(c.Server.Host) == "" {
		c.Server.Host = "0.0.0.0"
	}
	if strings.TrimSpace(c.Server.Port) == "" {
		c.Server.Port = "8080"
	}
	if len(c.Server.TrustedProxies) == 0 {
		c.Server.TrustedProxies = []string{"127.0.0.1", "::1"}
	}
	if strings.TrimSpace(c.App.LogLevel) == "" {
		c.App.LogLevel = "info"
	}
	if strings.TrimSpace(c.Database.PgDriver) == "" {
		c.Database.PgDriver = "postgres"
	}
	if strings.TrimSpace(c.Database.SSLMode) == "" {
		c.Database.SSLMode = "disable"
	}
	if strings.TrimSpace(c.Auth.RefreshSecret) == "" && strings.TrimSpace(c.Auth.AccessSecret) != "" {
		c.Auth.RefreshSecret = c.Auth.AccessSecret
	}
}

func (c *Config) validate(requireRuntimeSecrets bool) error {
	var missing []string

	if !c.Database.hasConnectionData() {
		missing = append(missing,
			"Database.url or Database.host/port/user/password/DBName",
		)
	}

	if requireRuntimeSecrets {
		if strings.TrimSpace(c.Telegram.BotToken) == "" {
			missing = append(missing, "Telegram.botToken")
		}
		if strings.TrimSpace(c.Telegram.WebhookSecret) == "" {
			missing = append(missing, "Telegram.webhookSecret")
		}
		if strings.TrimSpace(c.Auth.AccessSecret) == "" {
			missing = append(missing, "Auth.accessSecret")
		}
		if strings.TrimSpace(c.Auth.RefreshSecret) == "" {
			missing = append(missing, "Auth.refreshSecret")
		}
	}

	logLevel := strings.ToLower(strings.TrimSpace(c.App.LogLevel))
	switch logLevel {
	case "debug", "info", "warn", "error":
	default:
		missing = append(missing, "App.logLevel must be one of: debug|info|warn|error")
	}

	if len(missing) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (d DatabaseConfig) hasConnectionData() bool {
	if strings.TrimSpace(d.URL) != "" {
		return true
	}
	return strings.TrimSpace(d.Host) != "" &&
		strings.TrimSpace(d.Port) != "" &&
		strings.TrimSpace(d.User) != "" &&
		strings.TrimSpace(d.Password) != "" &&
		strings.TrimSpace(d.DBName) != ""
}
