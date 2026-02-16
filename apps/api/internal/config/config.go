package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	CorsOrigin  string
}

func Load() Config {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		Addr:        addr,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CorsOrigin:  os.Getenv("CORS_ORIGIN"),
	}
}
