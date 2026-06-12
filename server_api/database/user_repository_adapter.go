package database

import (
	"time"
)

// userRepositoryAdapter combines UserReader and UserWriter to implement UserRepository
// This provides backward compatibility while using read-write separation internally
type userRepositoryAdapter struct {
	reader UserReader
	writer UserWriter
}

// NewUserRepositoryAdapter creates a UserRepository from separate reader and writer
func NewUserRepositoryAdapter(reader UserReader, writer UserWriter) UserRepository {
	return &userRepositoryAdapter{
		reader: reader,
		writer: writer,
	}
}

// Read operations (delegated to UserReader)
func (a *userRepositoryAdapter) FindByID(id int) (*User, error) {
	return a.reader.FindByID(id)
}

func (a *userRepositoryAdapter) FindByUsername(username string) (*User, error) {
	return a.reader.FindByUsername(username)
}

func (a *userRepositoryAdapter) ExistsByID(id int) (bool, error) {
	return a.reader.ExistsByID(id)
}

func (a *userRepositoryAdapter) ExistsByUsername(username string) (bool, error) {
	return a.reader.ExistsByUsername(username)
}

func (a *userRepositoryAdapter) GetUserProfile(userID int) (*UserInfo, error) {
	return a.reader.GetUserProfile(userID)
}

func (a *userRepositoryAdapter) GetUserBasicInfo(userID int) (*UserBasicInfo, error) {
	return a.reader.GetUserBasicInfo(userID)
}

func (a *userRepositoryAdapter) GetLockStatus(identifier string) (*time.Time, error) {
	return a.reader.GetLockStatus(identifier)
}

func (a *userRepositoryAdapter) GetPasswordHash(userID int) (string, error) {
	return a.reader.GetPasswordHash(userID)
}

// Write operations (delegated to UserWriter)
func (a *userRepositoryAdapter) CreateUser(username, passwordHash string) (int, error) {
	return a.writer.CreateUser(username, passwordHash)
}

func (a *userRepositoryAdapter) UpdateProfile(userID int, key, value string) error {
	return a.writer.UpdateProfile(userID, key, value)
}

func (a *userRepositoryAdapter) UpdateLastLogin(userID int) error {
	return a.writer.UpdateLastLogin(userID)
}

func (a *userRepositoryAdapter) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	return a.writer.UpdateLockStatus(identifier, lockUntil)
}

func (a *userRepositoryAdapter) UpdatePassword(userID int, newPasswordHash string) error {
	return a.writer.UpdatePassword(userID, newPasswordHash)
}

func (a *userRepositoryAdapter) DeleteUser(userID int) error {
	return a.writer.DeleteUser(userID)
}

func (a *userRepositoryAdapter) ClearExpiredLock(identifier string) error {
	return a.writer.ClearExpiredLock(identifier)
}
