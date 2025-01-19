package schema

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// type DbError struct {
// 	action string
// 	err    error
// }

// func NewDbError(action string, err error) error {
// 	slog.Error("sql error", "action", action, "error", err)
// 	return DbError{action: action, err: err}
// }

// func (e DbError) Error() string {
// 	return fmt.Sprintf("sql error while %v: %v", e.action, e.err)
// }

// func (e DbError) Unwrap() error {
// 	return e.err
// }

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrModelNotFound    = errors.New("model not found")
	ErrTeamNotFound     = errors.New("team not found")
	ErrUserTeamNotFound = errors.New("user team relationship not found")
	ErrDbAccessFailed   = errors.New("db access failed")
)

func GetUser(userId uuid.UUID, db *gorm.DB) (User, error) {
	var user User

	result := db.First(&user, "id = ?", userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return user, ErrUserNotFound
		}
		slog.Error("sql error in get user", "user_id", userId, "error", result.Error)
		return user, ErrDbAccessFailed
	}

	return user, nil
}

func GetModel(modelId uuid.UUID, db *gorm.DB, loadDeps, loadAttrs, loadUser bool) (Model, error) {
	var model Model

	var result *gorm.DB = db
	if loadDeps {
		result = result.Preload("Dependencies").Preload("Dependencies.Dependency").Preload("Dependencies.Dependency.User")
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
			return model, ErrModelNotFound
		}
		slog.Error("sql error in get model", "model_id", modelId, "error", result.Error)
		return model, ErrDbAccessFailed
	}

	return model, nil
}

func GetTeam(teamId uuid.UUID, db *gorm.DB) (Team, error) {
	var team Team

	result := db.First(&team, "id = ?", teamId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return team, ErrTeamNotFound
		}
		slog.Error("sql error in get team", "team_id", teamId, "error", result.Error)
		return team, ErrDbAccessFailed
	}

	return team, nil
}

func GetUserTeamIds(userId uuid.UUID, db *gorm.DB) ([]uuid.UUID, error) {
	var teams []UserTeam
	result := db.Find(&teams, "user_id = ?", userId)
	if result.Error != nil {
		slog.Error("sql error in get user team ids", "user_id", userId, "error", result.Error)
		return nil, ErrDbAccessFailed
	}
	ids := make([]uuid.UUID, 0, len(teams))
	for _, team := range teams {
		ids = append(ids, team.TeamId)
	}
	return ids, nil
}

func GetUserTeam(teamId, userId uuid.UUID, db *gorm.DB) (UserTeam, error) {
	var team UserTeam
	result := db.First(&team, "team_id = ? and user_id = ?", teamId, userId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return team, ErrUserTeamNotFound
		}
		slog.Error("sql error in get user team", "team_id", teamId, "user_id", userId, "error", result.Error)
		return team, ErrDbAccessFailed
	}

	return team, nil
}
