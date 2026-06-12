package database

import (
	"time"
)

// cachedUserRepositoryAdapter combines CachedUserReader and CachedUserWriter
type cachedUserRepositoryAdapter struct {
	reader *CachedUserReader
	writer *CachedUserWriter
}

// NewCachedUserRepository creates a cached UserRepository from separate reader and writer
func NewCachedUserRepository(reader UserReader, writer UserWriter, cache CacheService) UserRepository {
	cachedReader := NewCachedUserReader(reader, cache)
	cachedWriter := NewCachedUserWriter(writer, cache, reader)

	return &cachedUserRepositoryAdapter{
		reader: cachedReader,
		writer: cachedWriter,
	}
}

// Read operations
func (a *cachedUserRepositoryAdapter) FindByID(id int) (*User, error) {
	return a.reader.FindByID(id)
}

func (a *cachedUserRepositoryAdapter) FindByUsername(username string) (*User, error) {
	return a.reader.FindByUsername(username)
}

func (a *cachedUserRepositoryAdapter) ExistsByID(id int) (bool, error) {
	return a.reader.ExistsByID(id)
}

func (a *cachedUserRepositoryAdapter) ExistsByUsername(username string) (bool, error) {
	return a.reader.ExistsByUsername(username)
}

func (a *cachedUserRepositoryAdapter) GetUserProfile(userID int) (*UserInfo, error) {
	return a.reader.GetUserProfile(userID)
}

func (a *cachedUserRepositoryAdapter) GetUserBasicInfo(userID int) (*UserBasicInfo, error) {
	return a.reader.GetUserBasicInfo(userID)
}

func (a *cachedUserRepositoryAdapter) GetLockStatus(identifier string) (*time.Time, error) {
	return a.reader.GetLockStatus(identifier)
}

func (a *cachedUserRepositoryAdapter) GetPasswordHash(userID int) (string, error) {
	return a.reader.GetPasswordHash(userID)
}

// Write operations
func (a *cachedUserRepositoryAdapter) CreateUser(username, passwordHash string) (int, error) {
	return a.writer.CreateUser(username, passwordHash)
}

func (a *cachedUserRepositoryAdapter) UpdateProfile(userID int, key, value string) error {
	return a.writer.UpdateProfile(userID, key, value)
}

func (a *cachedUserRepositoryAdapter) UpdateLastLogin(userID int) error {
	return a.writer.UpdateLastLogin(userID)
}

func (a *cachedUserRepositoryAdapter) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	return a.writer.UpdateLockStatus(identifier, lockUntil)
}

func (a *cachedUserRepositoryAdapter) UpdatePassword(userID int, newPasswordHash string) error {
	return a.writer.UpdatePassword(userID, newPasswordHash)
}

func (a *cachedUserRepositoryAdapter) DeleteUser(userID int) error {
	return a.writer.DeleteUser(userID)
}

func (a *cachedUserRepositoryAdapter) ClearExpiredLock(identifier string) error {
	return a.writer.ClearExpiredLock(identifier)
}
