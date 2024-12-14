package config

import "github.com/alexflint/go-arg"

type Config struct {
	GRPCPort int    `arg:"-p,help:Port to listen on"`
	DBPath   string `arg:"-d,help:Path to database file"`
}

func New() *Config {
	cfg := &Config{
		GRPCPort: 8080,
		DBPath:   "shorturl.db",
	}

	arg.MustParse(cfg)

	return cfg
}
