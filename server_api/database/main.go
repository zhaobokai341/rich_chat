package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserDatabase struct {
	Database      *sqlx.DB
	Cfg           Config
	Redis_manager RedisManager
}
