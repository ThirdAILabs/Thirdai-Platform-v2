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
		t.Fatal("no results returned for query")
	}

	return results[0]
}

func checkQuery(ndb *client.NdbClient, t *testing.T, checkId bool) {
	expectedText := "New Technology From AMD And IBM Boosts Chip Performance The manufacturing technology can result in faster chips"

	result := getResult(ndb, t, "manufacturing faster chips")
	if !strings.HasPrefix(result.Text, expectedText) {
		t.Fatal("incorrect result text returned")
	}

	if checkId && result.Id != 27 {
		t.Fatal("incorrect result id returned")
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

const (
	associateQuery  = "premier league teams in england"
	associateAnswer = "Eriksson: No Man Utd Arsenal rifts in squad England boss Sven Goran Eriksson insists"
)

func checkNoAssociate(ndb *client.NdbClient, t *testing.T) {
	if strings.Contains(getResult(ndb, t, associateQuery).Text, associateAnswer) {
		t.Fatal("query should not map to associate text before associate")
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
	if !strings.Contains(getResult(ndb, t, associateQuery).Text, associateAnswer) {
		t.Fatal("query should map to associated text after associate")
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
	files := make([]client.FileInfo, 0, len(fileNames))
	for _, file := range fileNames {
		files = append(files, client.FileInfo{Path: filepath.Join("./data/", file), Location: "upload"})
	}

	err := ndb.Insert(files)
	if err != nil {
		t.Fatal(err)
	}
}

func doDelete(ndb *client.NdbClient, t *testing.T) {
	sources := getSources(ndb, t)
	err := ndb.DeleteDocs([]string{sources[len(sources)-1].SourceId})
	if err != nil {
		t.Fatal(err)
	}
}

func checkSave(ndb *client.NdbClient, t *testing.T) {
	newNdb, err := ndb.Save(randomName("saved-ndb"))
	if err != nil {
		t.Fatal(err)
	}
	err = newNdb.AwaitTrain(30 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &newNdb.ModelClient, false)

	checkQuery(newNdb, t, false)
}

func createAndDeployNdb(t *testing.T, autoscaling bool) *client.NdbClient {
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{{
			Path: "./data/articles.csv", Location: "upload",
		}},
		nil,
		config.JobOptions{AllocationMemory: 1000},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &ndb.ModelClient, autoscaling)

	return ndb
}

func TestNdbDevMode(t *testing.T) {
	ndb := createAndDeployNdb(t, false)

	checkQuery(ndb, t, true)

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

	checkSave(ndb, t)
}

func TestNdbProdMode(t *testing.T) {
	c := getClient(t)

	baseNdb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{
			{Path: "./data/articles.csv", Location: "upload"},
			{Path: "./data/supervised.csv", Location: "upload"},
		},
		nil,
		config.JobOptions{AllocationMemory: 1000},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = baseNdb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &baseNdb.ModelClient, true)

	checkQuery(baseNdb, t, true)

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

	checkQuery(retrainedNdb, t, false)
	checkUpvote(retrainedNdb, t)
	checkAssociate(retrainedNdb, t)
	checkSources(retrainedNdb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx", "mutual_nda.pdf"})
}

func TestNdbTrainingFromBaseModel(t *testing.T) {
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{{
			Path: "./data/articles.csv", Location: "upload",
		}},
		nil,
		config.JobOptions{AllocationMemory: 1000},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	ndb2, err := c.TrainNdbWithBaseModel(
		randomName("ndb2"),
		ndb,
		[]client.FileInfo{{
			Path: "./data/four_english_words.docx", Location: "upload",
		}},
		[]client.FileInfo{{
			Path: "./data/supervised.csv", Location: "upload",
			Options: map[string]interface{}{
				"csv_query_column": "text",
				"csv_id_column":    "labels",
			},
		}},
		config.JobOptions{},
		false,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &ndb2.ModelClient, true)

	if !strings.Contains(getResult(ndb2, t, "manufacturing faster chips").Text, "AMD And IBM Boosts Chip Performance") {
		t.Fatal("incorrect document result")
	}

	if getResult(ndb2, t, "here is a new query that needs answering").Id != 4 {
		t.Fatal("incorrect supervised result")
	}
}

func createNdbAndInsert(t *testing.T, autoscaling bool) (*client.NdbClient, []client.Source, []client.Source) {
	ndb := createAndDeployNdb(t, autoscaling)

	oldSources, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.Insert([]client.FileInfo{{
		Path:     "./data/articles.csv",
		Location: "upload",
		SourceId: &oldSources[0].SourceId,
		Options:  map[string]interface{}{"upsert": true},
	}})
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

	deployModel(t, &retrainedNdb.ModelClient, false)

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

	model1, err := c.TrainNdb(
		randomName("ndb1"), []client.FileInfo{{
			Path: "./data/articles.csv", Location: "upload",
		}},
		nil,
		config.JobOptions{AllocationMemory: 1000},
	)
	if err != nil {
		t.Fatal(err)
	}

	model2, err := c.TrainNdb(
		randomName("ndb2"), []client.FileInfo{{
			Path: "./data/mutual_nda.pdf", Location: "upload",
		}},
		nil,
		config.JobOptions{AllocationMemory: 1000},
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
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{{Path: "./utils.go", Location: "upload"}},
		[]client.FileInfo{{Path: "./data/malformed.csv", Location: "upload"}},
		config.JobOptions{AllocationMemory: 1000},
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
	// Warning may be to far up in logs to find so we do not check for it here

	errorMsg := "The number of documents indexed and trained is 0"
	if len(status.Errors) < 1 || !strings.Contains(status.Errors[0], errorMsg) {
		t.Fatal("error not found in status messages")
	}
	if !strings.Contains(logs[0].Stderr, errorMsg) {
		t.Fatal("error not found in logs")
	}
}

func TestTrainWithGenerativeSupervision(t *testing.T) {
	c := getClient(t)

	ndb, err := c.TrainNdbWithGenerativeSupervision(
		randomName("ndb"),
		[]client.FileInfo{{
			Path: "./data/articles.csv", Location: "upload",
		}},
		nil,
		config.JobOptions{AllocationMemory: 1000},
		&config.LLMConfig{Provider: "mock"},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Just checking that training is completed successfully
	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
}
