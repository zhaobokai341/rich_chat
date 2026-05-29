package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// DatabaseService is the new refactored service that uses repositories
type DatabaseService struct {
	userRepo      UserRepository
	rateLimitRepo RateLimitRepository
	tokenRepo     TokenRepository
	cache         CacheService
	config        Config
	db            *sqlx.DB
}

// NewDatabaseService creates a new database service with dependency injection
func NewDatabaseService(
	db *sqlx.DB,
	cache CacheService,
	config Config,
) *DatabaseService {
	// Create PostgreSQL user repository
	pgUserRepo := NewPostgresUserRepository(db)

	// Wrap with caching decorator
	cachedUserRepo := NewCachedUserRepository(pgUserRepo, cache)

	// Create rate limit repository
	rateLimitRepo := NewRedisRateLimitRepository(cache, cachedUserRepo, config)

	// Create token repository
	tokenRepo := NewRedisTokenRepository(cache, config.IP_LIMIT_TIME)

	return &DatabaseService{
		userRepo:      cachedUserRepo,
		rateLimitRepo: rateLimitRepo,
		tokenRepo:     tokenRepo,
		cache:         cache,
		config:        config,
		db:            db,
	}
}

// GetUserRepository returns the user repository
func (ds *DatabaseService) GetUserRepository() UserRepository {
	return ds.userRepo
}

// GetRateLimitRepository returns the rate limit repository
func (ds *DatabaseService) GetRateLimitRepository() RateLimitRepository {
	return ds.rateLimitRepo
}

// GetTokenRepository returns the token repository
func (ds *DatabaseService) GetTokenRepository() TokenRepository {
	return ds.tokenRepo
}

// GetDB returns the underlying database connection
func (ds *DatabaseService) GetDB() *sqlx.DB {
	return ds.db
}

// InitializeDatabaseService creates and initializes a new DatabaseService
func InitializeDatabaseService(cfg Config, redisManager RedisManager) (*DatabaseService, error) {
	var err error
	db, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB_HOST, cfg.DB_PORT, cfg.DB_USER, cfg.DB_PASS, cfg.DB_NAME, cfg.DB_SSL,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info("Database connection established")

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MAXOPENCONNS)
	db.SetMaxIdleConns(cfg.MAXIDLECONNS)
	db.SetConnMaxLifetime(cfg.CONNMAXLIFETIME)
	db.SetConnMaxIdleTime(cfg.CONNMAXIDLETIME)

	// Create cache adapter
	cache := NewRedisCacheAdapter(redisManager)

	// Create and return database service
	service := NewDatabaseService(db, cache, cfg)
	return service, nil
}
