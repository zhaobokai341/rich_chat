package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// check if user exists in database with given user id
func (ud *UserDatabase) CheckUserIsExist(user_id int) bool {
	// Check cache first
	cacheKey := fmt.Sprintf("user:exists:%d", user_id)
	if cached, found := ud.Redis_manager.GetCache(cacheKey); found {
		exists, _ := strconv.ParseBool(cached)
		log.WithFields(log.Fields{
			"user_id": user_id,
			"source":  "cache",
		}).Debug("User existence check")
		return exists
	}

	// Cache miss - query database
	var temp int
	err := ud.Database.Get(&temp, "SELECT id FROM users WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warning("Error while selecting user")
		// Cache the negative result for a shorter time
		ud.Redis_manager.SetCache(cacheKey, "false")
		return false
	}

	// Cache the result
	ud.Redis_manager.SetCache(cacheKey, "true")
	return true
}

// check if user exists in database with given user id and password
func (ud *UserDatabase) SelectUserIdIsExist(user_id int, password string) bool {
	// Cache the password_hash to avoid repeated DB reads
	cacheKey := fmt.Sprintf("user:hash:%d", user_id)
	var password_hash string

	if cached, found := ud.Redis_manager.GetCache(cacheKey); found {
		password_hash = cached
		log.WithFields(log.Fields{
			"user_id": user_id,
			"source":  "cache",
		}).Debug("Password hash retrieved from cache")
	} else {
		// Cache miss - query database
		err := ud.Database.QueryRow(
			"SELECT id, password_hash FROM users WHERE id = $1", user_id,
		).Scan(&user_id, &password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"error":   err.Error(),
			}).Warning("Error while selecting user")
			// Cache negative result with shorter TTL (60 seconds) to prevent DB hammering
			ud.Redis_manager.SetNullCache(cacheKey)
			return false
		}
		// Cache the password hash
		ud.Redis_manager.SetCache(cacheKey, password_hash)
	}

	// Check if cached value indicates user not found
	if password_hash == "" {
		return false
	}

	// Compare password with password hash
	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return false
	}
	return true
}

// check if user exists in database with given username and password
func (ud *UserDatabase) SelectUserIsExist(username, password string) (int, bool) {
	// Check cache for user_id by username
	cacheKey := fmt.Sprintf("user:id:username:%s", username)
	var user_id int

	if cached, found := ud.Redis_manager.GetCache(cacheKey); found {
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
		err := ud.Database.QueryRow(
			"SELECT id, password_hash FROM users WHERE username = $1", username,
		).Scan(&user_id, &password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"error":    err.Error(),
			}).Warning("Error while selecting user")
			// Cache negative result with shorter TTL (60 seconds) to prevent DB hammering
			ud.Redis_manager.SetNullCache(cacheKey)
			return 0, false
		}
		// Cache the user_id
		ud.Redis_manager.SetCache(cacheKey, strconv.Itoa(user_id))
		// Also cache the password hash
		hashCacheKey := fmt.Sprintf("user:hash:%d", user_id)
		ud.Redis_manager.SetCache(hashCacheKey, password_hash)
	}

	// Get password hash (from cache or DB)
	hashCacheKey := fmt.Sprintf("user:hash:%d", user_id)
	var password_hash string
	if cached, found := ud.Redis_manager.GetCache(hashCacheKey); found {
		// Skip if it's a negative cache marker
		if cached == "" {
			return 0, false
		}
		password_hash = cached
	} else {
		err := ud.Database.QueryRow(
			"SELECT password_hash FROM users WHERE id = $1", user_id,
		).Scan(&password_hash)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"error":   err.Error(),
			}).Warning("Error while selecting password hash")
			return 0, false
		}
		ud.Redis_manager.SetCache(hashCacheKey, password_hash)
	}

	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return 0, false
	}

	// Update last_login in database (don't wait for this)
	_, err = ud.Database.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while updating user last login")
		return 0, false
	}

	return user_id, true
}

// check username is exist in database
func (ud *UserDatabase) CheckUsernameIsExist(username string) bool {
	// Check cache first
	cacheKey := fmt.Sprintf("user:id:username:%s", username)

	if cached, found := ud.Redis_manager.GetCache(cacheKey); found {
		// Check if cached value indicates user not found
		if cached == "" {
			log.WithFields(log.Fields{
				"username": username,
				"source":   "cache",
			}).Debug("User not found (cached)")
			return false
		}

		log.WithFields(log.Fields{
			"username": username,
			"source":   "cache",
		}).Debug("User found (cached)")
		return true
	}

	// Cache miss - query database
	var user_id int
	err := ud.Database.QueryRow(
		`SELECT
			id
		FROM
			users
		WHERE
			username = $1`, username).Scan(&user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"error":    err.Error(),
		}).Warning("Error while selecting user")
		ud.Redis_manager.SetNullCache(cacheKey)
		return false
	}

	// Cache the user_id
	ud.Redis_manager.SetCache(cacheKey, strconv.Itoa(user_id))
	log.WithFields(log.Fields{
		"username": username,
		"user_id":  user_id,
		"source":   "database",
	}).Debug("User found")
	return true
}

// Get user info
func (ud *UserDatabase) GetUserProfile(user_id int) (UserInfo, error) {
	var user_info UserInfo
	// Check cache first
	cacheKey := fmt.Sprintf("user:info:%d", user_id)
	if cached, found := ud.Redis_manager.GetCache(cacheKey); found {
		if cached == "" {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"source":  "cache",
			}).Debug("User info not found (cached)")
			return user_info, nil
		}
		err := json.Unmarshal([]byte(cached), &user_info)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": user_id,
				"error":   err.Error(),
			}).Error("Error while unmarshalling user info from cache")
			return user_info, err
		}
		log.WithFields(log.Fields{
			"user_id": user_id,
			"source":  "cache",
		}).Debug("User info found (cached)")
		return user_info, nil
	}

	// Cache miss - query database
	err := ud.Database.QueryRow(
		"SELECT username, nickname, bio FROM users WHERE id = $1", user_id,
	).Scan(&user_info.Username, &user_info.Nickname, &user_info.Bio)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while getting user info")
		ud.Redis_manager.SetNullCache(cacheKey)
		return user_info, err
	}

	// Cache the user info
	user_str, err := json.Marshal(user_info)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while marshalling user info")
		return user_info, err
	}
	ud.Redis_manager.SetCache(cacheKey, string(user_str))
	return user_info, nil
}
