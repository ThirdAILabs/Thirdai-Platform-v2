package integrationtests

import (
	"fmt"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

// TODO(pratik): Add tests for rest of the models and their functions

func TestCreateUserModelAPIKeyDeployAndQuery(t *testing.T) {
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{{
			Path: "./data/articles.csv", Location: "upload",
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

	nlp, err := c.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// create token for accessing ndb model
	ndbModelId := []string{ndb.ModelClient.GetModelID().String()}

	apiKeyName := fmt.Sprintf("test-api-key-%s", ndb.ModelClient.GetModelID().String())
	expiry := "2026-01-31T23:59:59Z"

	ndbApiKey, err := c.CreateAPIKey(ndbModelId, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if ndbApiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	ndb.ModelClient.UseApiKey(ndbApiKey)

	// Now model client should use api key
	err = ndb.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(ndb, t, true)

	nlp.ModelClient.UseApiKey(ndbApiKey)
	err = nlp.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key doesnot allows the nlp model, but got none")
	}

	// create key for nlp and check both the accesses

	nlpModelID := []string{nlp.ModelClient.GetModelID().String()}

	nlpKeyName := fmt.Sprintf("test-api-key-%s", ndb.ModelClient.GetModelID().String())

	nlpApiKey, err := c.CreateAPIKey(nlpModelID, nlpKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if nlpApiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	nlp.ModelClient.UseApiKey(nlpApiKey)

	err = nlp.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nlp.Predict("i really like to eat apples", 3)
	if err != nil {
		t.Fatal(err)
	}

	ndb.ModelClient.UseApiKey(nlpApiKey)
	err = ndb.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key doesnot allows the ndb model, but got none")
	}
	// Reuse ndbAPIKey for deployment
	ndb.UseApiKey(ndbApiKey)
	t.Cleanup(func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Cleanup(func() {
		err := nlp.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})
}
