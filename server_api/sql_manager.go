package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserDatabase struct {
	database *sqlx.DB
}

const (
	connection_str = "host=localhost port=5432 user=postgres password=123456 dbname=rich_chat sslmode=disable"
)

// initialize database connection
func database_init() *UserDatabase {
	var err error
	database, err := sqlx.Connect("postgres", connection_str)
	if err != nil {
		log.Fatal("Critital while establishing database connection: ", err.Error())
	}
	log.Info("Database connection established")
	return &UserDatabase{database}
}

// check if user exists in database
func (ud *UserDatabase) select_user_is_exist(username, password string) (int, bool) {
	var user_id int
	var password_hash string
	err := ud.database.QueryRow("SELECT id, password_hash FROM users WHERE username = $1", username).Scan(&user_id, &password_hash)
	if err != nil {
		log.Warning("Error while selecting user: ", err.Error())
		return 0, false
	}

	err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	if err != nil {
		log.Warning("Error while comparing password: ", err.Error())
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
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		username, string(password_hash),
	).Scan(&user_id)

	if err != nil {
		log.Error("Error while inserting user: ", err.Error())
		return 0, err
	}

	log.Info(fmt.Sprintf("User created successfully: %s (ID: %d)", username, user_id))
	return user_id, nil
}
