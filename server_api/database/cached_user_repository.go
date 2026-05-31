package database

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// CachedUserRepository decorates UserRepository with caching functionality
type CachedUserRepository struct {
	repo  UserRepository
	cache CacheService
}

// NewCachedUserRepository creates a new cached user repository
func NewCachedUserRepository(repo UserRepository, cache CacheService) *CachedUserRepository {
	return &CachedUserRepository{
		repo:  repo,
		cache: cache,
	}
}

// CreateUser creates a user and invalidates related caches
func (c *CachedUserRepository) CreateUser(username, passwordHash string) (int, error) {
	userID, err := c.repo.CreateUser(username, passwordHash)
	if err != nil {
		return 0, err
	}

	// Invalidate related caches
	c.cache.Delete(fmt.Sprintf("user:exists:%d", userID))
	c.cache.Delete(fmt.Sprintf("user:id:username:%s", username))

	// Cache the password hash for future authentication
	hashCacheKey := fmt.Sprintf("user:hash:%d", userID)
	c.cache.Set(hashCacheKey, passwordHash)

	return userID, nil
}

// FindByID retrieves a user by ID with caching
func (c *CachedUserRepository) FindByID(id int) (*User, error) {
	// Check cache first
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

	// Cache miss - query database
	user, err := c.repo.FindByID(id)
	if err != nil {
		// Cache negative result with short TTL
		c.cache.Set(cacheKey, "false")
		return nil, err
	}

	// Cache positive result
	c.cache.Set(cacheKey, "true")
	return user, nil
}

// FindByUsername retrieves a user by username with caching
func (c *CachedUserRepository) FindByUsername(username string) (*User, error) {
	// Check cache for user_id by username
	cacheKey := fmt.Sprintf("user:id:username:%s", username)

	if cached, found := c.cache.Get(cacheKey); found {
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return nil, fmt.Errorf("user not found")
		}

		log.WithFields(log.Fields{
			"username": username,
			"source":   "cache",
		}).Debug("User found (cached)")
	}

	// Query database
	user, err := c.repo.FindByUsername(username)
	if err != nil {
		// Cache negative result
		c.cache.SetNull(cacheKey)
		return nil, err
	}

	// Cache the user_id
	c.cache.Set(cacheKey, fmt.Sprintf("%d", user.ID))

	// Also cache the password hash
	hashCacheKey := fmt.Sprintf("user:hash:%d", user.ID)
	c.cache.Set(hashCacheKey, user.PasswordHash)

	return user, nil
}

// ExistsByID checks if a user exists by ID with caching
func (c *CachedUserRepository) ExistsByID(id int) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("user:exists:%d", id)
	if cached, found := c.cache.Get(cacheKey); found {
		exists := cached == "true"
		log.WithFields(log.Fields{
			"user_id": id,
			"source":  "cache",
		}).Debug("User existence check")
		return exists, nil
	}

	// Cache miss - query database
	exists, err := c.repo.ExistsByID(id)
	if err != nil {
		return false, err
	}

	// Cache the result
	if exists {
		c.cache.Set(cacheKey, "true")
	} else {
		c.cache.Set(cacheKey, "false")
	}

	return exists, nil
}

// ExistsByUsername checks if a user exists by username with caching
func (c *CachedUserRepository) ExistsByUsername(username string) (bool, error) {
	// Check cache first
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

	// Cache miss - query database
	exists, err := c.repo.ExistsByUsername(username)
	if err != nil {
		return false, err
	}

	// Cache the result
	if exists {
		c.cache.Set(cacheKey, "exists")
	} else {
		c.cache.SetNull(cacheKey)
	}

	return exists, nil
}

// GetUserProfile retrieves user profile with caching
func (c *CachedUserRepository) GetUserProfile(userID int) (*UserInfo, error) {
	// Check cache first
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

	// Cache miss - query database
	userInfo, err := c.repo.GetUserProfile(userID)
	if err != nil {
		c.cache.SetNull(cacheKey)
		return nil, err
	}

	// Cache the user info
	userStr, err := json.Marshal(userInfo)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Error while marshalling user info")
		return nil, err
	}

	c.cache.Set(cacheKey, string(userStr))
	return userInfo, nil
}

// UpdateProfile updates user profile and invalidates cache
func (c *CachedUserRepository) UpdateProfile(userID int, key, value string) error {
	err := c.repo.UpdateProfile(userID, key, value)
	if err != nil {
		return err
	}

	// Delete the user info cache
	c.cache.Delete(fmt.Sprintf("user:info:%d", userID))
	return nil
}

// UpdatePassword updates user password and invalidates cache
func (c *CachedUserRepository) UpdatePassword(userID int, newPasswordHash string) error {
	err := c.repo.UpdatePassword(userID, newPasswordHash)
	if err != nil {
		return err
	}

	// Invalidate password hash cache
	c.cache.Delete(fmt.Sprintf("user:hash:%d", userID))
	return nil
}

// UpdateLastLogin updates last login timestamp
func (c *CachedUserRepository) UpdateLastLogin(userID int) error {
	return c.repo.UpdateLastLogin(userID)
}

// UpdateLockStatus updates lock status
func (c *CachedUserRepository) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	return c.repo.UpdateLockStatus(identifier, lockUntil)
}

// DeleteUser deletes a user and invalidates all related caches
func (c *CachedUserRepository) DeleteUser(userID int) error {
	// Get username before deletion to invalidate cache
	user, err := c.repo.FindByID(userID)
	if err == nil && user != nil {
		// Invalidate all related caches
		c.cache.Delete(fmt.Sprintf("user:exists:%d", userID))
		c.cache.Delete(fmt.Sprintf("user:id:username:%s", user.Username))
		c.cache.Delete(fmt.Sprintf("user:hash:%d", userID))
		c.cache.Delete(fmt.Sprintf("user:info:%d", userID))
	}

	// Delete user from database
	err = c.repo.DeleteUser(userID)
	if err != nil {
		return err
	}

	return nil
}

// GetLockStatus retrieves lock status
func (c *CachedUserRepository) GetLockStatus(identifier string) (*time.Time, error) {
	return c.repo.GetLockStatus(identifier)
}

// ClearExpiredLock clears expired lock
func (c *CachedUserRepository) ClearExpiredLock(identifier string) error {
	return c.repo.ClearExpiredLock(identifier)
}
