package integrationtests

import (
	"path/filepath"
	"slices"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func getResult(ndb *client.NdbClient, t *testing.T, query string) client.NdbSearchResult {
	results, err := ndb.Search(query, 4)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) < 1 {
		t.Fatal("incorrect results returned")
	}

	return results[0]
}

func checkQuery(ndb *client.NdbClient, t *testing.T) {
	if getResult(ndb, t, "manufacturing faster chips").Id != 27 {
		t.Fatal("incorrect results returned")
	}
}

const upvoteQuery = "some random nonsense with no relevance to any article"

func checkNoUpvote(ndb *client.NdbClient, t *testing.T) {
	if getResult(ndb, t, upvoteQuery).Id == 78 {
		t.Fatal("query should not map to upvoted id before upvote")
	}
}

func doUpvote(ndb *client.NdbClient, t *testing.T) {
	err := ndb.Upvote([]client.UpvotePair{{QueryText: upvoteQuery, ReferenceId: 78}})
	if err != nil {
		t.Fatal(err)
	}
}

func checkUpvote(ndb *client.NdbClient, t *testing.T) {
	if getResult(ndb, t, upvoteQuery).Id != 78 {
		t.Fatal("query should map to upvoted id after upvote")
	}
}

const associateQuery = "premier league teams in england"

func checkNoAssociate(ndb *client.NdbClient, t *testing.T) {
	if getResult(ndb, t, associateQuery).Id == 16 {
		t.Fatal("query should not map to upvoted id before upvote")
	}
}

func doAssociate(ndb *client.NdbClient, t *testing.T) {
	target := "man utd manchester united arsenal"
	err := ndb.Associate([]client.AssociatePair{{Source: associateQuery, Target: target}})
	if err != nil {
		t.Fatal(err)
	}
}

func checkAssociate(ndb *client.NdbClient, t *testing.T) {
	if getResult(ndb, t, associateQuery).Id != 16 {
		t.Fatal("query should map to upvoted id after upvote")
	}
}

func getSources(ndb *client.NdbClient, t *testing.T) []client.Source {
	sources, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	slices.SortFunc(sources, func(a, b client.Source) int {
		name1, name2 := filepath.Base(a.Source), filepath.Base(b.Source)
		if name1 < name2 {
			return -1
		}
		if name1 > name2 {
			return 1
		}
		return 0
	})

	return sources
}

func checkSources(ndb *client.NdbClient, t *testing.T, expectedSources []string) {
	sources := getSources(ndb, t)

	if len(sources) != len(expectedSources) {
		t.Fatalf("expected %d sources, got %d", len(expectedSources), len(sources))
	}

	for i := range expectedSources {
		if name := filepath.Base(sources[i].Source); name != expectedSources[i] {
			t.Fatalf("expected source %d to be %v, got %v", i, expectedSources[i], name)
		}
	}
}

func doInsert(ndb *client.NdbClient, t *testing.T, fileNames []string) {
	files := make([]config.FileInfo, 0, len(fileNames))
	for _, file := range fileNames {
		files = append(files, config.FileInfo{Path: filepath.Join("./data/", file), Location: "local"})
	}

	err := ndb.Insert(files)
	if err != nil {
		t.Fatal(err)
	}
}

func doDelete(ndb *client.NdbClient, t *testing.T) {
	sources := getSources(ndb, t)
	err := ndb.Delete([]string{sources[len(sources)-1].SourceId})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNdbDevMode(t *testing.T) {
	client := getClient(t)

	ndb, err := client.TrainNdb(
		randomName("ndb"), []config.FileInfo{{
			Path: "./data/articles.csv", Location: "local",
		}},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = ndb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(ndb, t)

	checkNoUpvote(ndb, t)
	doUpvote(ndb, t)
	checkUpvote(ndb, t)

	checkNoAssociate(ndb, t)
	doAssociate(ndb, t)
	checkAssociate(ndb, t)

	checkSources(ndb, t, []string{"articles.csv"})
	doInsert(ndb, t, []string{"four_english_words.docx", "mutual_nda.pdf", "articles.csv"})
	checkSources(ndb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx", "mutual_nda.pdf"})

	doDelete(ndb, t)
	checkSources(ndb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx"})
}

func TestNdbProdMode(t *testing.T) {
	client := getClient(t)

	baseNdb, err := client.TrainNdb(
		randomName("base-ndb"), []config.FileInfo{
			{Path: "./data/articles.csv", Location: "local"},
			{Path: "./data/supervised.csv", Location: "local"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = baseNdb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = baseNdb.Deploy(true)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := baseNdb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = baseNdb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(baseNdb, t)

	checkNoUpvote(baseNdb, t)
	doUpvote(baseNdb, t)
	checkNoUpvote(baseNdb, t)

	checkNoAssociate(baseNdb, t)
	doAssociate(baseNdb, t)
	checkNoAssociate(baseNdb, t)

	checkSources(baseNdb, t, []string{"articles.csv", "supervised.csv"})
	doDelete(baseNdb, t)
	checkSources(baseNdb, t, []string{"articles.csv", "supervised.csv"})
	doInsert(baseNdb, t, []string{"four_english_words.docx", "mutual_nda.pdf", "articles.csv"})
	checkSources(baseNdb, t, []string{"articles.csv", "supervised.csv"})

	retrainedNdb, err := baseNdb.Retrain(randomName("retrained-ndb"))
	if err != nil {
		t.Fatal(err)
	}

	err = retrainedNdb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = retrainedNdb.Deploy(true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := retrainedNdb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = retrainedNdb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(retrainedNdb, t)
	checkUpvote(retrainedNdb, t)
	checkAssociate(retrainedNdb, t)
	checkSources(retrainedNdb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx", "mutual_nda.pdf"})

}
