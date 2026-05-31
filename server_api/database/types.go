package database

import (
	"time"
)

// User represents a user entity
type User struct {
	ID             int        `db:"id"`
	Username       string     `db:"username"`
	Nickname       string     `db:"nickname"`
	Bio            string     `db:"bio"`
	PasswordHash   string     `db:"password_hash"`
	RegisteredTime *time.Time `db:"registered_time"`
	LastLogin      *time.Time `db:"last_login"`
	LockUntil      *time.Time `db:"lock_until"`
}

// UserInfo represents user profile information (without sensitive data)
type UserInfo struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}
