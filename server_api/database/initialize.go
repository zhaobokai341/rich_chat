package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	DB_HOST                   string
	DB_PORT                   int
	DB_USER                   string
	DB_PASS                   string
	DB_NAME                   string
	DB_SSL                    string
	MAXOPENCONNS              int
	MAXIDLECONNS              int
	CONNMAXLIFETIME           time.Duration
	CONNMAXIDLETIME           time.Duration
	MAX_LOGIN_ATTEMPTS        int64
	LOCKOUT_DURATION          time.Duration
	IP_LIMIT_TIME             time.Duration
	IP_LIMIT_VISIT_TIMES      int
	IP_LIMIT_LOCKOUT_DURATION time.Duration
}

type RedisManager struct {
	GetCache         func(key string) (string, bool)
	GetIntValue      func(key string) (int64, bool)
	SetCache         func(key string, value string)
	SetNullCache     func(key string)
	SetCacheWithTTL  func(key string, value string, ttl int)
	SetLockoutCache  func(key string)
	IncrementCounter func(key string) int64
	SetKeyExpiration func(key string, ttl time.Duration)
	DeleteCache      func(key string)
}

// initialize database connection
func DatabaseInit(cfg Config, Redis_manager RedisManager) *UserDatabase {
	var err error
	db, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB_HOST, cfg.DB_PORT, cfg.DB_USER, cfg.DB_PASS, cfg.DB_NAME, cfg.DB_SSL,
		),
	)
	if err != nil {
		log.Fatal("Critital while establishing database connection: ", err.Error())
	}
	log.Info("Database connection established")

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MAXOPENCONNS)
	db.SetMaxIdleConns(cfg.MAXIDLECONNS)
	db.SetConnMaxLifetime(cfg.CONNMAXLIFETIME)
	db.SetConnMaxIdleTime(cfg.CONNMAXIDLETIME)

	return &UserDatabase{
		Cfg:           cfg,
		Database:      db,
		Redis_manager: Redis_manager,
	}
}
