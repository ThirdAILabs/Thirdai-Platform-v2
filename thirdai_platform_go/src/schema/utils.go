package schema

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func GetUser(userId string, db *gorm.DB, loadTeams bool) (User, error) {
	var user User

	var result *gorm.DB
	if loadTeams {
		result = db.Preload("Teams").Preload("Teams.Team").First(&user, "id = ?", userId)
	} else {
		result = db.First(&user, "id = ?", userId)
	}

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, fmt.Errorf("no user with id %v", userId)
		}
		return user, fmt.Errorf("error locating user with id %v: %v", userId, result.Error)
	}

	return user, nil
}
