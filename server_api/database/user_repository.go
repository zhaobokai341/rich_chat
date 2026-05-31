package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (r *PostgresUserRepository) CreateUser(username, passwordHash string) (int, error) {
	var userID int
	err := r.db.QueryRow(
		`INSERT INTO users (
			username,
			nickname,
			password_hash,
			bio,
			last_login
		) VALUES ($1, $1, $2, '', NOW()) 
		RETURNING id`,
		username, passwordHash,
	).Scan(&userID)

	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"error":    err.Error(),
		}).Warning("Error while inserting user")
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id":  userID,
		"username": username,
	}).Info("User created successfully")

	return userID, nil
}

// FindByID retrieves a user by their ID
func (r *PostgresUserRepository) FindByID(id int) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID %d: %w", id, err)
	}
	return &user, nil
}

// FindByUsername retrieves a user by their username
func (r *PostgresUserRepository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username %s: %w", username, err)
	}
	return &user, nil
}

// ExistsByID checks if a user exists by ID
func (r *PostgresUserRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence by ID %d: %w", id, err)
	}
	return exists, nil
}

// ExistsByUsername checks if a user exists by username
func (r *PostgresUserRepository) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence by username %s: %w", username, err)
	}
	return exists, nil
}

// GetUserProfile retrieves user profile information
func (r *PostgresUserRepository) GetUserProfile(userID int) (*UserInfo, error) {
	var userInfo UserInfo
	err := r.db.QueryRow(
		"SELECT username, nickname, bio FROM users WHERE id = $1", userID,
	).Scan(&userInfo.Username, &userInfo.Nickname, &userInfo.Bio)

	if err != nil {
		return nil, fmt.Errorf("failed to get user profile for ID %d: %w", userID, err)
	}

	return &userInfo, nil
}

// UpdateProfile updates a specific field in user profile
func (r *PostgresUserRepository) UpdateProfile(userID int, key, value string) error {
	// Validate the column name to prevent SQL injection
	allowedColumns := map[string]bool{
		"nickname": true,
		"bio":      true,
	}
	if !allowedColumns[key] {
		return fmt.Errorf("invalid column name: %s", key)
	}

	_, err := r.db.Exec(
		"UPDATE users SET "+key+" = $1 WHERE id = $2",
		value, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user profile for ID %d: %w", userID, err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(userID int) error {
	_, err := r.db.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to update last login for user ID %d: %w", userID, err)
	}
	return nil
}

// UpdateLockStatus updates the account lock status
func (r *PostgresUserRepository) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	if lockUntil == nil {
		// Unlock account
		_, err := r.db.Exec(
			"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
			identifier,
		)
		if err != nil {
			return fmt.Errorf("failed to unlock account for identifier %s: %w", identifier, err)
		}
	} else {
		// Lock account
		_, err := r.db.Exec(
			"UPDATE users SET lock_until = $1 WHERE username = $2 OR id::text = $2",
			lockUntil, identifier,
		)
		if err != nil {
			return fmt.Errorf("failed to lock account for identifier %s: %w", identifier, err)
		}
	}
	return nil
}

// UpdatePassword updates the user's password
func (r *PostgresUserRepository) UpdatePassword(userID int, newPasswordHash string) error {
	_, err := r.db.Exec(
		"UPDATE users SET password_hash = $1 WHERE id = $2",
		newPasswordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password for user ID %d: %w", userID, err)
	}
	return nil
}

// DeleteUser removes a user from the database
func (r *PostgresUserRepository) DeleteUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user with ID %d: %w", userID, err)
	}

	log.WithFields(log.Fields{
		"user_id": userID,
	}).Info("Deleted user successfully")

	return nil
}

// GetLockStatus retrieves the lock status for an account
func (r *PostgresUserRepository) GetLockStatus(identifier string) (*time.Time, error) {
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

// ClearExpiredLock removes expired lock from database
func (r *PostgresUserRepository) ClearExpiredLock(identifier string) error {
	_, err := r.db.Exec(
		"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
		identifier,
	)
	if err != nil {
		return fmt.Errorf("failed to clear expired lock for identifier %s: %w", identifier, err)
	}
	return nil
}
