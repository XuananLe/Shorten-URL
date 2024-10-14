package utils

import (
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Server struct {
		ServerPort []string
	}
	Database struct {
		DbName     string
		DbUsername string
		DbPassword string
		DbPort     string
	}
	Redis struct {
		RedisHost string
		RedisPort string
		RedisPass string
		RedisDb   string
		RedisUser string
		RedisClusterNodes []string
	}
	Kafka struct {
		KafkaBrokerUrl string
		KafkaTopic     string
		KafkaGroupId   string
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

		serverPorts := strings.Split(os.Getenv("SERVER_PORT"), ",")
		redisClusterNodes := strings.Split(os.Getenv("REDIS_CLUSTER_NODES"), ",")
		for i := range redisClusterNodes{
			redisClusterNodes[i] = ":" + redisClusterNodes[i]
		}

		Config = &config{
			Server: struct {
				ServerPort []string
			}{
				ServerPort: serverPorts,
			},
			Database: struct {
				DbName     string
				DbUsername string
				DbPassword string
				DbPort     string
			}{
				DbName:     os.Getenv("DB_NAME"),
				DbUsername: os.Getenv("DB_USERNAME"),
				DbPassword: os.Getenv("DB_PASSWORD"),
				DbPort:     os.Getenv("DB_PORT"),
			},
			Redis: struct {
				RedisHost string
				RedisPort string
				RedisPass string
				RedisDb   string
				RedisUser string
				RedisClusterNodes []string
			}{
				RedisHost: os.Getenv("REDIS_HOST"),
				RedisPort: os.Getenv("REDIS_PORT"),
				RedisPass: os.Getenv("REDIS_PASS"),
				RedisDb:   os.Getenv("REDIS_DB"),
				RedisUser: os.Getenv("REDIS_USER"),
				RedisClusterNodes: redisClusterNodes,
			},
			Kafka: struct {
				KafkaBrokerUrl string
				KafkaTopic     string
				KafkaGroupId   string
			}{
				KafkaBrokerUrl: os.Getenv("KAFKA_BROKER_URL"),
				KafkaTopic:     os.Getenv("KAFKA_TOPIC"),
				KafkaGroupId:   os.Getenv("KAFKA_GROUP_ID"),
			},
		}
	})

	return Config
}
