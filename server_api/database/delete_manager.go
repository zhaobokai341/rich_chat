package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Delete a user by user_id
func (ud *UserDatabase) DeleteUser(user_id int) error {
	// Get username before deletion to invalidate cache
	var username string
	err := ud.Database.QueryRow(
		`SELECT
			username
		FROM
			users
		WHERE
			id = $1", user_id`).Scan(&username)

	if err == nil {
		// Invalidate all related caches
		ud.Redis_manager.DeleteCache(fmt.Sprintf("user:exists:%d", user_id))
		ud.Redis_manager.DeleteCache(fmt.Sprintf("user:id:username:%s", username))
		ud.Redis_manager.DeleteCache(fmt.Sprintf("user:hash:%d", user_id))
	}

	// Delete user from database
	_, err = ud.Database.Exec("DELETE FROM users WHERE id = $1", user_id)
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
