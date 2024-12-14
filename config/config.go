package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Load() map[string]string {
	var env map[string]string = make(map[string]string)

	validEnv := []string{"DB_USER", "DB_PASSWORD", "DB_NAME", "DB_HOST", "DB_PORT"}

	envpath := "./.env"

	if _, err := os.Stat(envpath); err == nil {

		dotenv, err := godotenv.Read(envpath)
		if err != nil {
			fmt.Println("Error loading .env file: ", err)
		}

		env = dotenv
	} else {
		fmt.Println("No .env file found", err)
	}

	for _, key := range validEnv {
		tempenv := os.Getenv(key)
		if tempenv != "" {
			env[key] = tempenv
		}
	}

	if len(env) == 0 {
		fmt.Println("no environment variables are set")
		os.Exit(1)
	}
	
	// TODO: check if all the required environment variables are set
	return env
}

