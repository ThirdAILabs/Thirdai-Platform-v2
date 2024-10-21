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
	PublishDate  string  `json:"publish_date"`
	UserEmail    string  `json:"user_email"`
	Username     string  `json:"Username"`
	Access       string  `json:"access"`
	TrainStatus  string  `json:"train_status"`
	DeployStatus string  `json:"deploy_status"`
	TeamId       *string `json:"team_id"`

	Attributes map[string]string `json:"attributes"`

	Dependencies []modelDependency `json:"dependencies"`
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

	trainStatus, _, err := getModelStatus(model, m.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	deployStatus, _, err := getModelStatus(model, m.db, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	info := modelInfo{
		ModelId:      model.Id,
		ModelName:    model.Name,
		Type:         model.Type,
		Subtype:      model.Subtype,
		PublishDate:  model.PublishedDate.String(),
		UserEmail:    model.User.Username,
		Access:       model.Access,
		TrainStatus:  trainStatus,
		DeployStatus: deployStatus,
		TeamId:       model.TeamId,
		Attributes:   attributes,
		Dependencies: deps,
	}

	writeJsonResponse(w, info)
}
