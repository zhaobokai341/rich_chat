package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserDatabase struct {
	database *sqlx.DB
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
	return &UserDatabase{database}
}

// check if user exists in database with given user id
func (ud *UserDatabase) check_user_is_exist(user_id int) bool {
	var temp int
	err := ud.database.Get(&temp, "SELECT id FROM users WHERE id = $1", user_id)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warning("Error while selecting user")
		return false
	}
	return true
}

// check if user exists in database with given user id and password
func (ud *UserDatabase) select_user_id_is_exist(user_id int, password string) bool {
	var password_hash string
	err := ud.database.QueryRow(
		"SELECT id, password_hash FROM users WHERE id = $1", user_id,
	).Scan(&user_id, &password_hash)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warning("Error while selecting user")
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return false
	}
	return true
}

// check if user exists in database with given username and password
func (ud *UserDatabase) select_user_is_exist(username, password string) (int, bool) {
	var user_id int
	var password_hash string
	err := ud.database.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = $1", username,
	).Scan(&user_id, &password_hash)
	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"error":    err.Error(),
		}).Warning("Error while selecting user")
		return 0, false
	}

	err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
		return 0, false
	}
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
		}).Error("Error while inserting user: ")
		return 0, err
	}

	log.WithFields(log.Fields{
		"user_id":  user_id,
		"username": username,
	}).Info("User created successfully")
	return user_id, nil
}

func (ud *UserDatabase) delete_user(user_id int) error {
	_, err := ud.database.Exec("DELETE FROM users WHERE id = $1", user_id)
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
