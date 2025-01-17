package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"thirdai_platform/ndb"

	"github.com/go-chi/chi/v5"
)

type NDBServer struct {
	ndb ndb.NeuralDB
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

func (s *NDBServer) Search(w http.ResponseWriter, r *http.Request) {
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

func main() {
	var ndbPath string
	flag.StringVar(&ndbPath, "ndb_path", "", "the path to load the neural db from")
	flag.Parse()

	ndb, err := ndb.New(ndbPath)
	if err != nil {
		log.Fatalf("failed to open ndb: %v", err)
	}
	defer ndb.Free()

	server := NDBServer{ndb: ndb}

	r := chi.NewRouter()
	r.Post("/search", server.Search)

	log.Println("starting server on port 3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
