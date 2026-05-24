package main

import (
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// Helper function to get cached data
func (ud *UserDatabase) getCache(key string) (string, bool) {
	if ud.redis == nil {
		return "", false
	}

	val, err := ud.redis.Get(ud.ctx, key).Result()
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
func (ud *UserDatabase) setCacheWithTTL(key string, value string, ttl int) {
	if ud.redis == nil {
		return
	}

	err := ud.redis.Set(ud.ctx, key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error setting cache")
	}
}

// Helper function to set null cache with short TTL
func (ud *UserDatabase) setNullCache(key string) {
	ud.setCacheWithTTL(key, "", CACHE_NULL_TTL)
}

// Helper function to set cache with default TTL
func (ud *UserDatabase) setCache(key string, value string) {
	ud.setCacheWithTTL(key, value, CACHE_TTL)
}

// Helper function to set cache with lockout duration
func (ud *UserDatabase) setLockoutCache(key string) {
	ud.setCacheWithTTL(key, "", int(LOCKOUT_DURATION))
}

// Helper function to delete cache
func (ud *UserDatabase) deleteCache(key string) {
	if ud.redis == nil {
		return
	}

	err := ud.redis.Del(ud.ctx, key).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error deleting cache")
	}
}

// Helper function to increment counter in Redis
func (ud *UserDatabase) incrementCounter(key string) int64 {
	if ud.redis == nil {
		return 0
	}

	val, err := ud.redis.Incr(ud.ctx, key).Result()
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
func (ud *UserDatabase) setKeyExpiration(key string, ttl time.Duration) {
	if ud.redis == nil {
		return
	}

	err := ud.redis.Expire(ud.ctx, key, ttl).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warning("Error setting key expiration")
	}
}

// Helper function to get integer value from Redis
func (ud *UserDatabase) getIntValue(key string) (int64, bool) {
	if ud.redis == nil {
		return 0, false
	}

	val, err := ud.redis.Get(ud.ctx, key).Int64()
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
