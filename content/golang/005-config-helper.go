package config

import (
	"fmt"
	"log"
	"os"
)

func MustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		msg := fmt.Sprintf("FATAL: Environment variable %s is not set", key)
		log.Fatal(msg)
	}
	return value
}

var DatabaseURL = MustGetenv("DATABASE_URL")
