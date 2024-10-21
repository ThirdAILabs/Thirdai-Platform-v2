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

func GetModel(modelId string, db *gorm.DB, loadDeps, loadAttrs, loadUser bool) (Model, error) {
	var model Model

	var result *gorm.DB = db
	if loadDeps {
		result = result.Preload("Dependencies")
	}
	if loadAttrs {
		result = result.Preload("Attributes")
	}
	if loadUser {
		result = result.Preload("User")
	}
	result = result.First(&model, "id = ?", modelId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model, fmt.Errorf("no model with id %v", modelId)
		}
		return model, fmt.Errorf("error locating model with id %v: %v", modelId, result.Error)
	}

	return model, nil
}

func GetUserTeam(teamId, userId string, db *gorm.DB) (*UserTeam, error) {
	var team UserTeam
	result := db.First(&team, "team_id = ? and user_id = ?", teamId, userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %v", result.Error)
	}

	return &team, nil
}

func GetModelPermission(modelId, userId string, db *gorm.DB) (*ModelPermission, error) {
	var permission ModelPermission
	result := db.First(&permission, "model_id = ? and user_id = ?", modelId, userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %v", result.Error)
	}

	return &permission, nil
}
