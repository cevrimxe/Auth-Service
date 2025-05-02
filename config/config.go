package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// func LoadEnv() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// }

func LoadEnv() { //renderda deploy ettiğim için bu kısmı kullanıyorum.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with system environment variables...")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
