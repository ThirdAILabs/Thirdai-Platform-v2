package auth

import (
	"fmt"
	"thirdai_platform/src/schema"

	"gorm.io/gorm"
)

func ExpectAdmin(userId string, db *gorm.DB) error {
	user, err := schema.GetUser(userId, db, false)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return fmt.Errorf("user %v is not an admin", userId)
	}
	return nil
}
