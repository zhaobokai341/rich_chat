package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserDatabase struct {
	database *sqlx.DB
	redis    *redis.Client
	ctx      context.Context
}

// initialize database connection
func database_init() *UserDatabase {
	var err error
	database, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, DB_SSL,
		),
	)
	if err != nil {
		log.Fatal("Critital while establishing database connection: ", err.Error())
	}
	log.Info("Database connection established")

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", REDIS_HOST, REDIS_PORT),
		Password: REDIS_PASSWORD,
		DB:       REDIS_DB,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warning("Redis connection failed, continuing without cache")
		// Continue without Redis - cache will be disabled
		rdb = nil
	} else {
		log.Info("Redis connection established")
	}

	return &UserDatabase{
		database: database,
		redis:    rdb,
		ctx:      ctx,
	}
}

// check if user exists in database with given user id
func (ud *UserDatabase) check_user_is_exist(user_id int) bool {
	// Check cache first
	cacheKey := fmt.Sprintf("user:exists:%d", user_id)
	if cached, found := ud.getCache(cacheKey); found {
		exists, _ := strconv.ParseBool(cached)
		log.WithFields(log.Fields{
			"user_id": user_id,
			"source":  "cache",
		}).Debug("User existence check")
		return exists
	}

	// Cache miss - query database
	var temp int
	err := ud.database.Get(&temp, "SELECT id FROM users WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warning("Error while selecting user")
		// Cache the negative result for a shorter time
		ud.setCache(cacheKey, "false")
		return false
	}

	// Cache the result
	ud.setCache(cacheKey, "true")
	return true
}

// check if user exists in database with given user id and password
func (ud *UserDatabase) select_user_id_is_exist(user_id int, password string) bool {
	// For password verification, we don't cache as passwords should be verified fresh
	// But we can cache the password_hash to avoid repeated DB reads
	cacheKey := fmt.Sprintf("user:hash:%d", user_id)
	var password_hash string

	if cached, found := ud.getCache(cacheKey); found {
		password_hash = cached
		log.WithFields(log.Fields{
			"user_id": user_id,
			"source":  "cache",
		}).Debug("Password hash retrieved from cache")
	} else {
		// Cache miss - query database
		err := ud.database.QueryRow(
			"SELECT id, password_hash FROM users WHERE id = $1", user_id,
		).Scan(&user_id, &password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"error":   err.Error(),
			}).Warning("Error while selecting user")
			// Cache negative result with shorter TTL (60 seconds) to prevent DB hammering
			ud.setNullCache(cacheKey)
			return false
		}
		// Cache the password hash
		ud.setCache(cacheKey, password_hash)
	}

	// Check if cached value indicates user not found
	if password_hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return false
	}
	return true
}

// check if user exists in database with given username and password
func (ud *UserDatabase) select_user_is_exist(username, password string) (int, bool) {
	// Check cache for user_id by username
	cacheKey := fmt.Sprintf("user:id:username:%s", username)
	var user_id int

	if cached, found := ud.getCache(cacheKey); found {
		// Check if cached value indicates user not found
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return 0, false
		}

		user_id, _ = strconv.Atoi(cached)
		log.WithFields(log.Fields{
			"username": username,
			"user_id":  user_id,
			"source":   "cache",
		}).Debug("User ID retrieved from cache")
	} else {
		// Cache miss - query database
		var password_hash string
		err := ud.database.QueryRow(
			"SELECT id, password_hash FROM users WHERE username = $1", username,
		).Scan(&user_id, &password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"error":    err.Error(),
			}).Warning("Error while selecting user")
			// Cache negative result with shorter TTL (60 seconds) to prevent DB hammering
			ud.setNullCache(cacheKey)
			return 0, false
		}
		// Cache the user_id
		ud.setCache(cacheKey, strconv.Itoa(user_id))
		// Also cache the password hash
		hashCacheKey := fmt.Sprintf("user:hash:%d", user_id)
		ud.setCache(hashCacheKey, password_hash)
	}

	// Get password hash (from cache or DB)
	hashCacheKey := fmt.Sprintf("user:hash:%d", user_id)
	var password_hash string
	if cached, found := ud.getCache(hashCacheKey); found {
		// Skip if it's a negative cache marker
		if cached == "" {
			return 0, false
		}
		password_hash = cached
	} else {
		err := ud.database.QueryRow(
			"SELECT password_hash FROM users WHERE id = $1", user_id,
		).Scan(&password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"error":   err.Error(),
			}).Warning("Error while selecting password hash")
			return 0, false
		}
		ud.setCache(hashCacheKey, password_hash)
	}

	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return 0, false
	}

	// Update last_login in database (don't wait for this)
	_, err = ud.database.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while updating user last login")
		return 0, false
	}

	return user_id, true
}

func (ud *UserDatabase) check_username_is_exist(username string) bool {
	// Check cache first
	cacheKey := fmt.Sprintf("user:id:username:%s", username)

	if cached, found := ud.getCache(cacheKey); found {
		// Check if cached value indicates user not found
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return false
		} else {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User found (cached)")
			return true
		}
	} else {
		// Cache miss - query database
		var user_id int
		err := ud.database.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&user_id)
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"error":    err.Error(),
			}).Warning("Error while selecting user")
			ud.setNullCache(cacheKey)
			return false
		}
		ud.setCache(cacheKey, strconv.Itoa(user_id))
		log.WithFields(log.Fields{
			"username": username,
			"user_id":  user_id,
			"source":   "database",
		}).Debug("User found")
		return true
	}
}

// create a new user in database
func (ud *UserDatabase) insert_user(username, password string) (int, error) {
	// Hash the password
	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error while hashing password: ", err.Error())
		return 0, err
	}

	var user_id int
	err = ud.database.QueryRow(
		`INSERT INTO users (
			username,
			nickname,
			password_hash,
			registered_time,
			last_login
		)
		VALUES ($1, $1, $2, NOW(), NOW()) 
		RETURNING id`,
		username, string(password_hash),
	).Scan(&user_id)

	if err != nil {
		log.WithFields(log.Fields{
			"user_id":         user_id,
			"username":        username,
			"registered_time": time.Now(),
			"last_login":      time.Now(),
			"error":           err.Error(),
		}).Warning("Error while inserting user: ")
		return 0, err
	}

	// Invalidate related caches
	ud.deleteCache(fmt.Sprintf("user:exists:%d", user_id))
	ud.deleteCache(fmt.Sprintf("user:id:username:%s", username))
	ud.setCache(fmt.Sprintf("user:hash:%d", user_id), string(password_hash))

	log.WithFields(log.Fields{
		"user_id":  user_id,
		"username": username,
	}).Info("User created successfully")
	return user_id, nil
}

// Delete a user by user_id
func (ud *UserDatabase) delete_user(user_id int) error {
	// Get username before deletion to invalidate cache
	var username string
	err := ud.database.QueryRow("SELECT username FROM users WHERE id = $1", user_id).Scan(&username)
	if err == nil {
		// Invalidate all related caches
		ud.deleteCache(fmt.Sprintf("user:exists:%d", user_id))
		ud.deleteCache(fmt.Sprintf("user:id:username:%s", username))
		ud.deleteCache(fmt.Sprintf("user:hash:%d", user_id))
	}

	_, err = ud.database.Exec("DELETE FROM users WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while deleting user")
		return err
	}

	log.WithFields(log.Fields{
		"user_id": user_id,
	}).Info("Deleted user successfully")
	return nil
}
