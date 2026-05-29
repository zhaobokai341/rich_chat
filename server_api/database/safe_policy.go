package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Verify and consume verification token
func (ud *UserDatabase) VerifyAndConsumeToken(token string) bool {
	key := fmt.Sprintf("verify_token:%s", token)

	// Check if token exists
	val, found := ud.Redis_manager.GetCache(key)
	if !found {
		// Token not found or expired
		log.WithFields(log.Fields{
			"token": token,
		}).Warning("Verification token not found or expired")
		return false
	}

	// Delete the token (one-time use)
	ud.Redis_manager.DeleteCache(key)

	return val == "valid"
}

// Track login attempts for rate limiting (Redis + PostgreSQL)
func (ud *UserDatabase) TrackLoginAttempt(identifier string, success bool) {
	// Always update Redis first for fast access
	attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
	lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)

	if success {
		// Reset attempts on successful login
		ud.Redis_manager.DeleteCache(attemptsKey)
	} else {
		// Increment failed attempts in Redis
		attempts := ud.Redis_manager.IncrementCounter(attemptsKey)
		ud.Redis_manager.SetKeyExpiration(attemptsKey, ud.Cfg.LOCKOUT_DURATION)

		// Check if max attempts reached
		if attempts >= ud.Cfg.MAX_LOGIN_ATTEMPTS {
			// Lock the account in Redis
			ud.Redis_manager.SetLockoutCache(lockoutKey)
			log.WithFields(log.Fields{
				"identifier": identifier,
				"attempts":   attempts,
				"source":     "redis",
			}).Warning("Account locked due to too many failed attempts")
		}
	}

	// Also persist lock status to PostgreSQL for durability
	if success {
		// Reset lock in database
		_, err := ud.Database.Exec(
			"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
			identifier,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			}).Warning("Failed to reset lock status in database")
		} else {
			log.WithFields(log.Fields{
				"identifier": identifier,
				"source":     "postgresql",
			}).Debug("Reset lock status in database")
		}
	} else {
		// Check if we need to lock based on Redis attempts
		attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
		attempts, found := ud.Redis_manager.GetIntValue(attemptsKey)
		if found && attempts >= ud.Cfg.MAX_LOGIN_ATTEMPTS {
			// Lock the account in database
			_, err := ud.Database.Exec(
				`UPDATE users
						SET lock_until = NOW() + INTERVAL '15 minutes'
						WHERE username = $1
							OR id::text = $1`,
				identifier,
			)
			if err != nil {
				log.WithFields(log.Fields{
					"identifier": identifier,
					"error":      err.Error(),
				}).Warning("Failed to lock account in database")
			} else {
				log.WithFields(log.Fields{
					"identifier": identifier,
					"attempts":   attempts,
					"source":     "postgresql",
				}).Warning("Account locked in database due to too many failed attempts")
			}
		}
	}
}

// Check if account is locked (Redis + PostgreSQL)
func (ud *UserDatabase) IsAccountLocked(identifier string) bool {
	// Check Redis first for fast access
	lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)
	_, found := ud.Redis_manager.GetCache(lockoutKey)
	if found {
		// Locked in Redis
		return true
	}

	// Check PostgreSQL for persistent state
	var lockUntil interface{}
	err := ud.Database.QueryRow(
		"SELECT lock_until FROM users WHERE username = $1 OR id::text = $1",
		identifier,
	).Scan(&lockUntil)

	if err != nil {
		log.WithFields(log.Fields{
			"identifier": identifier,
			"error":      err.Error(),
		}).Warning("Failed to check lock status in database")
		return false
	}

	// If lock_until is NULL, account is not locked
	if lockUntil == nil {
		return false
	}

	// Check if lock has expired by comparing with current time
	var lockExpired bool
	err = ud.Database.QueryRow(
		"SELECT lock_until < NOW() FROM users WHERE username = $1 OR id::text = $1",
		identifier,
	).Scan(&lockExpired)

	if err != nil {
		log.WithFields(log.Fields{
			"identifier": identifier,
			"error":      err.Error(),
		}).Warning("Failed to check lock expiration in database")
		return true // Assume locked if we can't verify
	}

	// If lock has expired, clear it
	if lockExpired {
		_, err = ud.Database.Exec(
			"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
			identifier,
		)
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
		lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)
		attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
		ud.Redis_manager.DeleteCache(lockoutKey)
		ud.Redis_manager.DeleteCache(attemptsKey)

		return false
	}

	// Lock is still active
	return true
}

// Helper function to track IP visit count in Redis
func (ud *UserDatabase) TrackIPVisit(ip string) int64 {
	key := fmt.Sprintf("ip_visit:%s", ip)

	// Increment counter
	count := ud.Redis_manager.IncrementCounter(key)

	// Set expiration on first visit
	if count == 1 {
		ud.Redis_manager.SetKeyExpiration(key, ud.Cfg.IP_LIMIT_TIME)
	}

	return count
}

// Helper function to check if IP is blocked (Redis cache + database)
func (ud *UserDatabase) IsIPBlocked(ip string) bool {
	// Check Redis cache first for fast access
	cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
	cachedValue, found := ud.Redis_manager.GetCache(cacheKey)

	if found {
		// Cache hit - IP is blocked
		if cachedValue == "blocked" {
			log.WithFields(log.Fields{
				"ip":     ip,
				"source": "redis_cache",
			}).Debug("IP block status retrieved from cache")
			return true
		}
		// Cache hit - IP is not blocked (null cache)
		return false
	}

	// Cache miss - query database
	var until_time time.Time
	err := ud.Database.QueryRow(
		`SELECT 
			block_until 
		FROM
			ip_block_list
		WHERE
			ip = $1
			AND block_until < NOW()
		ORDER BY
			created_time DESC
		LIMIT 1`,
		ip,
	).Scan(&until_time)

	if err != nil {
		// No blocking record found or error
		// Cache the negative result with short TTL
		cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
		ud.Redis_manager.SetNullCache(cacheKey)
		return false
	}

	// Check if the block has expired
	if until_time.After(time.Now()) {
		cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
		ud.Redis_manager.SetCacheWithTTL(
			cacheKey,
			"blocked",
			int(ud.Cfg.IP_LIMIT_LOCKOUT_DURATION),
		)
		return true
	}

	_, err = ud.Database.Exec(
		"DELETE FROM ip_block_list WHERE ip = $1", ip,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"ip":    ip,
			"error": err.Error(),
		}).Error("Failed to delete IP block record from database")
	}

	return false

}

// Helper function to block IP in database and update cache
func (ud *UserDatabase) BlockIP(ip string, reason string, duration time.Duration) error {
	blockUntil := time.Now().Add(duration)

	_, err := ud.Database.Exec(
		"INSERT INTO ip_block_list (ip, block_until, reason) VALUES ($1, $2, $3)",
		ip, blockUntil, reason,
	)

	if err != nil {
		log.WithFields(log.Fields{
			"ip":     ip,
			"reason": reason,
			"error":  err.Error(),
		}).Error("Failed to block IP in database")
		return err
	}

	// Update Redis cache immediately
	cacheKey := fmt.Sprintf("ip_blocked:%s", ip)
	ttlSeconds := int(duration.Seconds())
	ud.Redis_manager.SetCacheWithTTL(cacheKey, "blocked", ttlSeconds)

	// Also delete the negative cache if it exists
	negativeCacheKey := fmt.Sprintf("ip_not_blocked:%s", ip)
	ud.Redis_manager.DeleteCache(negativeCacheKey)

	log.WithFields(log.Fields{
		"ip":       ip,
		"cacheKey": cacheKey,
		"ttl":      ttlSeconds,
		"source":   "redis_cache_updated",
	}).Debug("IP block status updated in cache")

	log.WithFields(log.Fields{
		"ip":         ip,
		"blockUntil": blockUntil,
		"reason":     reason,
	}).Warning("IP blocked in database")

	return nil
}
