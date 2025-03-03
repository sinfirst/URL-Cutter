package config

import "strings"

type Config struct {
	ServerAdress string
	Host         string
	Letters      []string
}

func NewConfig() Config {
	var conf Config
	conf.ServerAdress = "http://localhost:8080"
	conf.Host = ":8080"
	conf.Letters = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", "")
	return conf
}
