package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type SmtpConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromNoReply string
}

type Config struct {
	DatabaseUrl    string
	Listen         string
	UseTls         bool
	PrivateKey     string
	Certificate    string
	SessionTimeout time.Duration
	HotReload      bool
	AppUrl         string
	SmtpConfig
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.DatabaseUrl, "database-url", "./data/colablist.db", "Database URL")
	flag.StringVar(&config.Listen, "listen", ":8080", "Listen")
	flag.BoolVar(&config.UseTls, "tls", false, "Listen")
	flag.StringVar(&config.PrivateKey, "private-key", "", "Path to file with private key")
	flag.StringVar(&config.Certificate, "certificate", "", "Path to file with certificate")
	flag.DurationVar(&config.SessionTimeout, "session-timeout", 4*time.Hour, "Session timeout")
	flag.StringVar(&config.Host, "smtp-host", "", "SMTP Host")
	flag.IntVar(&config.Port, "smtp-port", 0, "SMTP Port")
	flag.StringVar(&config.FromNoReply, "smtp-noreply", "something.something.noreply@domain.com", "SMTP Password")
	flag.StringVar(&config.Password, "smtp-password", "", "SMTP Password")
	flag.StringVar(&config.Username, "smtp-username", "", "SMTP Username")
	flag.BoolVar(&config.HotReload, "hot-reload", false, "If passed, will serve a websocket endpoint that identifies this run, allowing the client to restart")
	flag.StringVar(&config.AppUrl, "app-url", "https://lists.vilmasoftware.com.br", "the URL of the app")

	flag.Parse()
	if config.DatabaseUrl == "" {
		panic("--database-url is required")
	}
	if config.Listen == "" {
		panic("-listen is required")
	}
	if !allOrNoneFilled([]string{config.Host, config.Username, strconv.Itoa(config.Port), config.Password, config.FromNoReply}) {
		panic("All SMTP parameters must be filled, or none if you don't intend to use email service")
	}
	if config.UseTls {
		_, err := os.Stat(config.PrivateKey)
		if err != nil {
			log.Println("Failed to find the private-key at " + config.PrivateKey)
			log.Fatal(err)
		}
		_, err = os.Stat(config.Certificate)
		if err != nil {
			log.Println("Failed to find the certificate at " + config.Certificate)
			log.Fatal(err)

		}
	}

	return config
}

func allOrNoneFilled(arr []string) bool {
	var emptyCount int
	for i, v := range arr {
		if v == "" {
			fmt.Printf("SMTP parameter %d is empty\n", i)
			emptyCount++
		}
	}
	if emptyCount != 0 && emptyCount != len(arr) {
		return false
	}
	return true
}

var config *Config

func GetConfig() Config {
	if config == nil {
		config = ParseConfig()
	}
	return *config
}
