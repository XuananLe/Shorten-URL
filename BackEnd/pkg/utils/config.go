package utils

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/joho/godotenv"
)

type config struct {
	Server struct {
		SERVER_PORT string
	}
	Database struct {
		DB_NAME     string
		DB_USERNAME string
		DB_PASSWORD string
		DB_PORT     string
	}
	Redis struct {
		REDIS_HOST string 
		REDIS_PORT string 	
		REDIS_PASS string 
		REDIS_DB string
		REDIS_USER string
	}
	Kafka struct {
		KAFKA_BROKER_URL string
		KAFKA_TOPIC      string
		KAFKA_GROUP_ID   string
	}
}

var Config *config = &config{}
var once sync.Once

func LoadEnv() *config {
	// Init only once
	once.Do(func() {
		
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		Config = &config{
			Server: struct {
				SERVER_PORT string
			}{
				SERVER_PORT: os.Getenv("SERVER_PORT"),
			},
			Database: struct {
				DB_NAME     string
				DB_USERNAME string
				DB_PASSWORD string
				DB_PORT     string
			}{
				DB_NAME:     os.Getenv("DB_NAME"),
				DB_USERNAME: os.Getenv("DB_USERNAME"),
				DB_PASSWORD: os.Getenv("DB_PASSWORD"),
				DB_PORT:     os.Getenv("DB_PORT"),
			},
			Redis: struct{REDIS_HOST string; REDIS_PORT string; REDIS_PASS string; REDIS_DB string; REDIS_USER string}{
				REDIS_HOST: os.Getenv("REDIS_HOST"),
				REDIS_PORT: os.Getenv("REDIS_PORT"),
				REDIS_PASS: os.Getenv("REDIS_PASS"),
				REDIS_DB: os.Getenv("REDIS_DB"),
				REDIS_USER: os.Getenv("REDIS_USER"),
			},
			Kafka: struct {
				KAFKA_BROKER_URL string
				KAFKA_TOPIC      string
				KAFKA_GROUP_ID   string
			}{
				KAFKA_BROKER_URL: os.Getenv("KAFKA_BROKER_URL"),
				KAFKA_TOPIC:      os.Getenv("KAFKA_TOPIC"),
				KAFKA_GROUP_ID:   os.Getenv("KAFKA_GROUP_ID"),
			},
		}
	})

	return Config
}
