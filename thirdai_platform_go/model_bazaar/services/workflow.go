package services

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowService struct {
	db       *gorm.DB
	userAuth *auth.JwtManager
}

func (s *WorkflowService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(s.userAuth.Verifier())
	r.Use(s.userAuth.Authenticator())

	r.Post("/enterprise-search", s.EnterpriseSearch)

	return r
}

type EnterpriseSearchRequest struct {
	ModelName       string  `json:"model_name"`
	RetrievalId     string  `json:"retrieval_id"`
	GuardrailId     *string `json:"guardrail_id"`
	LlmProvider     *string `json:"llm_provider"`
	NlpClassifierId *string `json:"nlp_classifier_id"`
	DefaultMode     *string `json:"default_mode"`
}

func (r *EnterpriseSearchRequest) models() map[string]string {
	models := map[string]string{"retrieval_id": r.RetrievalId}
	if r.GuardrailId != nil {
		models["guardrail_id"] = *r.GuardrailId
	}
	if r.NlpClassifierId != nil {
		models["nlp_classifier_id"] = *r.NlpClassifierId
	}
	return models
}

func (s *WorkflowService) EnterpriseSearch(w http.ResponseWriter, r *http.Request) {
	var params EnterpriseSearchRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		err := checkForDuplicateModel(txn, params.ModelName, userId)
		if err != nil {
			return err
		}

		models := params.models()
		deps := make([]schema.ModelDependency, 0, len(models))
		attrs := make([]schema.ModelAttribute, 0, len(models)+2)
		for key, depId := range models {
			// TODO: check dep types
			exists, err := schema.ModelExists(txn, depId)
			if err != nil {
				return fmt.Errorf("error checking if %v exists: %w", key, err)
			}
			if !exists {
				return fmt.Errorf("model specified for %v does not exist", key)
			}

			perm, err := auth.GetModelPermissions(depId, userId, txn)
			if err != nil {
				return fmt.Errorf("error verifying permissions for %v: %w", key, err)
			}
			if perm < auth.ReadPermission {
				return fmt.Errorf("user does not have permissiont to access %v", key)
			}

			deps = append(deps, schema.ModelDependency{ModelId: modelId, DependencyId: depId})
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: key, Value: depId})
		}

		if params.LlmProvider != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "llm_provider", Value: *params.LlmProvider})
		}
		if params.DefaultMode != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "default_mode", Value: *params.DefaultMode})
		}

		model := createModel(modelId, params.ModelName, schema.EnterpriseSearch, nil, userId)
		model.Dependencies = deps
		model.Attributes = attrs

		result := txn.Create(&model)
		if result.Error != nil {
			return schema.NewDbError("create enterprise search model", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating enterpise search model: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteJsonResponse(w, map[string]string{"model_id": modelId})
}
