package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// Generate a JWT token for a given user ID
func generateToken(user_id int, valid_time time.Duration) (string, error) {
	expirationTime := time.Now().Add(valid_time)

	claims := &Claims{
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rich_chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Generate a random verification token
func generateVerifyToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Store verification token in Redis
func (ud *UserDatabase) storeVerifyToken(token string) error {
	if ud.redis == nil {
		return fmt.Errorf("Redis not available")
	}

	key := fmt.Sprintf("verify_token:%s", token)
	ud.setCacheWithTTL(key, "valid", int(VERIFY_TOKEN_EXPIRE_TIME.Seconds()))

	log.WithFields(log.Fields{
		"token": token,
	}).Debug("Verification token stored")
	return nil
}

// Verify and consume verification token
func (ud *UserDatabase) verifyAndConsumeToken(token string) bool {
	if ud.redis == nil {
		return false
	}

	key := fmt.Sprintf("verify_token:%s", token)

	// Check if token exists
	val, found := ud.getCache(key)
	if !found {
		// Token not found or expired
		log.WithFields(log.Fields{
			"token": token,
		}).Warning("Verification token not found or expired")
		return false
	}

	// Delete the token (one-time use)
	ud.deleteCache(key)

	return val == "valid"
}

// Track login attempts for rate limiting (Redis + PostgreSQL)
func (ud *UserDatabase) trackLoginAttempt(identifier string, success bool) {
	// Always update Redis first for fast access
	if ud.redis != nil {
		attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
		lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)

		if success {
			// Reset attempts on successful login
			ud.deleteCache(attemptsKey)
		} else {
			// Increment failed attempts in Redis
			attempts := ud.incrementCounter(attemptsKey)
			ud.setKeyExpiration(attemptsKey, LOCKOUT_DURATION)

			// Check if max attempts reached
			if attempts >= MAX_LOGIN_ATTEMPTS {
				// Lock the account in Redis
				ud.setLockoutCache(lockoutKey)
				log.WithFields(log.Fields{
					"identifier": identifier,
					"attempts":   attempts,
					"source":     "redis",
				}).Warning("Account locked due to too many failed attempts")
			}
		}
	}

	// Also persist lock status to PostgreSQL for durability
	if success {
		// Reset lock in database
		_, err := ud.database.Exec(
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
		if ud.redis != nil {
			attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
			attempts, found := ud.getIntValue(attemptsKey)
			if found && attempts >= MAX_LOGIN_ATTEMPTS {
				// Lock the account in database
				_, err := ud.database.Exec(
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
}

// Check if account is locked (Redis + PostgreSQL)
func (ud *UserDatabase) isAccountLocked(identifier string) bool {
	// Check Redis first for fast access
	if ud.redis != nil {
		lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)
		_, found := ud.getCache(lockoutKey)
		if found {
			// Locked in Redis
			return true
		}
	}

	// Check PostgreSQL for persistent state
	var lockUntil interface{}
	err := ud.database.QueryRow(
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
	err = ud.database.QueryRow(
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
		_, err = ud.database.Exec(
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
		if ud.redis != nil {
			lockoutKey := fmt.Sprintf("login_lockout:%s", identifier)
			attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
			ud.deleteCache(lockoutKey)
			ud.deleteCache(attemptsKey)
		}

		return false
	}

	// Lock is still active
	return true
}

// Get remaining login attempts (Redis primary, PostgreSQL fallback)
func (ud *UserDatabase) getRemainingAttempts(identifier string) int64 {
	// Try Redis first
	if ud.redis != nil {
		attemptsKey := fmt.Sprintf("login_attempts:%s", identifier)
		attempts, found := ud.getIntValue(attemptsKey)
		if found {
			remaining := MAX_LOGIN_ATTEMPTS - attempts
			if remaining < 0 {
				return 0
			}
			return remaining
		}
	}

	// Fallback to PostgreSQL - check if locked
	var lockUntil interface{}
	err := ud.database.QueryRow(
		"SELECT lock_until FROM users WHERE username = $1 OR id::text = $1",
		identifier,
	).Scan(&lockUntil)

	if err != nil {
		log.WithFields(log.Fields{
			"identifier": identifier,
			"error":      err.Error(),
		}).Warning("Failed to get lock status from database")
		return MAX_LOGIN_ATTEMPTS
	}

	// If account is locked, return 0
	if lockUntil != nil {
		return 0
	}

	// If not locked but Redis unavailable, assume fresh start
	return MAX_LOGIN_ATTEMPTS
}
