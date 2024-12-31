package schema

import (
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type DbError struct {
	action string
	err    error
}

func NewDbError(action string, err error) error {
	slog.Error("sql error", "action", action, "error", err)
	return DbError{action: action, err: err}
}

func (e DbError) Error() string {
	return fmt.Sprintf("sql error while %v: %v", e.action, e.err)
}

func (e DbError) Unwrap() error {
	return e.err
}

func GetUser(userId string, db *gorm.DB) (User, error) {
	var user User

	result := db.First(&user, "id = ?", userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, fmt.Errorf("no user with id %v", userId)
		}
		return user, NewDbError("retrieving user by id", result.Error)
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
		return model, NewDbError("retrieving model by id", result.Error)
	}

	return model, nil
}

func GetUserTeamIds(userId string, db *gorm.DB) ([]string, error) {
	var teams []UserTeam
	result := db.Find(&teams, "user_id = ?", userId)
	if result.Error != nil {
		return nil, NewDbError("retrieving user_team entries", result.Error)
	}
	ids := make([]string, 0, len(teams))
	for _, team := range teams {
		ids = append(ids, team.TeamId)
	}
	return ids, nil
}

func GetUserTeam(teamId, userId string, db *gorm.DB) (*UserTeam, error) {
	var team UserTeam
	result := db.First(&team, "team_id = ? and user_id = ?", teamId, userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, NewDbError("retrieving user_team entry", result.Error)
	}

	return &team, nil
}

func ModelExists(db *gorm.DB, modelId string) (bool, error) {
	var model Model
	result := db.First(&model, "id = ?", modelId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, NewDbError("checking if model exists", result.Error)
	}
	return true, nil
}

func UserExists(db *gorm.DB, userId string) (bool, error) {
	var user User
	result := db.First(&user, "id = ?", userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, NewDbError("checking if user exists", result.Error)
	}
	return true, nil
}

func TeamExists(db *gorm.DB, teamId string) (bool, error) {
	var team Team
	result := db.First(&team, "id = ?", teamId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, NewDbError("checking if team exists", result.Error)
	}
	return true, nil
}
