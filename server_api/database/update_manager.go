package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Update user info
func (ud *UserDatabase) ChangeUserProfile(user_id int, user_info_key string, user_info_value string) error {
	// Validate the column name to prevent SQL injection
	allowedColumns := map[string]bool{
		"nickname": true,
		"bio":      true,
	}
	if !allowedColumns[user_info_key] {
		return fmt.Errorf("invalid column name: %s", user_info_key)
	}

	// Update the user info in the database
	_, err := ud.Database.Exec(
		"UPDATE users SET "+user_info_key+" = $1 WHERE id = $2",
		user_info_value, user_id,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("Error while updating user info")
		return err
	}

	// Delete the user info cache
	ud.Redis_manager.DeleteCache(fmt.Sprintf("user:info:%d", user_id))
	return nil
}
