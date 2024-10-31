package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"

	"gorm.io/gorm"
)

func parseRequestBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dest)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
		return false
	}
	return true
}

func dbError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
}

func writeJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func listModelDependencies(modelId string, db *gorm.DB) ([]schema.Model, error) {
	queue := make(chan string)
	queue <- modelId
	visited := map[string]struct{}{}

	models := []schema.Model{}

	for len(queue) > 0 {
		id := <-queue
		if _, ok := visited[id]; ok {
			continue
		}

		visited[id] = struct{}{}

		model, err := schema.GetModel(id, db, true, false, false)
		if err != nil {
			return nil, err
		}

		models = append(models, model)

		for _, dep := range model.Dependencies {
			queue <- dep.DependencyId
		}
	}

	return models, nil
}

func getStatus(model *schema.Model, trainStatus bool) string {
	if trainStatus {
		return model.TrainStatus
	}
	return model.DeployStatus
}

func getModelStatus(model schema.Model, db *gorm.DB, trainStatus bool) (string, []string, error) {
	status := getStatus(&model, trainStatus)
	if status == schema.NotStarted || status == schema.Stopped || status == schema.Failed {
		return status, []string{fmt.Sprintf("workflow %v has status %v", model.Name, status)}, nil
	}

	statusPriority := []string{
		schema.Failed, schema.NotStarted, schema.Stopped,
		schema.Starting, schema.InProgress, schema.Complete,
	}

	statuses := map[string][]string{}
	for _, s := range statusPriority {
		statuses[s] = []string{}
	}

	deps, err := listModelDependencies(model.Id, db)
	if err != nil {
		return "", nil, err
	}

	for _, dep := range deps {
		status := getStatus(&dep, trainStatus)
		if dep.Id == model.Id {
			statuses[status] = append(statuses[status], fmt.Sprintf("workflow %v has status %v", dep.Id, status))
		} else {
			statuses[status] = append(statuses[status], fmt.Sprintf("the workflow depends on %v which has status %v", dep.Id, status))
		}
	}

	for _, statusType := range statusPriority {
		if len(statuses[statusType]) > 0 {
			return statusType, statuses[statusType], nil
		}
	}

	return "", nil, fmt.Errorf("error finding statuses")
}

func countDownstreamModels(modelId string, db *gorm.DB, activeOnly bool) (int64, error) {

	query := db.Model(&schema.ModelDependency{}).Where("dependency_id = ?", modelId)
	if activeOnly {
		query = query.Joins("Model").Where("deploy_status IN ?", []string{schema.Starting, schema.InProgress, schema.Complete})
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("database error: %v", result.Error)
	}

	return count, nil
}
