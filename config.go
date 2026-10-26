package main

import (
	"os"
)

type Config struct {
	BaseURL  string
	HTTPAddr string
	GRPCAddr string
}

func Load() Config {
	return Config{
		BaseURL:  getEnv("MOVIE_BASE_URL", "https://homeworksite.site"),
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr: getEnv("GRPC_ADDR", ":9090"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
