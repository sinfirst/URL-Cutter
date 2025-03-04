package config

import (
	"flag"
	"strings"
)

type Config struct {
	ServerAdress string
	Host         string
	Letters      []string
}

func NewConfig() Config {
	var conf Config
	conf.Letters = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "")
	flag.StringVar(&conf.ServerAdress, "a", "localhost:8080", "server adress")
	flag.StringVar(&conf.Host, "b", "http://localhost:8080", "host")
	flag.Parse()
	return conf
}
