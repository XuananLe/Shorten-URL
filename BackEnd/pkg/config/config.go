package config

import (
	"os"
	"strings"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Ports []string
}

type DatabaseConfig struct {
	Host     string
	Name     string
	Username string
	Password string
	Port     string
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           string
	User         string
	SingleNodes  string
	ClusterNodes []string
}

type KafkaConfig struct {
	BrokerURL string
	Topic     string
	GroupID   string
}

var AppConfig Config

func LoadEnv() *Config {
	os.Clearenv()
	err := godotenv.Load("../../.env")	
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	AppConfig = Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		Redis:    loadRedisConfig(),
		Kafka:    loadKafkaConfig(),
	}

	return &AppConfig
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Ports: strings.Split(os.Getenv("SERVER_PORT"), ","),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Name:     os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Port:     os.Getenv("DB_PORT"),
	}
}

func loadRedisConfig() RedisConfig {
	clusterNodes := strings.Split(os.Getenv("REDIS_CLUSTER_NODES"), ",")
	for i := range clusterNodes {
		clusterNodes[i] = ":" + clusterNodes[i]
	}

	return RedisConfig{
		Host:         os.Getenv("REDIS_HOST"),
		Port:         os.Getenv("REDIS_PORT"),
		Password:     os.Getenv("REDIS_PASS"),
		DB:           os.Getenv("REDIS_DB"),
		User:         os.Getenv("REDIS_USER"),
		ClusterNodes: clusterNodes,
	}
}

func loadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		BrokerURL: os.Getenv("KAFKA_BROKER_URL"),
		Topic:     os.Getenv("KAFKA_TOPIC"),
		GroupID:   os.Getenv("KAFKA_GROUP_ID"),
	}
}