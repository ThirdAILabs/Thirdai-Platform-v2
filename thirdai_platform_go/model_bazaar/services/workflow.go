package services

import (
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
	"thirdai_platform/model_bazaar/utils"

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

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		err := checkForDuplicateModel(txn, params.ModelName, user.Id)
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

			perm, err := auth.GetModelPermissions(model.Id, user, txn)
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

		model := createModel(modelId, params.ModelName, schema.EnterpriseSearch, nil, user.Id)
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

type QuestionKeywords struct {
	Question string   `json:"question"`
	Keywords []string `json:"keywords"`
}

type KnowledgeExtractionRequest struct {
	ModelName string             `json:"model_name"`
	Questions []QuestionKeywords `json:"questions"`

	LlmProvider      string
	AdvancedIndexing *bool `json:"advanced_indexing"`
	Rerank           *bool `json:"rerank"`
	GenerateAnswers  *bool `json:"generate_answers"`
}

func (r *KnowledgeExtractionRequest) validate() error {
	if r.ModelName == "" {
		return fmt.Errorf("model_name must be specified")
	}
	if r.Questions == nil || len(r.Questions) == 0 {
		return fmt.Errorf("Must provide at least one question to create knowledge extraction model")
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
		return fmt.Errorf("unable to open knowledge extraction metadata db: %w", err)
	}

	err = db.AutoMigrate(&Question{}, &Keyword{})
	if err != nil {
		return fmt.Errorf("unable to initialize knowledge extraction metadata db: %w", err)
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
		return schema.NewDbError("populating knowledge extraction metadata", result.Error)
	}

	return nil
}

func (s *WorkflowService) createQuestionDb(modelId string, questions []QuestionKeywords) error {
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
		return fmt.Errorf("unable to open knowledge extraction metadata for copying to share: %w", err)
	}

	err = s.storage.Write(filepath.Join(storage.ModelPath(modelId), "knowledge.db"), file)
	if err != nil {
		return fmt.Errorf("unable to copy knowledge extraction metadata to share: %w", err)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := params.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		err := checkForDuplicateModel(txn, params.ModelName, user.Id)
		if err != nil {
			return err
		}

		model := createModel(modelId, params.ModelName, schema.KnowledgeExtraction, nil, user.Id)

		model.Attributes = []schema.ModelAttribute{
			{ModelId: modelId, Key: "llm_provider", Value: params.LlmProvider},
			{ModelId: modelId, Key: "advanced_indexing", Value: strconv.FormatBool(*params.AdvancedIndexing)},
			{ModelId: modelId, Key: "rerank", Value: strconv.FormatBool(*params.Rerank)},
			{ModelId: modelId, Key: "generate_answers", Value: strconv.FormatBool(*params.GenerateAnswers)},
		}

		result := txn.Create(&model)
		if result.Error != nil {
			return schema.NewDbError("create knowledge extraction model", result.Error)
		}

		return nil
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating knowledge extraction model: %v", err), http.StatusBadRequest)
		return
	}

	err = s.createQuestionDb(modelId, params.Questions)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to populate knowledge extraction metadata: %v", err), http.StatusInternalServerError)
		return
	}

	result := s.db.Model(&schema.Model{}).Where("id = ?", modelId).Update("train_status", schema.Complete)
	if result.Error != nil {
		err := schema.NewDbError("updating model train status", result.Error)
		http.Error(w, fmt.Sprintf("finalizing knowledge extraction initialization: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, map[string]string{"model_id": modelId})
}
