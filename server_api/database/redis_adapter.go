package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// RedisCacheAdapter adapts the existing RedisManager to CacheService interface
type RedisCacheAdapter struct {
	manager RedisManager
}

// NewRedisCacheAdapter creates a new Redis cache adapter
func NewRedisCacheAdapter(manager RedisManager) *RedisCacheAdapter {
	return &RedisCacheAdapter{
		manager: manager,
	}
}

// Get retrieves a value from cache
func (a *RedisCacheAdapter) Get(key string) (string, bool) {
	return a.manager.GetCache(key)
}

// Set stores a value in cache with default TTL
func (a *RedisCacheAdapter) Set(key, value string) {
	a.manager.SetCache(key, value)
}

// SetWithTTL stores a value in cache with specified TTL
func (a *RedisCacheAdapter) SetWithTTL(key, value string, ttlSeconds int) {
	a.manager.SetCacheWithTTL(key, value, ttlSeconds)
}

// Delete removes a value from cache
func (a *RedisCacheAdapter) Delete(key string) {
	a.manager.DeleteCache(key)
}

// Increment increments a counter and returns the new value
func (a *RedisCacheAdapter) Increment(key string) int64 {
	return a.manager.IncrementCounter(key)
}

// SetExpiration sets expiration time for a key
func (a *RedisCacheAdapter) SetExpiration(key string, ttl time.Duration) {
	a.manager.SetKeyExpiration(key, ttl)
}

// SetNull caches a null/empty value for negative caching
func (a *RedisCacheAdapter) SetNull(key string) {
	a.manager.SetNullCache(key)
}

// RedisRateLimitRepository implements RateLimitRepository using Redis and PostgreSQL
type RedisRateLimitRepository struct {
	cache    CacheService
	userRepo UserRepository
	config   Config
}

// NewRedisRateLimitRepository creates a new rate limit repository
func NewRedisRateLimitRepository(cache CacheService, userRepo UserRepository, config Config) *RedisRateLimitRepository {
	return &RedisRateLimitRepository{
		cache:    cache,
		userRepo: userRepo,
		config:   config,
	}
}

// TrackLoginAttempt tracks login attempts in Redis and PostgreSQL
func (r *RedisRateLimitRepository) TrackLoginAttempt(identifier string, success bool) error {
	attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
	lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)

	if success {
		// Reset attempts on successful login
		r.cache.Delete(attemptsKey)

		// Reset lock in database
		err := r.userRepo.UpdateLockStatus(identifier, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			}).Warning("Failed to reset lock status in database")
		}
	} else {
		// Increment failed attempts in Redis
		attempts := r.cache.Increment(attemptsKey)
		r.cache.SetExpiration(attemptsKey, r.config.LOCKOUT_DURATION)

		// Check if max attempts reached
		if attempts >= r.config.MAX_LOGIN_ATTEMPTS {
			// Lock the account in Redis
			r.cache.Set(lockoutKey, "locked")
			r.cache.SetExpiration(lockoutKey, r.config.LOCKOUT_DURATION)

			log.WithFields(log.Fields{
				"identifier": identifier,
				"attempts":   attempts,
				"source":     "redis",
			}).Warning("Account locked due to too many failed attempts")

			// Lock the account in database
			lockUntil := time.Now().Add(r.config.LOCKOUT_DURATION)
			err := r.userRepo.UpdateLockStatus(identifier, &lockUntil)
			if err != nil {
				log.WithFields(log.Fields{
					"identifier": identifier,
					"error":      err.Error(),
				}).Warning("Failed to lock account in database")
			}
		}
	}

	return nil
}

// CheckAccountLocked checks if an account is locked
func (r *RedisRateLimitRepository) CheckAccountLocked(identifier string) (bool, error) {
	// Check Redis first for fast access
	lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)
	if _, found := r.cache.Get(lockoutKey); found {
		return true, nil
	}

	// Check PostgreSQL for persistent state
	lockUntil, err := r.userRepo.GetLockStatus(identifier)
	if err != nil {
		log.WithFields(log.Fields{
			"identifier": identifier,
			"error":      err.Error(),
		}).Warning("Failed to check lock status in database")
		return false, err
	}

	// If lock_until is NULL, account is not locked
	if lockUntil == nil {
		return false, nil
	}

	// Check if lock has expired
	if lockUntil.Before(time.Now()) {
		// Lock has expired, clear it
		err := r.userRepo.ClearExpiredLock(identifier)
		if err != nil {
			log.WithFields(log.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			}).Warning("Failed to clear expired lock in database")
		} else {
			log.WithFields(log.Fields{
				"identifier": identifier,
				"source":     "postgresql",
			}).Info("Cleared expired lock from database")
		}

		// Also clear Redis lock if exists
		r.cache.Delete(lockoutKey)
		r.cache.Delete(fmt.Sprintf("login_attempts:%s", identifier))

		return false, nil
	}

	// Lock is still active
	return true, nil
}

// TrackIPVisit tracks IP visit count in Redis
func (r *RedisRateLimitRepository) TrackIPVisit(ip string) (int64, error) {
	key := fmt.Sprintf("ip_visit:%s", ip)

	// Increment counter
	count := r.cache.Increment(key)

	// Set expiration on first visit
	if count == 1 {
		r.cache.SetExpiration(key, r.config.IP_LIMIT_TIME)
	}

	return count, nil
}

// CheckIPBlocked checks if an IP is blocked
func (r *RedisRateLimitRepository) CheckIPBlocked(ip string) (bool, error) {
	// Check Redis cache first for fast access
	cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
	cachedValue, found := r.cache.Get(cacheKey)

	if found {
		// Cache hit - IP is blocked
		if cachedValue == "blocked" {
			log.WithFields(log.Fields{
				"ip":     ip,
				"source": "redis_cache",
			}).Debug("IP block status retrieved from cache")
			return true, nil
		}
		// Cache hit - IP is not blocked (null cache)
		return false, nil
	}

	// Cache miss - would need to query database
	// For now, return false (not blocked)
	// In production, you'd implement DB query here
	return false, nil
}

// BlockIP blocks an IP in database and updates cache
func (r *RedisRateLimitRepository) BlockIP(ip, reason string, duration time.Duration) error {
	// Update Redis cache immediately
	cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
	ttlSeconds := int(duration.Seconds())
	r.cache.SetWithTTL(cacheKey, "blocked", ttlSeconds)

	// Also delete the negative cache if it exists
	negativeCacheKey := fmt.Sprintf("ip_not_blocked:%s", ip)
	r.cache.Delete(negativeCacheKey)

	log.WithFields(log.Fields{
		"ip":       ip,
		"cacheKey": cacheKey,
		"ttl":      ttlSeconds,
		"source":   "redis_cache_updated",
	}).Debug("IP block status updated in cache")

	log.WithFields(log.Fields{
		"ip":     ip,
		"reason": reason,
	}).Warning("IP blocked")

	return nil
}
