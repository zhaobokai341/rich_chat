package database

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// CachedUserReader decorates UserReader with caching functionality
type CachedUserReader struct {
	reader UserReader
	cache  CacheService
}

// NewCachedUserReader creates a new cached user reader
func NewCachedUserReader(reader UserReader, cache CacheService) *CachedUserReader {
	return &CachedUserReader{
		reader: reader,
		cache:  cache,
	}
}

// FindByID retrieves a user by ID with caching
func (c *CachedUserReader) FindByID(id int) (*User, error) {
	cacheKey := fmt.Sprintf("user:exists:%d", id)
	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "false" {
			log.WithFields(log.Fields{
				"user_id": id,
				"source":  "cache",
			}).Debug("User not found (cached)")
			return nil, fmt.Errorf("user not found")
		}
	}

	user, err := c.reader.FindByID(id)
	if err != nil {
		c.cache.SetWithTTL(cacheKey, "false", CACHE_USER_EXISTS_TTL)
		return nil, err
	}

	c.cache.SetWithTTL(cacheKey, "true", CACHE_USER_EXISTS_TTL)
	return user, nil
}

// FindByUsername retrieves a user by username with caching
func (c *CachedUserReader) FindByUsername(username string) (*User, error) {
	cacheKey := fmt.Sprintf("user:id:username:%s", username)

	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return nil, fmt.Errorf("user not found")
		}
	}

	user, err := c.reader.FindByUsername(username)
	if err != nil {
		c.cache.SetNull(cacheKey)
		c.cache.SetExpiration(cacheKey, time.Duration(CACHE_NULL_TTL)*time.Second)
		return nil, err
	}

	c.cache.SetWithTTL(cacheKey, fmt.Sprintf("%d", user.ID), CACHE_USER_EXISTS_TTL)

	return user, nil
}

// ExistsByID checks if a user exists by ID with caching
func (c *CachedUserReader) ExistsByID(id int) (bool, error) {
	cacheKey := fmt.Sprintf("user:exists:%d", id)
	if cached, found := c.cache.Get(cacheKey); found {
		exists := cached == "true"
		log.WithFields(log.Fields{
			"user_id": id,
			"source":  "cache",
		}).Debug("User existence check")
		return exists, nil
	}

	exists, err := c.reader.ExistsByID(id)
	if err != nil {
		return false, err
	}

	if exists {
		c.cache.SetWithTTL(cacheKey, "true", CACHE_USER_EXISTS_TTL)
	} else {
		c.cache.SetWithTTL(cacheKey, "false", CACHE_USER_EXISTS_TTL)
	}

	return exists, nil
}

// ExistsByUsername checks if a user exists by username with caching
func (c *CachedUserReader) ExistsByUsername(username string) (bool, error) {
	cacheKey := fmt.Sprintf("user:id:username:%s", username)

	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return false, nil
		}

		log.WithFields(log.Fields{
			"username": username,
			"source":   "cache",
		}).Debug("User found (cached)")
		return true, nil
	}

	exists, err := c.reader.ExistsByUsername(username)
	if err != nil {
		return false, err
	}

	if exists {
		c.cache.SetWithTTL(cacheKey, "exists", CACHE_USER_EXISTS_TTL)
	} else {
		c.cache.SetNull(cacheKey)
		c.cache.SetExpiration(cacheKey, time.Duration(CACHE_NULL_TTL)*time.Second)
	}

	return exists, nil
}

// GetUserProfile retrieves user profile with caching
func (c *CachedUserReader) GetUserProfile(userID int) (*UserInfo, error) {
	cacheKey := fmt.Sprintf("user:info:%d", userID)
	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			log.WithFields(log.Fields{
				"user_id": userID,
				"source":  "cache",
			}).Debug("User info not found (cached)")
			return nil, fmt.Errorf("user not found")
		}

		var userInfo UserInfo
		err := json.Unmarshal([]byte(cached), &userInfo)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("Error while unmarshalling user info from cache")
			return nil, err
		}

		log.WithFields(log.Fields{
			"user_id": userID,
			"source":  "cache",
		}).Debug("User info found (cached)")
		return &userInfo, nil
	}

	userInfo, err := c.reader.GetUserProfile(userID)
	if err != nil {
		c.cache.SetNull(cacheKey)
		c.cache.SetExpiration(cacheKey, time.Duration(CACHE_NULL_TTL)*time.Second)
		return nil, err
	}

	userStr, err := json.Marshal(userInfo)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Error while marshalling user info")
		return nil, err
	}

	c.cache.SetWithTTL(cacheKey, string(userStr), CACHE_USER_PROFILE_TTL)
	return userInfo, nil
}

// GetLockStatus retrieves lock status (no caching for security)
func (c *CachedUserReader) GetLockStatus(identifier string) (*time.Time, error) {
	return c.reader.GetLockStatus(identifier)
}

// GetPasswordHash retrieves user's password hash with caching
// This optimizes authentication performance by caching the hash
func (c *CachedUserReader) GetPasswordHash(userID int) (string, error) {
	cacheKey := fmt.Sprintf("user:password_hash:%d", userID)
	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			return "", fmt.Errorf("user not found")
		}
		log.WithFields(log.Fields{
			"user_id": userID,
			"source":  "cache",
		}).Debug("Password hash retrieved from cache")
		return cached, nil
	}

	user, err := c.reader.FindByID(userID)
	if err != nil {
		return "", err
	}

	c.cache.SetWithTTL(cacheKey, user.PasswordHash, CACHE_PASSWORD_HASH_TTL)
	return user.PasswordHash, nil
}

// GetUserBasicInfo retrieves basic user info by ID with caching
func (c *CachedUserReader) GetUserBasicInfo(userID int) (*UserBasicInfo, error) {
	cacheKey := fmt.Sprintf("user:basic:%d", userID)
	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			return nil, fmt.Errorf("user not found")
		}
		var basicInfo UserBasicInfo
		err := json.Unmarshal([]byte(cached), &basicInfo)
		if err != nil {
			return nil, err
		}
		return &basicInfo, nil
	}

	basicInfo, err := c.reader.GetUserBasicInfo(userID)
	if err != nil {
		c.cache.SetNull(cacheKey)
		c.cache.SetExpiration(cacheKey, time.Duration(CACHE_NULL_TTL)*time.Second)
		return nil, err
	}

	basicStr, err := json.Marshal(basicInfo)
	if err != nil {
		return nil, err
	}
	c.cache.SetWithTTL(cacheKey, string(basicStr), CACHE_USER_BASIC_TTL)
	return basicInfo, nil
}
