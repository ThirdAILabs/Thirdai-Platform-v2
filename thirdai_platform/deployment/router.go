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
	"thirdai_platform/utils/logging"

	slogmulti "github.com/samber/slog-multi"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	topKSelectionsToTrack = 5
	queryMetric           = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_query", Help: "NDB Queries"})
	upvoteMetric          = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_upvote", Help: "NDB Upvotes"})
	associateMetric       = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_associate", Help: "NDB Associations"})
	insertMetric          = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_insert", Help: "NDB Inserts"})
	deleteMetric          = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_delete", Help: "NDB Deletes"})

	implicitFeedbackMetric = promauto.NewSummary(prometheus.SummaryOpts{Name: "ndb_implicit_feedback", Help: "NDB Implicit Feedback"})

	// Add counters for tracking top-k selections
	ndbTopKSelections = make([]prometheus.Counter, topKSelectionsToTrack)
)

func init() {
	// inserting values in an init function to make sure prometheus metrics are initialized whenever router is imported
	for i := 0; i < topKSelectionsToTrack; i++ {
		ndbTopKSelections[i] = promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("ndb_result_%d_selections", i+1),
			Help: fmt.Sprintf("Number of selections of result %d by user.", i+1),
		})
	}
}

type NdbRouter struct {
	Ndb         ndb.NeuralDB
	Config      *config.DeployConfig
	Reporter    Reporter
	Permissions PermissionsInterface
	LLMCache    *LLMCache
	LLM         llm_generation.LLM
}

func InitLogging(logFile *os.File, config *config.DeployConfig) {
	// victoria logs option transform keys like msg and time into victoria log keys _msg and _time
	var jsonHandler slog.Handler = slog.NewJSONHandler(logFile, logging.GetVictoriaLogsOptions(true))

	// add default values to add to json logs
	// these fields will be used for filtering logs
	jsonHandler = jsonHandler.WithAttrs([]slog.Attr{
		slog.String("service_type", "deployment"),
		slog.String("model_id", config.ModelId.String()),
		slog.String("user_id", config.UserId.String()),
		slog.String("model_type", config.ModelType),
	})
	textHandler := slog.NewTextHandler(os.Stderr, nil)

	logger := slog.New(slogmulti.Fanout(jsonHandler, textHandler))
	slog.SetDefault(logger)
}

func NewNdbRouter(config *config.DeployConfig, reporter Reporter) (*NdbRouter, error) {
	ndbPath := filepath.Join(config.ModelBazaarDir, "models", config.ModelId.String(), "model", "model.ndb")
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		slog.Error("failed to open ndb", "error", err, "code", logging.MODEL_INIT)
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
		r.Use(s.Permissions.ModelPermissionsCheck(WritePermission))

		r.Post("/insert", s.Insert)
		r.Post("/delete", s.Delete)
		r.Post("/upvote", s.Upvote)
		r.Post("/associate", s.Associate)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.Permissions.ModelPermissionsCheck(ReadPermission))

		r.Post("/query", s.Search)
		r.Get("/sources", s.Sources)
		// r.Post("/implicit-feedback", s.ImplicitFeedback)
		// r.Get("/highlighted-pdf", s.HighlightedPdf)

		if s.LLM != nil {
			r.Post("/generate", s.GenerateFromReferences)
		}

		if s.LLMCache != nil {
			r.Post("/cache-suggestions", s.CacheSuggestions)
		}
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w)
	})

	r.Handle("/metrics", promhttp.Handler())

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
	// log time taken for serving the request
	timer := prometheus.NewTimer(queryMetric)
	defer timer.ObserveDuration()

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
			slog.Error("invalid constraint operator", "operator", c.Op, "key", key, "code", logging.MODEL_SEARCH)
			http.Error(w, fmt.Sprintf("invalid constraint operator '%s' for key '%s'", c.Op, key), http.StatusUnprocessableEntity)
			return
		}
	}

	chunks, err := s.Ndb.Query(req.Query, req.Topk, constraints)
	if err != nil {
		slog.Error("ndb query error", "error", err, "code", logging.MODEL_SEARCH)
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
	slog.Debug("searched ndb", "query", req.Query, "top_k", req.Topk, "code", logging.MODEL_SEARCH)
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
	timer := prometheus.NewTimer(insertMetric)
	defer timer.ObserveDuration()

	var req InsertRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	if err := s.Ndb.Insert(req.Document, req.DocId, req.Chunks, req.Metadata, req.Version); err != nil {
		slog.Error("insert error", "error", err, "code", logging.MODEL_INSERT)
		http.Error(w, fmt.Sprintf("insert error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
	slog.Info("inserted document", "doc_id", req.DocId, "code", logging.MODEL_INSERT)
}

type DeleteRequest struct {
	DocIds            []string `json:"source_ids"`
	KeepLatestVersion bool     `json:"keep_latest_version"`
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(deleteMetric)
	defer timer.ObserveDuration()

	var req DeleteRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	keepLatest := req.KeepLatestVersion

	for _, docID := range req.DocIds {
		if err := s.Ndb.Delete(docID, keepLatest); err != nil {
			slog.Error("delete error", "error", err, "doc_id", docID, "code", logging.MODEL_DELETE)
			http.Error(w, fmt.Sprintf("delete error for doc '%s': %v", docID, err), http.StatusInternalServerError)
			return
		}
	}

	utils.WriteSuccess(w)
	slog.Info("deleted documents", "doc_ids", req.DocIds, "code", logging.MODEL_DELETE)
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
	timer := prometheus.NewTimer(upvoteMetric)
	defer timer.ObserveDuration()

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
		slog.Error("upvote error", "error", err, "code", logging.MODEL_RLHF)
		http.Error(w, fmt.Sprintf("upvote error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
	slog.Debug("upvoted document", "text_id_pairs", req.TextIdPairs, "code", logging.MODEL_RLHF)
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
	timer := prometheus.NewTimer(associateMetric)
	defer timer.ObserveDuration()

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
		slog.Error("associate error", "error", err, "code", logging.MODEL_RLHF)
		http.Error(w, fmt.Sprintf("associate error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
	slog.Debug("associated text pairs", "code", logging.MODEL_RLHF)
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
		slog.Error("sources error", "error", err, "code", logging.MODEL_INFO)
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
	slog.Debug("retrieved sources", "sources", results.Sources, "code", logging.MODEL_INFO)
}

type ImplicitFeedback struct {
	QueryText     string `json:"query_text"`
	ReferenceId   int    `json:"reference_id"`
	EventDesc     string `json:"event_desc"`
	ReferenceRank int    `json:"reference_rank"`
}

func (s *NdbRouter) ImplicitFeedback(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(implicitFeedbackMetric)
	defer timer.ObserveDuration()

	var req ImplicitFeedback
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	var queries = ([]string{req.QueryText})
	var labels = []uint64{uint64(req.ReferenceId)}

	if err := s.Ndb.Finetune(queries, labels); err != nil {
		slog.Error("implicit feedback error", "error", err, "code", logging.MODEL_RLHF)
		http.Error(w, fmt.Sprintf("implicit feedback error: %v", err), http.StatusInternalServerError)
		return
	}

	// Increment the counter for the rank of the selected result
	if req.ReferenceRank > 0 && req.ReferenceRank <= topKSelectionsToTrack {
		ndbTopKSelections[req.ReferenceRank-1].Inc()
	}
	slog.Debug("implicit feedback received", "query_text", req.QueryText, "reference_id", req.ReferenceId, "event_desc", req.EventDesc, "reference_rank", req.ReferenceRank, "code", logging.MODEL_RLHF)
}

func (s *NdbRouter) HighlightedPdf(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) GenerateFromReferences(w http.ResponseWriter, r *http.Request) {
	slog.Info("generating from references")
	var req llm_generation.GenerateRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	// query the cache first
	if s.LLMCache != nil {
		cachedResult, err := s.FindCachedResult(req)
		if err != nil {
			slog.Error("cache error", "error", err)
		}

		if cachedResult != "" {
			// stream response rather than returning a json response
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, ok := w.(http.Flusher)
			if !ok {
				slog.Error("streaming unsupported")
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", cachedResult)
			flusher.Flush()
			return
		}

		slog.Info("no cached result found, generating", "query", req.Query)
	}

	// if no cached result found, generate from scratch
	if s.LLM == nil {
		slog.Error("LLM provider not found")
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

	err = json.NewEncoder(w).Encode(map[string]interface{}{"suggestions": suggestions})
	if err != nil {
		http.Error(w, fmt.Sprintf("cache suggestions error: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *NdbRouter) FindCachedResult(generateRequest llm_generation.GenerateRequest) (string, error) {
	if s.LLMCache == nil {
		return "", fmt.Errorf("LLM cache is not initialized")
	}

	referenceIds := make([]uint64, len(generateRequest.References))
	for i, ref := range generateRequest.References {
		referenceIds[i] = ref.Id
	}

	result, err := s.LLMCache.Query(generateRequest.Query, referenceIds)
	if err != nil {
		return "", fmt.Errorf("cache query error: %v", err)
	}

	if result == "" {
		return "", fmt.Errorf("no cached result found")
	}

	return result, nil
}
