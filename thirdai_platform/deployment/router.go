package deployment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"thirdai_platform/ndb"
	"thirdai_platform/utils"

	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/client"

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
}

func (s *NdbRouter) Delete(w http.ResponseWriter, r *http.Request) {
}

type searchRequest struct {
	Query string `json:"query"`
	Topk  int    `json:"top_k"`
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

func (s *NdbRouter) Search(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("request parsing error: %v", err), http.StatusBadRequest)
		return
	}

	chunks, err := s.ndb.Query(req.Query, req.Topk)
	if err != nil {
		http.Error(w, fmt.Sprintf("ndb query error: %v", err), http.StatusInternalServerError)
		return
	}

	results := searchResults{References: make([]searchResult, 0, len(chunks))}
	for _, chunk := range chunks {
		results.References = append(results.References, searchResult{
			Id:     int(chunk.Id),
			Text:   chunk.Text,
			Source: chunk.Document,
			Score:  chunk.Score,
		})
	}

	if err := json.NewEncoder(w).Encode(&results); err != nil {
		http.Error(w, fmt.Sprintf("error writing request: %v", err), http.StatusBadRequest)
		return
	}
}

func (s *NdbRouter) Upvote(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) Associate(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) ImplicitFeedback(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) Sources(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) Save(w http.ResponseWriter, r *http.Request) {
}

func (s *NdbRouter) HighlightedPdf(w http.ResponseWriter, r *http.Request) {
}
