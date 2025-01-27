package deployment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"path/filepath"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/search/ndb"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type NdbRouter struct {
	ndb         ndb.NeuralDB
	config      *config.DeployConfig
	reporter    client.ModelClient
	permissions Permissions
}

func NewNdbRouter(config *config.DeployConfig, reporter client.ModelClient) (*NdbRouter, error) {
	ndbPath := filepath.Join(config.ModelBazaarDir, "model", config.ModelId.String(), "model", "model.ndb")
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ndb: %v", err)
	}
	return &NdbRouter{ndb, config, reporter, Permissions{config.ModelBazaarEndpoint, config.ModelId}}, nil
}

func (r *NdbRouter) Close() {
	r.ndb.Free()
}

func (m *NdbRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(os.Stderr, "", log.LstdFlags), NoColor: false,
	}))

	r.Group(func(r chi.Router) {
		r.Use(m.permissions.ModelPermissionsCheck("write"))

		r.Post("/insert", m.Insert)
		r.Post("/delete", m.Delete)
	})

	r.Group(func(r chi.Router) {
		r.Use(m.permissions.ModelPermissionsCheck("read"))

		r.Post("/query", m.Search)
		r.Post("/upvote", m.Upvote)
		r.Post("/associate", m.Associate)
		r.Post("/implicit-feedback", m.ImplicitFeedback)
		r.Get("/sources", m.Sources)
		r.Post("/save", m.Save) // TODO Check low disk usage
		r.Get("/highlighted-pdf", m.HighlightedPdf)

	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w)
	})

	return r
}

func (s *NdbRouter) Insert(w http.ResponseWriter, r *http.Request) {
	type insertRequest struct {
		Document string                   `json:"document"`
		DocId    string                   `json:"doc_id"`
		Chunks   []string                 `json:"chunks"`
		Metadata []map[string]interface{} `json:"metadata,omitempty"`
		Version  *uint                    `json:"version,omitempty"`
	}

	var req insertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.ndb.Insert(req.Document, req.DocId, req.Chunks, req.Metadata, req.Version); err != nil {
		http.Error(w, fmt.Sprintf("insert error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
	type deleteRequest struct {
		DocIds            []string `json:"source_ids"`
		KeepLatestVersion *bool    `json:"keep_latest_version,omitempty"`
	}

	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	keepLatest := false
	if req.KeepLatestVersion != nil {
		keepLatest = *req.KeepLatestVersion
	}

	for _, docID := range req.DocIds {
		if err := s.ndb.Delete(docID, keepLatest); err != nil {
			http.Error(w, fmt.Sprintf("delete error for doc '%s': %v", docID, err), http.StatusInternalServerError)
			return
		}
	}

	utils.WriteSuccess(w)
}

func (s *NdbRouter) Search(w http.ResponseWriter, r *http.Request) {
	//TODO(any) add reranking and context radius options
	type constraintInput struct {
		Op    string      `json:"op"`
		Value interface{} `json:"value"`
	}
	type searchRequest struct {
		Query       string                     `json:"query"`
		Topk        int                        `json:"top_k"`
		Constraints map[string]constraintInput `json:"constraints,omitempty"`
	}
	type searchResult struct {
		Id     int     `json:"id"`
		Text   string  `json:"text"`
		Source string  `json:"source"`
		Score  float32 `json:"score"`
	}
	type searchResults struct {
		References []searchResult `json:"references"`
	}

	var req searchRequest
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

	chunks, err := s.ndb.Query(req.Query, req.Topk, constraints)
	if err != nil {
		http.Error(w, fmt.Sprintf("ndb query error: %v", err), http.StatusInternalServerError)
		return
	}

	results := searchResults{References: make([]searchResult, len(chunks))}
	for i, chunk := range chunks {
		results.References[i] = searchResult{
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

func (s *NdbRouter) Upvote(w http.ResponseWriter, r *http.Request) {
	type upvoteInputSingle struct {
		QueryText     string `json:"query_text"`
		ReferenceId   int    `json:"reference_id"`
		ReferenceText string `json:"reference_text"`
	}
	type upvoteInput struct {
		TextIdPairs []upvoteInputSingle `json:"text_id_pairs"`
	}

	var req upvoteInput
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

	if err := s.ndb.Finetune(queries, labels); err != nil {
		http.Error(w, fmt.Sprintf("upvote error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

func (s *NdbRouter) Associate(w http.ResponseWriter, r *http.Request) {
	type associateInputSingle struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	type associateInput struct {
		TextPairs []associateInputSingle `json:"text_pairs"`
	}

	var req associateInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	sources := make([]string, len(req.TextPairs))
	targets := make([]string, len(req.TextPairs))
	for i, pair := range req.TextPairs {
		sources[i] = pair.Source
		targets[i] = pair.Target
	}

	if err := s.ndb.Associate(sources, targets); err != nil {
		http.Error(w, fmt.Sprintf("associate error: %v", err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

//TODO(any) change the "source" field to return the full source path?
func (s *NdbRouter) Sources(w http.ResponseWriter, r *http.Request) {
	srcs, err := s.ndb.Sources()
	if err != nil {
		http.Error(w, fmt.Sprintf("sources error: %v", err), http.StatusInternalServerError)
		return
	}

	sort.Slice(srcs, func(i, j int) bool {
		return srcs[i].Document < srcs[j].Document
	})

	type sourceOutput struct {
		Source   string `json:"source"`
		SourceID string `json:"source_id"`
		Version  uint32 `json:"version"`
	}

	results := make([]sourceOutput, len(srcs))
	for i, doc := range srcs {
		results[i] = sourceOutput{
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
