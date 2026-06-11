package database

import (
	"time"
)

// User represents a user entity
type User struct {
	ID             int        `db:"id"`
	Username       string     `db:"username"`
	Email          string     `db:"email"`
	Nickname       string     `db:"nickname"`
	Bio            string     `db:"bio"`
	PasswordHash   string     `db:"password_hash"`
	RegisteredTime *time.Time `db:"registered_time"`
	LastLogin      *time.Time `db:"last_login"`
	LockUntil      *time.Time `db:"lock_until"`
}

// UserInfo represents user profile information (without sensitive data)
type UserInfo struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Email    string `db:"email" json:"email"`
	Nickname string `db:"nickname" json:"nickname"`
	Bio      string `db:"bio" json:"bio"`
}

// UserBasicInfo represents basic user info for chat (without email)
type UserBasicInfo struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Nickname string `db:"nickname" json:"nickname"`
	Bio      string `db:"bio" json:"bio"`
}
