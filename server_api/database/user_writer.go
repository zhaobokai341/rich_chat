package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// PostgresUserWriter implements UserWriter using PostgreSQL
type PostgresUserWriter struct {
	db *sqlx.DB
}

// NewPostgresUserWriter creates a new PostgreSQL user writer
func NewPostgresUserWriter(db *sqlx.DB) *PostgresUserWriter {
	return &PostgresUserWriter{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (w *PostgresUserWriter) CreateUser(username, passwordHash string) (int, error) {
	var userID int
	err := w.db.QueryRow(
		`INSERT INTO users (
			username,
			password_hash,
			last_login
		) VALUES ($1, $2, NOW()) 
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

// UpdateProfile updates a specific field in user profile
func (w *PostgresUserWriter) UpdateProfile(userID int, key, value string) error {
	allowedColumns := map[string]bool{
		"nickname": true,
		"bio":      true,
		"email":    true,
	}
	if !allowedColumns[key] {
		return fmt.Errorf("invalid column name: %s", key)
	}

	query := fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2", key)
	_, err := w.db.Exec(query, value, userID)
	if err != nil {
		return fmt.Errorf("failed to update user profile for ID %d: %w", userID, err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (w *PostgresUserWriter) UpdateLastLogin(userID int) error {
	_, err := w.db.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to update last login for user ID %d: %w", userID, err)
	}
	return nil
}

// UpdateLockStatus updates the account lock status
func (w *PostgresUserWriter) UpdateLockStatus(identifier string, lockUntil *time.Time) error {
	if lockUntil == nil {
		_, err := w.db.Exec(
			"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
			identifier,
		)
		if err != nil {
			return fmt.Errorf("failed to unlock account for identifier %s: %w", identifier, err)
		}
	} else {
		_, err := w.db.Exec(
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
func (w *PostgresUserWriter) UpdatePassword(userID int, newPasswordHash string) error {
	_, err := w.db.Exec(
		"UPDATE users SET password_hash = $1 WHERE id = $2",
		newPasswordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password for user ID %d: %w", userID, err)
	}
	return nil
}

// DeleteUser removes a user from the database
func (w *PostgresUserWriter) DeleteUser(userID int) error {
	_, err := w.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user with ID %d: %w", userID, err)
	}

	log.WithFields(log.Fields{
		"user_id": userID,
	}).Info("Deleted user successfully")

	return nil
}

// ClearExpiredLock removes expired lock from database
func (w *PostgresUserWriter) ClearExpiredLock(identifier string) error {
	_, err := w.db.Exec(
		"UPDATE users SET lock_until = NULL WHERE username = $1 OR id::text = $1",
		identifier,
	)
	if err != nil {
		return fmt.Errorf("failed to clear expired lock for identifier %s: %w", identifier, err)
	}
	return nil
}
