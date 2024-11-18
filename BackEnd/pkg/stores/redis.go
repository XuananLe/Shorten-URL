package stores

import (
	"context"
	"fmt"
	"shorten-url/backend/pkg/config"
	"time"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var RedisCluster *redis.ClusterClient

func InitRedis(maxMemory string, evictionStrategy string) *redis.ClusterClient {
	redisConfig := config.AppConfig.Redis.ClusterNodes

	RedisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          redisConfig,
		RouteByLatency: true,
		ReadOnly:       true,
		MaxRedirects: 	3,
		PoolSize:       200,               
		MinIdleConns:   10,                 
		DialTimeout:    3 * time.Second,   
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		RouteRandomly: true,
	})

	ctx := context.Background()

	_, err := RedisCluster.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis Cluster: %v", err)
	}

	go func() {
		err = setClusterConfig(ctx, RedisCluster, "maxmemory", maxMemory)
		if err != nil {
			log.Fatalf("Failed to set Redis Cluster maxmemory: %v", err)
		}

		err = setClusterConfig(ctx, RedisCluster, "maxmemory-policy", evictionStrategy)
		if err != nil {
			log.Fatalf("Failed to set Redis Cluster maxmemory-policy: %v", err)
		}
	}()

	fmt.Println("Redis Cluster Connected")
	return RedisCluster
}


func setClusterConfig(ctx context.Context, client *redis.ClusterClient, key, value string) error {
	var firstErr error
	err := client.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		_, err := shard.ConfigSet(ctx, key, value).Result()
		if err != nil && firstErr == nil {
			firstErr = err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return firstErr
}
