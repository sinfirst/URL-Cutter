package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAdress string `env:"SERVER_ADDRESS"`
	Host         string `env:"BASE_URL"`
	FilePath     string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn  string `env:"DATABASE_DSN"`
}

func NewConfig() Config {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		fmt.Println(err)
	}
	if conf.Host != "" && conf.ServerAdress != "" {
		return conf
	}
	if conf.DatabaseDsn == "" {
		flag.StringVar(&conf.DatabaseDsn, "d", "", "database dsn") //"postgres://postgres:12345@localhost:5432/sqlForURLCutter"
	}
	if conf.FilePath == "" {
		flag.StringVar(&conf.FilePath, "f", "", "path to file") //"storage.txt"
	}
	flag.StringVar(&conf.ServerAdress, "a", "localhost:8080", "server adress")
	flag.StringVar(&conf.Host, "b", "http://localhost:8080", "host")
	flag.Parse()
	return conf
}
