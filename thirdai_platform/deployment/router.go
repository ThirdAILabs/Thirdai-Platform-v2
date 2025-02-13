package deployment

import (
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
	"thirdai_platform/utils/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type NdbRouter struct {
	Ndb         ndb.NeuralDB
	Config      *config.DeployConfig
	Reporter    Reporter
	Permissions PermissionsInterface
}

func NewNdbRouter(config *config.DeployConfig, reporter Reporter) (*NdbRouter, error) {
	ndbPath := filepath.Join(config.ModelBazaarDir, "models", config.ModelId.String(), "model", "model.ndb")
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		slog.Error("failed to open ndb", "error", err, "code", logging.MODEL_INIT)
		return nil, fmt.Errorf("failed to open ndb: %v", err)
	}
	return &NdbRouter{ndb, config, reporter, &Permissions{config.ModelBazaarEndpoint, config.ModelId}}, nil
}

func (r *NdbRouter) Close() {
	r.Ndb.Free()
}

func (m *NdbRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(os.Stderr, "", log.LstdFlags), NoColor: false,
	}))

	r.Group(func(r chi.Router) {
		r.Use(m.Permissions.ModelPermissionsCheck("write"))

		r.Post("/insert", m.Insert)
		r.Post("/delete", m.Delete)
		r.Post("/upvote", m.Upvote)
		r.Post("/associate", m.Associate)
	})

	r.Group(func(r chi.Router) {
		r.Use(m.Permissions.ModelPermissionsCheck("read"))

		r.Post("/query", m.Search)
		r.Get("/sources", m.Sources)
		// r.Post("/implicit-feedback", m.ImplicitFeedback)
		// r.Get("/highlighted-pdf", m.HighlightedPdf)
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
			http.Error(w, fmt.Sprintf("invalid constraint operator '%s' for key '%s'", c.Op, key), http.StatusBadRequest)
			return
		}
	}

	chunks, err := s.Ndb.Query(req.Query, req.Topk, constraints)
	if err != nil {
		slog.Error("ndb query error", "error", err, "code", logging.MODEL_SEARCH)
		http.Error(w, fmt.Sprintf("ndb query error: %v", err), http.StatusInternalServerError)
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
		slog.Error("insert error", "error", err, "code", logging.MODEL_INSERT)
		http.Error(w, fmt.Sprintf("insert error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
	slog.Info("inserted document", "doc_id", req.DocId, "code", logging.MODEL_INSERT)
}

type DeleteRequest struct {
	DocIds            []string `json:"source_ids"`
	KeepLatestVersion *bool    `json:"keep_latest_version,omitempty"`
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return
	}

	keepLatest := false
	if req.KeepLatestVersion != nil {
		keepLatest = *req.KeepLatestVersion
	}

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
	slog.Debug("upvoted text pairs", "text_id_pairs", req.TextIdPairs, "code", logging.MODEL_RLHF)
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
		slog.Error("associate error", "error", err, "code", logging.MODEL_RLHF)
		http.Error(w, fmt.Sprintf("associate error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
	slog.Debug("associated text pairs", "text_pairs", req.TextPairs, "code", logging.MODEL_RLHF)
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

func (s *NdbRouter) ImplicitFeedback(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) HighlightedPdf(w http.ResponseWriter, r *http.Request) {
}
