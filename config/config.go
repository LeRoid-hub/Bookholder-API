package config

import (
	"fmt"
	"os"

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

	checkDB(env)
	return env
}

func checkDB(env map[string]string) {
	required := []string{"DB_USER", "DB_PASSWORD", "DB_NAME", "DB_HOST", "DB_PORT"}
	for _, item := range required {
		checkEnv(item, env)
	}
}

func checkEnv(check string, env map[string]string) {
	if val, ok := env[check]; ok {
		if val == "" {
			fmt.Println(check, "is not set")
			os.Exit(1)
		}
	} else {
		fmt.Println(check, "is not set")
		os.Exit(1)
	}
}
