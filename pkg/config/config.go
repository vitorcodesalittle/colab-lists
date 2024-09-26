package config

import (
	"flag"
	"log"
	"os"
	"time"
)

type Config struct {
	DatabaseUrl    string
	Listen         string
	UseTls         bool
	PrivateKey     string
	Certificate    string
	SessionTimeout time.Duration
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.DatabaseUrl, "database-url", "./data/colablist.db", "Database URL")
	flag.StringVar(&config.Listen, "listen", ":8080", "Listen")
	flag.BoolVar(&config.UseTls, "tls", false, "Listen")
	flag.StringVar(&config.PrivateKey, "private-key", "", "Path to file with private key")
	flag.StringVar(&config.Certificate, "certificate", "", "Path to file with certificate")
	flag.DurationVar(&config.SessionTimeout, "session-timeout", 4*time.Hour, "Session timeout")
	flag.Parse()
	if config.DatabaseUrl == "" {
		panic("Database URL is required")
	}
	if config.Listen == "" {
		panic("Listen is required")
	}
	if config.UseTls {
		_, err := os.Stat(config.PrivateKey)
		if err != nil {
			log.Println("Failed to find the private-key at " + config.PrivateKey)
			panic(err)
		}
		_, err = os.Stat(config.Certificate)
		if err != nil {
			log.Println("Failed to find the certificate at " + config.Certificate)
			panic(err)
		}
	}
	return config
}

var config *Config

func GetConfig() Config {
	if config == nil {
		config = ParseConfig()
	}
	return *config
}
