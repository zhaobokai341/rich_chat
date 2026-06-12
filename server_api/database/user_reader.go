package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// PostgresUserReader implements UserReader using PostgreSQL
type PostgresUserReader struct {
	db *sqlx.DB
}

// NewPostgresUserReader creates a new PostgreSQL user reader
func NewPostgresUserReader(db *sqlx.DB) *PostgresUserReader {
	return &PostgresUserReader{
		db: db,
	}
}

// FindByID retrieves a user by their ID
func (r *PostgresUserReader) FindByID(id int) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID %d: %w", id, err)
	}
	return &user, nil
}

// FindByUsername retrieves a user by their username
func (r *PostgresUserReader) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username %s: %w", username, err)
	}
	return &user, nil
}

// ExistsByID checks if a user exists by ID
func (r *PostgresUserReader) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence by ID %d: %w", id, err)
	}
	return exists, nil
}

// ExistsByUsername checks if a user exists by username
func (r *PostgresUserReader) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence by username %s: %w", username, err)
	}
	return exists, nil
}

// GetUserProfile retrieves user profile information
func (r *PostgresUserReader) GetUserProfile(userID int) (*UserInfo, error) {
	var userInfo UserInfo
	err := r.db.QueryRow(
		"SELECT username, email, nickname, bio FROM users WHERE id = $1", userID,
	).Scan(&userInfo.Username, &userInfo.Email, &userInfo.Nickname, &userInfo.Bio)

	if err != nil {
		return nil, fmt.Errorf("failed to get user profile for ID %d: %w", userID, err)
	}

	return &userInfo, nil
}

// GetLockStatus retrieves the lock status for an account
func (r *PostgresUserReader) GetLockStatus(identifier string) (*time.Time, error) {
	var lockUntil *time.Time
	err := r.db.QueryRow(
		"SELECT lock_until FROM users WHERE username = $1 OR id::text = $1",
		identifier,
	).Scan(&lockUntil)

	if err != nil {
		return nil, fmt.Errorf("failed to get lock status for identifier %s: %w", identifier, err)
	}

	return lockUntil, nil
}

// GetUserBasicInfo retrieves basic user info by ID (for chat)
func (r *PostgresUserReader) GetUserBasicInfo(userID int) (*UserBasicInfo, error) {
	var userInfo UserBasicInfo
	err := r.db.QueryRow(
		"SELECT id, username, nickname, bio FROM users WHERE id = $1", userID,
	).Scan(&userInfo.ID, &userInfo.Username, &userInfo.Nickname, &userInfo.Bio)
	if err != nil {
		return nil, fmt.Errorf("failed to get user basic info for ID %d: %w", userID, err)
	}
	return &userInfo, nil
}

// GetPasswordHash retrieves user's password hash (optimized for authentication)
func (r *PostgresUserReader) GetPasswordHash(userID int) (string, error) {
	var passwordHash string
	err := r.db.QueryRow(
		"SELECT password_hash FROM users WHERE id = $1", userID,
	).Scan(&passwordHash)
	if err != nil {
		return "", fmt.Errorf("failed to get password hash for ID %d: %w", userID, err)
	}
	return passwordHash, nil
}
