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
	userAuth auth.IdentityProvider
}

func (s *WorkflowService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(s.userAuth.AuthMiddleware()...)

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

type searchComponent struct {
	component    string
	id           string
	expectedType string
}

func (r *EnterpriseSearchRequest) components() []searchComponent {
	components := []searchComponent{{component: "retrieval_id", id: r.RetrievalId, expectedType: schema.NdbModel}}
	if r.GuardrailId != nil {
		components = append(components, searchComponent{component: "guardrail_id", id: *r.GuardrailId, expectedType: schema.NlpTokenModel})
	}
	if r.NlpClassifierId != nil {
		components = append(components, searchComponent{component: "nlp_classifier_id", id: *r.NlpClassifierId, expectedType: schema.NlpTextModel})
	}
	return components
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

		components := params.components()
		deps := make([]schema.ModelDependency, 0, len(components))
		attrs := make([]schema.ModelAttribute, 0, len(components)+2)
		for _, component := range components {
			// TODO: check dep types
			model, err := schema.GetModel(component.id, txn, false, false, false)
			if err != nil {
				return fmt.Errorf("error getting model for %v: %w", component.component, err)
			}
			if model.Type != component.expectedType {
				return fmt.Errorf("component %v was expected to have type %v, but specified model has type %v", component.component, component.expectedType, model.Type)
			}

			perm, err := auth.GetModelPermissions(model.Id, userId, txn)
			if err != nil {
				return fmt.Errorf("error verifying permissions for %v: %w", component.component, err)
			}
			if perm < auth.ReadPermission {
				return fmt.Errorf("user does not have permissiont to access %v", component.component)
			}

			deps = append(deps, schema.ModelDependency{ModelId: modelId, DependencyId: model.Id})
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: component.component, Value: model.Id})
		}

		if params.LlmProvider != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "llm_provider", Value: *params.LlmProvider})
		}
		if params.DefaultMode != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "default_mode", Value: *params.DefaultMode})
		}

		model := createModel(modelId, params.ModelName, schema.EnterpriseSearch, nil, userId)
		model.TrainStatus = schema.Complete
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
