package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config хранит все конфигурационные параметры приложения
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Cache       CacheConfig
	MOEX        MOEXConfig
	NewsAPI     NewsAPIConfig
	APIKeys     APIKeysConfig
	LogLevel    string
	Environment string
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port           int
	Host           string
	TimeoutSeconds int
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	URI        string
	Database   string
	Collection string
	Username   string
	Password   string
	Timeout    time.Duration
}

// CacheConfig конфигурация кэша
type CacheConfig struct {
	RedisURI   string
	RedisDB    int
	DefaultTTL time.Duration
	StocksTTL  time.Duration
	NewsTTL    time.Duration
}

// MOEXConfig конфигурация API для работы с MOEX
type MOEXConfig struct {
	BaseURL  string
	Timeout  time.Duration
	UseCache bool
	APIKey   string
}

// NewsAPIConfig конфигурация API для получения новостей
type NewsAPIConfig struct {
	BaseURL  string
	Timeout  time.Duration
	UseCache bool
	APIKey   string
	Sources  []string
}

// APIKeysConfig конфигурация API ключей
type APIKeysConfig struct {
	MOEXKey    string
	NewsAPIKey string
}

// LoadConfig загружает конфигурацию из файла или переменных окружения
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("ошибка при анмаршалинге конфигурации: %w", err)
	}

	// Установка значений по умолчанию
	setDefaults(&config)

	return &config, nil
}

// setDefaults устанавливает значения по умолчанию, если они не указаны
func setDefaults(config *Config) {
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}

	if config.Server.TimeoutSeconds == 0 {
		config.Server.TimeoutSeconds = 30
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if config.Cache.DefaultTTL == 0 {
		config.Cache.DefaultTTL = 5 * time.Minute
	}

	if config.Cache.StocksTTL == 0 {
		config.Cache.StocksTTL = 15 * time.Minute
	}

	if config.Cache.NewsTTL == 0 {
		config.Cache.NewsTTL = 30 * time.Minute
	}

	if config.MOEX.Timeout == 0 {
		config.MOEX.Timeout = 10 * time.Second
	}

	if config.NewsAPI.Timeout == 0 {
		config.NewsAPI.Timeout = 10 * time.Second
	}
}
