package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type RedisManager struct {
	redis *redis.Client
	ctx   context.Context
}

// Initialize Redis client
func redis_init() *RedisManager {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", REDIS_HOST, REDIS_PORT),
		Password: REDIS_PASSWORD,
		DB:       REDIS_DB,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Redis connection failed.")

	}

	log.Info("Redis connection established")

	return &RedisManager{
		redis: rdb,
		ctx:   ctx,
	}
}

// Helper function to get cached data
func (rdb *RedisManager) GetCache(key string) (string, bool) {
	val, err := rdb.redis.Get(rdb.ctx, key).Result()
	if err == redis.Nil {
		// Key not found
		return "", false
	} else if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error getting cache")
		return "", false
	}

	return val, true
}

// Helper function to set cache with custom TTL
func (rdb *RedisManager) SetCacheWithTTL(key string, value string, ttl int) {
	err := rdb.redis.Set(rdb.ctx, key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error setting cache")
	}
}

// Helper function to set null cache with short TTL
func (rdb *RedisManager) SetNullCache(key string) {
	rdb.SetCacheWithTTL(key, "", CACHE_NULL_TTL)
}

// Helper function to set cache with default TTL
func (rdb *RedisManager) SetCache(key string, value string) {
	rdb.SetCacheWithTTL(key, value, CACHE_TTL)
}

// Helper function to set cache with lockout duration
func (rdb *RedisManager) SetLockoutCache(key string) {
	rdb.SetCacheWithTTL(key, "", int(LOCKOUT_DURATION.Seconds()))
}

// Helper function to delete cache
func (rdb *RedisManager) DeleteCache(key string) {
	err := rdb.redis.Del(rdb.ctx, key).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error deleting cache")
	}
}

// Helper function to increment counter in Redis
func (rdb *RedisManager) IncrementCounter(key string) int64 {
	val, err := rdb.redis.Incr(rdb.ctx, key).Result()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error incrementing counter")
		return 0
	}

	return val
}

// Helper function to set expiration on a key
func (rdb *RedisManager) SetKeyExpiration(key string, ttl time.Duration) {
	err := rdb.redis.Expire(rdb.ctx, key, ttl).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error setting key expiration")
	}
}

// Helper function to get integer value from Redis
func (rdb *RedisManager) GetIntValue(key string) (int64, bool) {
	val, err := rdb.redis.Get(rdb.ctx, key).Int64()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error getting integer value from cache")
		return 0, false
	}

	return val, true
}

// Store verification token in Redis
func (rdb *RedisManager) StoreVerifyToken(token string) error {
	key := fmt.Sprintf("verify_token:%s", token)
	rdb.SetCacheWithTTL(key, "valid", int(VERIFY_TOKEN_EXPIRE_TIME.Seconds()))

	log.WithFields(log.Fields{
		"token": token,
	}).Debug("Verification token stored")
	return nil
}
