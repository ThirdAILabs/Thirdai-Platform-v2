package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"thirdai_platform/ndb"
	"time"
)

func basicTest() {
	ndb, err := ndb.New("./tmp2.db")
	if err != nil {
		log.Fatalf("failed to open ndb: %v", err)
	}
	defer ndb.Free()

	err = ndb.Insert("some_doc", "abc", []string{"a b c d e", "x y z", "b d", "e z"}, []map[string]interface{}{{"a": "a", "c": false}, {}, {"a": 1, "b": 2.3}, {}}, nil)
	if err != nil {
		log.Fatalf("insert failed: %v", err)
	}

	res, err := ndb.Query("a b c", 2, nil)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	fmt.Println(res)
}

type sample struct {
	label uint64
	query string
}

func loadQueries(filename string) []sample {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("failed to read csv: %v", err)
	}

	samples := make([]sample, 0)
	for _, row := range rows[1:] {
		label, err := strconv.ParseUint(row[0], 10, 64)
		if err != nil {
			log.Fatalf("failed to parse label: %v", err)
		}
		samples = append(samples, sample{label: label, query: row[1]})
	}

	return samples
}

func wiki() {
	ndb, err := ndb.New("../wiki_large.ndb")
	if err != nil {
		log.Fatalf("failed to open ndb: %v", err)
	}
	defer ndb.Free()

	samples := loadQueries("../../data/semantic_benchmarks/wiki/len_10.csv")

	correct := 0
	start := time.Now()

	for _, sample := range samples {
		results, err := ndb.Query(sample.query, 5, nil)
		if err != nil {
			log.Fatalf("ndb query failed: %v", err)
		}

		if len(results) > 0 && results[0].Id == sample.label {
			correct++
		}
	}

	end := time.Now()

	fmt.Printf("p@1=%f\n", float64(correct)/float64(len(samples)))
	fmt.Printf("avg_time=%f\n", float64(end.Sub(start).Milliseconds())/float64(len(samples)))
}

func main() {
	// wiki()
	basicTest()
}
