package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/basketikun/infinite-canvas/config"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func Redis() *redis.Client {
	redisOnce.Do(func() {
		addr := strings.TrimSpace(config.Cfg.RedisAddr)
		if addr == "" {
			return
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: config.Cfg.RedisPassword,
			DB:       config.Cfg.RedisDB,
		})
	})
	return redisClient
}

func CacheTTL() time.Duration {
	seconds := config.Cfg.RedisCacheTTL
	if seconds <= 0 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

func CacheKey(parts ...string) string {
	prefix := strings.Trim(config.Cfg.RedisKeyPrefix, ": ")
	if prefix == "" {
		prefix = "infinite-canvas"
	}
	return prefix + ":" + strings.Join(parts, ":")
}

func cacheContext() context.Context {
	return context.Background()
}

func CacheSetNX(key string, value string, ttl time.Duration) bool {
	client := Redis()
	if client == nil {
		return true
	}
	ok, err := client.SetNX(cacheContext(), key, value, ttl).Result()
	return err == nil && ok
}

func CacheDelete(key string) {
	if client := Redis(); client != nil {
		_ = client.Del(cacheContext(), key).Err()
	}
}
