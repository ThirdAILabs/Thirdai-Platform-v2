package deployment

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/search/ndb"
	"thirdai_platform/utils"
	"thirdai_platform/utils/llm_generation"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type NdbRouter struct {
	Ndb         ndb.NeuralDB
	Config      *config.DeployConfig
	Reporter    Reporter
	Permissions PermissionsInterface
	LLMCache    *LLMCache
	LLM         llm_generation.LLM
}

func NewNdbRouter(config *config.DeployConfig, reporter Reporter) (*NdbRouter, error) {
	ndbPath := filepath.Join(config.ModelBazaarDir, "models", config.ModelId.String(), "model", "model.ndb")
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ndb: %v", err)
	}

	var llmCache *LLMCache
	var llm llm_generation.LLM

	if provider, exists := config.Options["llm_provider"]; exists {
		// TODO api key should be passed as environment variable based on provider
		// rather than passing it in the /generate endpoint from the frontend
		// Same goes for model and provider
		llm, err = llm_generation.NewLLM(llm_generation.LLMProvider(provider), config.Options["genai_key"])
		if err != nil {
			return nil, err
		}

		llmCache, err = NewLLMCache(config.ModelBazaarDir, config.ModelId.String())
		if err != nil {
			return nil, err
		}

		llmCache, err = NewLLMCache(config.ModelBazaarDir, config.ModelId.String())
		if err != nil {
			return nil, err
		}
	}

	return &NdbRouter{
		Ndb:         ndb,
		Config:      config,
		Reporter:    reporter,
		Permissions: &Permissions{config.ModelBazaarEndpoint, config.ModelId},
		LLMCache:    llmCache,
		LLM:         llm,
	}, nil
}

func (s *NdbRouter) Close() {
	s.Ndb.Free()
	if s.LLMCache != nil {
		s.LLMCache.Close()
	}
}

func (s *NdbRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(os.Stderr, "", log.LstdFlags), NoColor: false,
	}))

	r.Group(func(r chi.Router) {
		r.Use(m.Permissions.ModelPermissionsCheck(WritePermission))

		r.Post("/insert", s.Insert)
		r.Post("/delete", s.Delete)
		r.Post("/upvote", s.Upvote)
		r.Post("/associate", s.Associate)
	})

	r.Group(func(r chi.Router) {
		r.Use(m.Permissions.ModelPermissionsCheck(ReadPermission))

		r.Post("/query", s.Search)
		r.Get("/sources", s.Sources)
		// r.Post("/implicit-feedback", s.ImplicitFeedback)
		// r.Get("/highlighted-pdf", s.HighlightedPdf)

		if s.LLMProvider != nil {
			r.Post("/generate", s.GenerateFromReferences)
		}

		if s.LLMCache != nil {
			r.Post("/cache-suggestions", s.CacheSuggestions)
			r.Post("/generation-cache", s.GenerationCache)
		}
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w)
	})

	return r
}

// TODO(any) add reranking and context radius options
type ConstraintInput struct {
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type SearchRequest struct {
	Query       string                     `json:"query"`
	Topk        int                        `json:"top_k"`
	Constraints map[string]ConstraintInput `json:"constraints,omitempty"`
}

type SearchResult struct {
	Id     int     `json:"id"`
	Text   string  `json:"text"`
	Source string  `json:"source"`
	Score  float32 `json:"score"`
}

type SearchResults struct {
	References []SearchResult `json:"references"`
}

func (s *NdbRouter) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	if req.Topk <= 0 {
		http.Error(w, "top_k must be greater than 0", http.StatusBadRequest)
		return
	}

	constraints := make(ndb.Constraints)
	for key, c := range req.Constraints {
		switch c.Op {
		case "eq":
			constraints[key] = ndb.EqualTo(c.Value)
		case "lt":
			constraints[key] = ndb.LessThan(c.Value)
		case "gt":
			constraints[key] = ndb.GreaterThan(c.Value)
		default:
			http.Error(w, fmt.Sprintf("invalid constraint operator '%s' for key '%s'", c.Op, key), http.StatusUnprocessableEntity)
			return
		}
	}

	chunks, err := s.Ndb.Query(req.Query, req.Topk, constraints)
	if err != nil {
		http.Error(w, "could not process query", http.StatusInternalServerError)
		return
	}

	results := SearchResults{References: make([]SearchResult, len(chunks))}
	for i, chunk := range chunks {
		results.References[i] = SearchResult{
			Id:     int(chunk.Id),
			Text:   chunk.Text,
			Source: chunk.Document,
			Score:  chunk.Score,
		}
	}

	utils.WriteJsonResponse(w, &results)
}

type InsertRequest struct {
	Document string                   `json:"document"`
	DocId    string                   `json:"doc_id"`
	Chunks   []string                 `json:"chunks"`
	Metadata []map[string]interface{} `json:"metadata,omitempty"`
	Version  *uint                    `json:"version,omitempty"`
}

// TODO how to do insert from files that already have been uploaded?
// do we need go bindings for documents or to parse them with a service beforehand?
func (s *NdbRouter) Insert(w http.ResponseWriter, r *http.Request) {
	var req InsertRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	if err := s.Ndb.Insert(req.Document, req.DocId, req.Chunks, req.Metadata, req.Version); err != nil {
		http.Error(w, fmt.Sprintf("insert error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

type DeleteRequest struct {
	DocIds            []string `json:"source_ids"`
	KeepLatestVersion bool     `json:"keep_latest_version"`
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	keepLatest := req.KeepLatestVersion

	for _, docID := range req.DocIds {
		if err := s.Ndb.Delete(docID, keepLatest); err != nil {
			http.Error(w, fmt.Sprintf("delete error for doc '%s': %v", docID, err), http.StatusInternalServerError)
			return
		}
	}

	utils.WriteSuccess(w)
}

type UpvoteInputSingle struct {
	QueryText     string `json:"query_text"`
	ReferenceId   int    `json:"reference_id"`
	ReferenceText string `json:"reference_text"`
}

type UpvoteInput struct {
	TextIdPairs []UpvoteInputSingle `json:"text_id_pairs"`
}

func (s *NdbRouter) Upvote(w http.ResponseWriter, r *http.Request) {
	var req UpvoteInput
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	queries := make([]string, len(req.TextIdPairs))
	labels := make([]uint64, len(req.TextIdPairs))
	for i, pair := range req.TextIdPairs {
		queries[i] = pair.QueryText
		labels[i] = uint64(pair.ReferenceId)
	}

	if err := s.Ndb.Finetune(queries, labels); err != nil {
		http.Error(w, fmt.Sprintf("upvote error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

type AssociateInputSingle struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type AssociateInput struct {
	TextPairs []AssociateInputSingle `json:"text_pairs"`
	Strength  *uint32                `json:"strength,omitempty"`
}

func (s *NdbRouter) Associate(w http.ResponseWriter, r *http.Request) {
	var req AssociateInput
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	var strength uint32 = 4
	if req.Strength != nil {
		strength = *req.Strength
	}

	sources := make([]string, len(req.TextPairs))
	targets := make([]string, len(req.TextPairs))
	for i, pair := range req.TextPairs {
		sources[i] = pair.Source
		targets[i] = pair.Target
	}

	if err := s.Ndb.Associate(sources, targets, strength); err != nil {
		http.Error(w, fmt.Sprintf("associate error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

type Source struct {
	Source   string `json:"source"`
	SourceID string `json:"source_id"`
	Version  uint32 `json:"version"`
}

type Sources struct {
	Sources []Source `json:"sources"`
}

// TODO(any) change the "source" field to return the full source path?
func (s *NdbRouter) Sources(w http.ResponseWriter, r *http.Request) {
	srcs, err := s.Ndb.Sources()
	if err != nil {
		http.Error(w, fmt.Sprintf("sources error: %v", err), http.StatusInternalServerError)
		return
	}

	sort.Slice(srcs, func(i, j int) bool {
		return srcs[i].Document < srcs[j].Document
	})

	results := Sources{Sources: make([]Source, len(srcs))}
	for i, doc := range srcs {
		results.Sources[i] = Source{
			Source:   doc.Document,
			SourceID: doc.DocId,
			Version:  doc.DocVersion,
		}
	}

	utils.WriteJsonResponse(w, results)
}

func (s *NdbRouter) ImplicitFeedback(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) HighlightedPdf(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) GenerateFromReferences(w http.ResponseWriter, r *http.Request) {
	var req llm_generation.GenerateRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	if s.LLM == nil {
		http.Error(w, "LLM provider not found", http.StatusInternalServerError)
		return
	}

	slog.Info("started generation", "query", req.Query)

	llmRes, err := s.LLM.StreamResponse(req, w, r)
	if err != nil {
		// Any error has already been sent to the client, just return
		return
	}

	slog.Info("completed generation", "query", req.Query, "llmRes", llmRes)

	referenceIds := make([]uint64, len(req.References))
	for i, ref := range req.References {
		referenceIds[i] = ref.Id
	}

	if s.LLMCache != nil {
		err = s.LLMCache.Insert(req.Query, llmRes, referenceIds)
		if err != nil {
			slog.Error("failed cache insertion", "error", err)
		}
	}
}

type CacheSuggestionsQuery struct {
	Query string `json:"query"`
}

func (s *NdbRouter) CacheSuggestions(w http.ResponseWriter, r *http.Request) {
	var req CacheSuggestionsQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	if s.LLMCache == nil {
		http.Error(w, "LLM cache is not initialized", http.StatusInternalServerError)
		return
	}

	suggestions, err := s.LLMCache.Suggestions(req.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("cache suggestions error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"suggestions": suggestions})
}

type CacheQuery struct {
	Query        string   `json:"query"`
	ReferenceIds []uint64 `json:"reference_ids"`
}

func (s *NdbRouter) GenerationCache(w http.ResponseWriter, r *http.Request) {
	var req CacheQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	if s.LLMCache == nil {
		http.Error(w, "LLM cache is not initialized", http.StatusInternalServerError)
		return
	}

	result, err := s.LLMCache.Query(req.Query, req.ReferenceIds)
	if err != nil {
		http.Error(w, fmt.Sprintf("cache query error: %v", err), http.StatusInternalServerError)
		return
	}

	if result == "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No cached result found",
			"result":  result,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}
