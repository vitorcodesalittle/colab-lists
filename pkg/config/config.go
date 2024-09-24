package config

import "flag"


type Config struct {
    DatabaseUrl string
    Listen string
}

func ParseConfig() *Config {
    config := &Config{}
    flag.StringVar(&config.DatabaseUrl, "database-url", "./data/colablist.db", "Database URL")
    flag.StringVar(&config.Listen, "listen", ":8080", "Listen")
    flag.Parse()
    if config.DatabaseUrl == "" {
        panic("Database URL is required")
    }
    if config.Listen == "" {
        panic("Listen is required")
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
