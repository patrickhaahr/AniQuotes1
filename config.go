package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Envs = initConfig()

type Config struct {
	SendGridAPIKey    string
	SendGridFromEmail string
}

func initConfig() *Config {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		SendGridAPIKey:    getEnvOrPanic("SENDGRID_API_KEY", "SendGrid API Key is required"),
		SendGridFromEmail: getEnvOrPanic("SENDGRID_FROM_EMAIL", "SendGrid From Email is required"),
	}
}

func getEnvOrPanic(key, err string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(err)
}
