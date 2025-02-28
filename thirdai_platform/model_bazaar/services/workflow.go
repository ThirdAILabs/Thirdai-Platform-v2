package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type WorkflowService struct {
	db       *gorm.DB
	storage  storage.Storage
	userAuth auth.IdentityProvider
}

func (s *WorkflowService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(s.userAuth.AuthMiddleware()...)

	r.Post("/enterprise-search", s.EnterpriseSearch)
	r.Post("/knowledge-extraction", s.KnowledgeExtraction)

	return r
}

type EnterpriseSearchRequest struct {
	ModelName       string     `json:"model_name"`
	RetrievalId     uuid.UUID  `json:"retrieval_id"`
	GuardrailId     *uuid.UUID `json:"guardrail_id"`
	LlmProvider     *string    `json:"llm_provider"`
	NlpClassifierId *uuid.UUID `json:"nlp_classifier_id"`
	DefaultMode     *string    `json:"default_mode"`
}

type searchComponent struct {
	component    string
	id           uuid.UUID
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

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	modelId := uuid.New()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		components := params.components()
		deps := make([]schema.ModelDependency, 0, len(components))
		// Each component is stored as an attribute, and then we have 2 additional hyperparameters
		attrs := make([]schema.ModelAttribute, 0, len(components)+2)
		for _, component := range components {
			model, err := schema.GetModel(component.id, txn, false, false, false)
			if err != nil {
				if errors.Is(err, schema.ErrModelNotFound) {
					return CodedError(fmt.Errorf("model %v for component %s does not exist", component.id, component.component), http.StatusNotFound)
				}
				return CodedError(fmt.Errorf("error loading model for component %s: %w", component.component, schema.ErrDbAccessFailed), http.StatusInternalServerError)
			}
			if model.Type != component.expectedType {
				return CodedError(fmt.Errorf("component %v was expected to have type %v, but specified model has type %v", component.component, component.expectedType, model.Type), http.StatusUnprocessableEntity)
			}

			perm, err := auth.GetModelPermissions(model.Id, user, txn)
			if err != nil {
				slog.Error("error verify permissions for component of enterprise search workflow", "model_id", component.id, "error", err)
				return CodedError(fmt.Errorf("error verifying permissions for component %v", component.component), http.StatusInternalServerError)
			}
			if perm < auth.ReadPermission {
				return CodedError(fmt.Errorf("user does not have permissions to access %v", component.component), http.StatusForbidden)
			}

			deps = append(deps, schema.ModelDependency{ModelId: modelId, DependencyId: model.Id})
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: component.component, Value: component.id.String()})
		}

		if params.LlmProvider != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "llm_provider", Value: *params.LlmProvider})
		}
		if params.DefaultMode != nil {
			attrs = append(attrs, schema.ModelAttribute{ModelId: modelId, Key: "default_mode", Value: *params.DefaultMode})
		}

		model := newModel(modelId, params.ModelName, schema.EnterpriseSearch, nil, user.Id)
		model.TrainStatus = schema.Complete
		model.Dependencies = deps
		model.Attributes = attrs

		return saveModel(txn, model, user)
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating enterprise search model: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteJsonResponse(w, trainResponse{ModelId: modelId})
}

type QuestionKeywords struct {
	Question string   `json:"question"`
	Keywords []string `json:"keywords"`
}

type KnowledgeExtractionRequest struct {
	ModelName string             `json:"model_name"`
	Questions []QuestionKeywords `json:"questions"`

	LlmProvider      string      `json:"llm_provider"`
	AdvancedIndexing *bool 		 `json:"advanced_indexing"`
	Rerank           *bool 		 `json:"rerank"`
	GenerateAnswers  *bool 		 `json:"generate_answers"`
}

func (r *KnowledgeExtractionRequest) validate() error {
	if r.ModelName == "" {
		return fmt.Errorf("model_name must be specified")
	}
	if len(r.Questions) == 0 {
		return fmt.Errorf("must provide at least one question to create knowledge extraction model")
	}

	questions := make(map[string]bool)
	for _, q := range r.Questions {
		qLower := strings.ToLower(q.Question)
		if questions[qLower] {
			return fmt.Errorf("duplicate question '%v' detected", q.Question)
		}
		questions[qLower] = true
	}

	defaultTrue := true
	if r.AdvancedIndexing == nil {
		r.AdvancedIndexing = &defaultTrue
	}
	if r.Rerank == nil {
		r.Rerank = &defaultTrue
	}
	if r.GenerateAnswers == nil {
		r.GenerateAnswers = &defaultTrue
	}

	if r.LlmProvider == "" && *r.GenerateAnswers {
		return fmt.Errorf("llm_provider must be be specifed unless generate_answers is false")
	}

	return nil
}

type Question struct {
	Id           string `gorm:"primaryKey"`
	QuestionText string `gorm:"uniqueIndex"`
	Keywords     []Keyword
}

type Keyword struct {
	Id          string `gorm:"primaryKey"`
	QuestionId  string
	KeywordText string
}

func populateQuestions(dbPath string, questions []QuestionKeywords) error {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		slog.Error("unable to open knowledge extraction metadata db", "error", err)
		return CodedError(errors.New("unable to open knowledge extraction metadata db"), http.StatusInternalServerError)
	}

	err = db.AutoMigrate(&Question{}, &Keyword{})
	if err != nil {
		slog.Error("unable to initialize knowledge extraction metadata db", "error", err)
		return CodedError(errors.New("unable to initialize knowledge extraction metadata db"), http.StatusInternalServerError)
	}

	entries := make([]Question, 0, len(questions))
	for _, question := range questions {
		entry := Question{
			Id:           uuid.New().String(),
			QuestionText: question.Question,
			Keywords:     make([]Keyword, 0, len(question.Keywords)),
		}
		for _, kw := range question.Keywords {
			entry.Keywords = append(entry.Keywords, Keyword{Id: uuid.New().String(), QuestionId: entry.Id, KeywordText: kw})
		}

		entries = append(entries, entry)
	}

	result := db.Create(&entries)
	if result.Error != nil {
		slog.Error("sql error populating knowledge extraction metadata", "error", result.Error)
		return CodedError(errors.New("error populating knowledge extraction metadata"), http.StatusInternalServerError)
	}

	return nil
}

func (s *WorkflowService) createQuestionDb(modelId uuid.UUID, questions []QuestionKeywords) error {
	dbPath := fmt.Sprintf("%v_metadata.db", modelId)

	defer func() {
		err := os.RemoveAll(dbPath)
		if err != nil {
			slog.Error("error cleaning up local knowledge extraction metadata db", "model_id", modelId, "error", err)
		}
	}()

	err := populateQuestions(dbPath, questions)
	if err != nil {
		return err
	}

	file, err := os.Open(dbPath)
	if err != nil {
		slog.Error("unable to open knowledge extraction metadata for copying", "error", err)
		return CodedError(errors.New("unable to open knowledge extraction metadata for copying"), http.StatusInternalServerError)
	}

	err = s.storage.Write(filepath.Join(storage.ModelPath(modelId), "model", "knowledge.db"), file)
	if err != nil {
		slog.Error("unable to copy knowledge extraction metadata", "error", err)
		return CodedError(errors.New("unable to copy knowledge extraction metadata"), http.StatusInternalServerError)
	}

	return nil
}

func (s *WorkflowService) KnowledgeExtraction(w http.ResponseWriter, r *http.Request) {
	var params KnowledgeExtractionRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := params.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	modelId := uuid.New()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		model := newModel(modelId, params.ModelName, schema.KnowledgeExtraction, nil, user.Id)

		model.Attributes = []schema.ModelAttribute{
			{ModelId: modelId, Key: "llm_provider", Value: params.LlmProvider},
			{ModelId: modelId, Key: "advanced_indexing", Value: strconv.FormatBool(*params.AdvancedIndexing)},
			{ModelId: modelId, Key: "rerank", Value: strconv.FormatBool(*params.Rerank)},
			{ModelId: modelId, Key: "generate_answers", Value: strconv.FormatBool(*params.GenerateAnswers)},
		}

		return saveModel(txn, model, user)
	})
	if err != nil {
		slog.Error("error creating knowledge extraction model", "error", err)
		http.Error(w, "error creating knowledge extraction model", GetResponseCode(err))
		return
	}

	err = s.createQuestionDb(modelId, params.Questions)
	if err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	result := s.db.Model(&schema.Model{}).Where("id = ?", modelId).Update("train_status", schema.Complete)
	if result.Error != nil {
		slog.Error("error updating knowledge extraction train status", "error", result.Error)
		http.Error(w, "error updating knowledge extraction train status", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, trainResponse{ModelId: modelId})
}
