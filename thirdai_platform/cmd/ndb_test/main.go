package main

import (
	"fmt"
	"log"
	"thirdai_platform/ndb"
)

func main() {
	ndb, err := ndb.New("./tmp2.db")
	if err != nil {
		log.Fatalf("failed to open ndb: %v", err)
	}

	err = ndb.Insert("some_doc", "abc", []string{"a b c d e", "x y z", "b d", "e z"}, nil)
	if err != nil {
		log.Fatalf("insert failed: %v", err)
	}

	res, err := ndb.Query("a b c", 2)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	fmt.Println(res)

	defer ndb.Free()
}
