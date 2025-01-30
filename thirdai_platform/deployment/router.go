package deployment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/search/ndb"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type NdbRouter struct {
	Ndb    ndb.NeuralDB
	Config *config.DeployConfig
	// reporter    Reporter
	// permissions Permissions
}

func NewNdbRouter(config *config.DeployConfig, reporter Reporter) (*NdbRouter, error) {
	ndbPath := filepath.Join(config.ModelBazaarDir, "models", config.ModelId.String(), "model", "model.ndb")
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ndb: %v", err)
	}
	return &NdbRouter{ndb, config}, nil
	// return &NdbRouter{ndb, config, reporter, Permissions{config.ModelBazaarEndpoint, config.ModelId}}, nil
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
		// r.Use(m.permissions.ModelPermissionsCheck("write"))

		r.Post("/insert", m.Insert)
		r.Post("/delete", m.Delete)
	})

	r.Group(func(r chi.Router) {
		// r.Use(m.permissions.ModelPermissionsCheck("read"))

		r.Post("/query", m.Search)
		r.Post("/upvote", m.Upvote)
		r.Post("/associate", m.Associate)
		r.Get("/sources", m.Sources)
		r.Post("/save", m.Save) // TODO Check low disk usage
		r.Post("/implicit-feedback", m.ImplicitFeedback)
		r.Get("/highlighted-pdf", m.HighlightedPdf)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
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
			http.Error(w, fmt.Sprintf("invalid constraint operator '%s' for key '%s'", c.Op, key), http.StatusBadRequest)
			return
		}
	}

	chunks, err := s.Ndb.Query(req.Query, req.Topk, constraints)
	if err != nil {
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

	if err := json.NewEncoder(w).Encode(&results); err != nil {
		http.Error(w, fmt.Sprintf("response encoding error: %v", err), http.StatusInternalServerError)
		return
	}
}

type InsertRequest struct {
	Document string                   `json:"document"`
	DocId    string                   `json:"doc_id"`
	Chunks   []string                 `json:"chunks"`
	Metadata []map[string]interface{} `json:"metadata,omitempty"`
	Version  *uint                    `json:"version,omitempty"`
}

func (s *NdbRouter) Insert(w http.ResponseWriter, r *http.Request) {
	var req InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
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
	KeepLatestVersion *bool    `json:"keep_latest_version,omitempty"`
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	keepLatest := false
	if req.KeepLatestVersion != nil {
		keepLatest = *req.KeepLatestVersion
	}

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
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

	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, fmt.Sprintf("response encoding error: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *NdbRouter) Save(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) ImplicitFeedback(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) HighlightedPdf(w http.ResponseWriter, r *http.Request) {
}
