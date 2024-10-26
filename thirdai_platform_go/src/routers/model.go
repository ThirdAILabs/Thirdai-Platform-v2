package routers

import (
	"net/http"
	"thirdai_platform/src/auth"
	"thirdai_platform/src/schema"

	"gorm.io/gorm"
)

type ModelRouter struct {
	db          *gorm.DB
	userAuth    *auth.JwtManager
	sessionAuth *auth.JwtManager
}

type modelDependency struct {
	ModelId   string `json:"model_id"`
	ModelName string `json:"model_name"`
	Type      string `json:"type"`
	Subtype   string `json:"sub_type"`
	Username  string `json:"username"`
}

type modelInfo struct {
	ModelId      string  `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Type         string  `json:"type"`
	Subtype      string  `json:"sub_type"`
	Access       string  `json:"access"`
	TrainStatus  string  `json:"train_status"`
	DeployStatus string  `json:"deploy_status"`
	PublishDate  string  `json:"publish_date"`
	UserEmail    string  `json:"user_email"`
	Username     string  `json:"Username"`
	TeamId       *string `json:"team_id"`

	Attributes map[string]string `json:"attributes"`

	Dependencies []modelDependency `json:"dependencies"`
}

func convertToModelInfo(model schema.Model, db *gorm.DB) (modelInfo, error) {
	trainStatus, _, err := getModelStatus(model, db, true)
	if err != nil {
		return modelInfo{}, err
	}
	deployStatus, _, err := getModelStatus(model, db, false)
	if err != nil {
		return modelInfo{}, err
	}

	attributes := make(map[string]string, len(model.Attributes))
	for _, attr := range model.Attributes {
		attributes[attr.Key] = attr.Value
	}

	deps := make([]modelDependency, 0, len(model.Dependencies))
	for _, dep := range model.Dependencies {
		deps = append(deps, modelDependency{
			ModelId:   dep.ModelId,
			ModelName: dep.Model.Name,
			Type:      dep.Model.Type,
			Subtype:   dep.Model.Subtype,
			Username:  dep.Model.User.Username,
		})
	}

	return modelInfo{
		ModelId:      model.Id,
		ModelName:    model.Name,
		Type:         model.Type,
		Subtype:      model.Subtype,
		Access:       model.Access,
		TrainStatus:  trainStatus,
		DeployStatus: deployStatus,
		PublishDate:  model.PublishedDate.String(),
		UserEmail:    model.User.Email,
		Username:     model.User.Username,
		TeamId:       model.TeamId,
		Attributes:   attributes,
		Dependencies: deps,
	}, nil
}

func (m *ModelRouter) Details(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

	model, err := schema.GetModel(modelId, m.db, true, true, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, err := convertToModelInfo(model, m.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, info)
}

func (m *ModelRouter) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, m.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var models []schema.Model
	var result *gorm.DB
	if user.IsAdmin {
		result = m.db.Preload("Dependencies").Preload("Attributes").Preload("User").Find(&models)
	} else {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = m.db.
			Preload("Dependencies").Preload("Attributes").Preload("User").
			Where("access = ?", schema.Public).
			Or("access = ? AND user_id = ?", schema.Private, user.Id).
			Or("access = ? AND team_id IN ?", schema.Protected, userTeams).
			Find(&models)
	}

	if result.Error != nil {
		dbError(w, result.Error)
		return
	}

	infos := make([]modelInfo, 0, len(models))
	for _, model := range models {
		info, err := convertToModelInfo(model, m.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		infos = append(infos, info)
	}

	writeJsonResponse(w, infos)
}
