package osUtil

import (
	"log"
	"os"
)

func GetOsEnv() string {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		log.Fatalf("🔔🔔🔔 fatal error : APP_ENV required 🔔🔔🔔")
	}

	return appEnv
}

func IsLocal() bool {
	return GetOsEnv() == "local"
}
