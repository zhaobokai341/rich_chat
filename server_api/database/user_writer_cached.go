package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// CachedUserWriter decorates UserWriter with cache invalidation
type CachedUserWriter struct {
	writer UserWriter
	cache  CacheService
	reader UserReader // Needed to get username before deletion
}

// NewCachedUserWriter creates a new cached user writer
func NewCachedUserWriter(writer UserWriter, cache CacheService, reader UserReader) *CachedUserWriter {
	return &CachedUserWriter{
		writer: writer,
		cache:  cache,
		reader: reader,
	}
}

// CreateUser creates a user and invalidates related caches
func (c *CachedUserWriter) CreateUser(username, passwordHash string) (int, error) {
	userID, err := c.writer.CreateUser(username, passwordHash)
	if err != nil {
		return 0, err
	}

	// Invalidate related caches
	c.cache.Delete(fmt.Sprintf("user:exists:%d", userID))
	c.cache.Delete(fmt.Sprintf("user:id:username:%s", username))

	return userID, nil
}

// UpdateProfile updates user profile and invalidates cache
func (c *CachedUserWriter) UpdateProfile(userID int, key, value string) error {
	err := c.writer.UpdateProfile(userID, key, value)
	if err != nil {
		return err
	}

	// Delete the user info cache
	c.cache.Delete(fmt.Sprintf("user:info:%d", userID))
	c.cache.Delete(fmt.Sprintf("user:basic:%d", userID))
	return nil
}

// UpdateLastLogin updates last login timestamp (no cache invalidation needed)
func (c *CachedUserWriter) UpdateLastLogin(userID int) error {
	return c.writer.UpdateLastLogin(userID)
}

// UpdateLockStatus updates lock status and invalidates cache
func (c *CachedUserWriter) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	err := c.writer.UpdateLockStatus(identifier, lockUntil)
	if err != nil {
		return err
	}

	// Invalidate lock status cache
	c.cache.Delete(fmt.Sprintf("login_lockout:%s", identifier))
	return nil
}

// UpdatePassword updates user password and invalidates cache
func (c *CachedUserWriter) UpdatePassword(userID int, newPasswordHash string) error {
	err := c.writer.UpdatePassword(userID, newPasswordHash)
	if err != nil {
		return err
	}

	// Invalidate user caches
	c.cache.Delete(fmt.Sprintf("user:info:%d", userID))
	c.cache.Delete(fmt.Sprintf("user:basic:%d", userID))
	return nil
}

// DeleteUser deletes a user and invalidates all related caches
func (c *CachedUserWriter) DeleteUser(userID int) error {
	// Get username before deletion to invalidate cache
	user, err := c.reader.FindByID(userID)
	if err == nil && user != nil {
		// Invalidate all related caches
		c.cache.Delete(fmt.Sprintf("user:exists:%d", userID))
		c.cache.Delete(fmt.Sprintf("user:id:username:%s", user.Username))
		c.cache.Delete(fmt.Sprintf("user:info:%d", userID))
		c.cache.Delete(fmt.Sprintf("user:basic:%d", userID))
	}

	// Delete user from database
	err = c.writer.DeleteUser(userID)
	if err != nil {
		return err
	}

	return nil
}

// ClearExpiredLock clears expired lock
func (c *CachedUserWriter) ClearExpiredLock(identifier string) error {
	err := c.writer.ClearExpiredLock(identifier)
	if err != nil {
		return err
	}

	// Clear Redis lock if exists
	c.cache.Delete(fmt.Sprintf("login_lockout:%s", identifier))
	c.cache.Delete(fmt.Sprintf("login_attempts:%s", identifier))

	log.WithFields(log.Fields{
		"identifier": identifier,
		"source":     "cache_cleared",
	}).Info("Cleared expired lock from cache")

	return nil
}
