// Package config пакет с инициализацией конфига
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
)

var once sync.Once

// TokenExp для работы jwt
var TokenExp = time.Hour * 12

// SecretKey для работы jwt
var SecretKey = "supersecretkey"

// Config структура
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" json:"server_address"`
	Host          string `env:"BASE_URL" json:"base_url"`
	FilePath      string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDsn   string `env:"DATABASE_DSN" json:"database_dsn"`
	HTTPSEnable   bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFile    string `env:"CONFIG"`
}

// NewConfig конструктор для конфига
func NewConfig() (Config, error) {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		return Config{}, err
	}

	if conf.Host != "" && conf.ServerAddress != "" {
		return conf, nil
	}
	once.Do(func() {
		if conf.DatabaseDsn == "" {
			flag.StringVar(&conf.DatabaseDsn, "d", "", "database dsn") //"postgres://postgres:qwerty12345@localhost:5432/postgres"
		}

		if conf.FilePath == "" {
			flag.StringVar(&conf.FilePath, "f", "", "path to file") //"storage.txt"
		}

		flag.StringVar(&conf.ServerAddress, "a", "localhost:8080", "server adress")
		flag.StringVar(&conf.Host, "b", "http://localhost:8080", "host")
		flag.BoolVar(&conf.HTTPSEnable, "s", false, "https")
		flag.Parse()
	})
	if conf.DatabaseDsn == "" && conf.FilePath == "" && conf.Host == "" && conf.ServerAddress == "" && conf.ConfigFile != "" {
		conf, err := fileConfig(conf.ConfigFile)
		if err != nil {
			return Config{}, fmt.Errorf("failed to load config from file: %w", err)
		}
		return *conf, nil
	}
	return conf, nil
}

// fileConfig сканирует конфигурационные данные из файла
func fileConfig(path string) (*Config, error) {
	fmt.Println("Load config from file")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
