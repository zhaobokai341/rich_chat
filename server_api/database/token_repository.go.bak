package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// RedisTokenRepository implements TokenRepository using Redis
type RedisTokenRepository struct {
	cache CacheService
	ttl   time.Duration
}

// NewRedisTokenRepository creates a new token repository
func NewRedisTokenRepository(cache CacheService, ttl time.Duration) *RedisTokenRepository {
	return &RedisTokenRepository{
		cache: cache,
		ttl:   ttl,
	}
}

// StoreVerifyToken stores a verification token in Redis
func (r *RedisTokenRepository) StoreVerifyToken(token string, ttl time.Duration) error {
	key := fmt.Sprintf("verify_token:%s", token)
	ttlSeconds := int(ttl.Seconds())
	r.cache.SetWithTTL(key, "valid", ttlSeconds)

	log.WithFields(log.Fields{
		"token": token,
		"ttl":   ttl,
	}).Debug("Verification token stored")

	return nil
}

// VerifyAndConsumeToken verifies and consumes a verification token
func (r *RedisTokenRepository) VerifyAndConsumeToken(token string) (bool, error) {
	key := fmt.Sprintf("verify_token:%s", token)

	// Check if token exists
	val, found := r.cache.Get(key)
	if !found {
		// Token not found or expired
		log.WithFields(log.Fields{
			"token": token,
		}).Warning("Verification token not found or expired")
		return false, nil
	}

	// Delete the token (one-time use)
	r.cache.Delete(key)

	return val == "valid", nil
}
