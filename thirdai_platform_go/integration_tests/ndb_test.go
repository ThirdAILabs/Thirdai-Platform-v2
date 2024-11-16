package integrationtests

import (
	"path/filepath"
	"slices"
	"strings"
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

func createAndDeployNdb(t *testing.T, autoscaling bool) *client.NdbClient {
	client := getClient(t)

	ndb, err := client.TrainNdbWithJobOptions(
		randomName("ndb"),
		[]config.FileInfo{{
			Path: "./data/articles.csv", Location: "local",
		}},
		nil,
		config.JobOptions{AllocationMemory: 600},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.Deploy(autoscaling)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = ndb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	return ndb
}

func TestNdbDevMode(t *testing.T) {
	ndb := createAndDeployNdb(t, false)

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

	baseNdb, err := client.TrainNdbWithJobOptions(
		randomName("ndb"),
		[]config.FileInfo{
			{Path: "./data/articles.csv", Location: "local"},
			{Path: "./data/supervised.csv", Location: "local"},
		},
		nil,
		config.JobOptions{AllocationMemory: 600},
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

	t.Cleanup(func() {
		err := baseNdb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

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

	t.Cleanup(func() {
		err := retrainedNdb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = retrainedNdb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(retrainedNdb, t)
	checkUpvote(retrainedNdb, t)
	checkAssociate(retrainedNdb, t)
	checkSources(retrainedNdb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx", "mutual_nda.pdf"})
}

func createNdbAndInsert(t *testing.T, autoscaling bool) (*client.NdbClient, []client.Source, []client.Source) {
	ndb := createAndDeployNdb(t, autoscaling)

	oldSources, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.Insert([]config.FileInfo{
		{
			Path:     "./data/articles.csv",
			Location: "local",
			DocId:    &oldSources[0].SourceId,
			Options:  map[string]interface{}{"upsert": true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	newSources, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	return ndb, oldSources, newSources
}

func TestNdbUpsertDevMode(t *testing.T) {
	_, oldSources, newSources := createNdbAndInsert(t, false)

	if len(oldSources) != 1 || oldSources[0].Version != 1 {
		t.Fatalf("incorrect old sources: %v", oldSources)
	}

	if len(newSources) != 1 ||
		newSources[0].Version != 2 ||
		filepath.Base(newSources[0].Source) != filepath.Base(oldSources[0].Source) ||
		newSources[0].SourceId != oldSources[0].SourceId {
		t.Fatalf("incorrect new sources: %v", newSources)
	}
}

func TestNdbUpsertProdMode(t *testing.T) {
	ndb, oldSources, newSources := createNdbAndInsert(t, true)

	if len(oldSources) != 1 || oldSources[0].Version != 1 {
		t.Fatalf("incorrect old sources: %v", oldSources)
	}
	if newSources[0] != oldSources[0] {
		t.Fatal("old and new sources should match in prod mode")
	}

	retrainedNdb, err := ndb.Retrain(randomName("retrained-ndb"))
	if err != nil {
		t.Fatal(err)
	}

	err = retrainedNdb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = retrainedNdb.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := retrainedNdb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = retrainedNdb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	retrainedSources, err := retrainedNdb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	if len(retrainedSources) != 1 ||
		retrainedSources[0].Version != 2 ||
		filepath.Base(retrainedSources[0].Source) != filepath.Base(oldSources[0].Source) ||
		retrainedSources[0].SourceId != oldSources[0].SourceId {
		t.Fatalf("incorrect retrained sources: %v", newSources)
	}
}

func TestDeploymentName(t *testing.T) {
	c := getClient(t)

	model1, err := c.TrainNdbWithJobOptions(
		randomName("ndb1"), []config.FileInfo{{
			Path: "./data/articles.csv", Location: "local",
		}},
		nil,
		config.JobOptions{AllocationMemory: 600},
	)
	if err != nil {
		t.Fatal(err)
	}

	model2, err := c.TrainNdbWithJobOptions(
		randomName("ndb2"), []config.FileInfo{{
			Path: "./data/mutual_nda.pdf", Location: "local",
		}},
		nil,
		config.JobOptions{AllocationMemory: 600},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model1.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = model2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deploymentName := "custom_deployment_name"

	testDeployment := func(ndb *client.NdbClient, doc string) {
		err := ndb.DeployWithName(false, deploymentName)
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

		c := ndb.ClientForDeployment(deploymentName)

		sources, err := c.Sources()
		if err != nil {
			t.Fatal(err)
		}
		if len(sources) != 1 || filepath.Base(sources[0].Source) != doc {
			t.Fatalf("incorrect source: %v", sources)
		}

		_, err = c.Search("American Express Profit Rises 14", 4)
		if err != nil {
			t.Fatal(err)
		}
	}

	testDeployment(model1, "articles.csv")
	testDeployment(model2, "mutual_nda.pdf")
}

func TestTrainErrorHandling(t *testing.T) {
	client := getClient(t)

	ndb, err := client.TrainNdb(
		randomName("ndb"),
		[]config.FileInfo{{Path: "./utils.go", Location: "local"}},
		[]config.FileInfo{{Path: "./data/malformed.csv", Location: "local"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err == nil {
		t.Fatal("training should fail on malformed files")
	}

	status, err := ndb.TrainStatus()
	if err != nil {
		t.Fatal(err)
	}

	if status.Status != "failed" {
		t.Fatal("job should have status failed")
	}

	logs, err := ndb.TrainLogs()
	if err != nil {
		t.Fatal(err)
	}

	warningMsg := "utils.go. Unsupported filetype"
	if len(status.Warnings) < 1 || !strings.Contains(status.Warnings[0], warningMsg) {
		t.Fatal("warning not found in status messages")
	}
	if !strings.Contains(logs[0].Stderr, warningMsg) {
		t.Fatal("warning not found in logs")
	}

	errorMsg := "Error tokenizing data. C error:"
	if len(status.Errors) < 1 || !strings.Contains(status.Errors[0], errorMsg) {
		t.Fatal("error not found in status messages")
	}
	if !strings.Contains(logs[0].Stderr, errorMsg) {
		t.Fatal("error not found in logs")
	}

}
