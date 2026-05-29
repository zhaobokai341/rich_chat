package database

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// create a new user in database
func (ud *UserDatabase) InsertUser(username, password string) (int, error) {
	// Hash the password
	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error while hashing password: ", err.Error())
		return 0, err
	}

	// Insert the user into the database
	var user_id int
	err = ud.Database.QueryRow(
		`INSERT INTO users (
			username,
			nickname,
			password_hash,
			bio,
			last_login
		)
		VALUES ($1, $1, $2, '', NOW()) 
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
	ud.Redis_manager.DeleteCache(fmt.Sprintf("user:exists:%d", user_id))
	ud.Redis_manager.DeleteCache(fmt.Sprintf("user:id:username:%s", username))
	ud.Redis_manager.SetCache(fmt.Sprintf("user:hash:%d", user_id), string(password_hash))

	log.WithFields(log.Fields{
		"user_id":  user_id,
		"username": username,
	}).Info("User created successfully")
	return user_id, nil
}
