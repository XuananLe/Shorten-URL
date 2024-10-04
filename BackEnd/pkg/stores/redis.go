package stores

import (
	"context"
	"shorten-url/backend/pkg/utils"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var once sync.Once

func InitRedis(maxMemory string, evictionStrategy string) *redis.Client {
	once.Do(func() {
		redisConfig := utils.LoadEnv().Redis
		DB, err := strconv.Atoi(redisConfig.REDIS_DB)
		if err != nil {
			log.Fatal(err)
		}
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfig.REDIS_HOST + ":" + redisConfig.REDIS_PORT,
			Password: redisConfig.REDIS_PASS,
			DB:       DB,
		})
		_, err = RedisClient.ConfigSet(context.Background(), "maxmemory", maxMemory).Result()
		if err != nil {
			log.Fatalf("Failed to set Redis maxmemory: %v", err)
		}
		_, err = RedisClient.ConfigSet(context.Background(), "maxmemory-policy", evictionStrategy).Result()
		if err != nil {
			log.Fatalf("Failed to set Redis maxmemory-policy: %v", err)
		}
	})
	return RedisClient
}
