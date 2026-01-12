package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	MongoURI          string
	MongoDBName       string
	JWTSecret         string
	AzureOpenAIEndpoint string
	AzureOpenAIKey      string
	AWSRegion         string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:       getEnv("MONGO_DB_NAME", "ai_proxy"),
		JWTSecret:         getEnv("JWT_SECRET", "default_secret"),
		AzureOpenAIEndpoint: getEnv("AZURE_OPENAI_ENDPOINT", ""),
		AzureOpenAIKey:      getEnv("AZURE_OPENAI_API_KEY", ""),
		AWSRegion:         getEnv("AWS_REGION", "us-east-1"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
