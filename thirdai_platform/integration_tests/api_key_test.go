package integrationtests

import (
	"fmt"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"

	"github.com/google/uuid"
)

// TODO(pratik): Add tests for rest of the models and their functions

func TestNdbNlpModelsAPIKeyDeployAndQuery(t *testing.T) {
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName("ndb"),
		[]client.FileInfo{{
			Path:     "./data/articles.csv",
			Location: "upload",
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

	ndbModelID := []string{ndb.ModelClient.GetModelID().String()}
	apiKeyName := fmt.Sprintf("test-api-key-ndb-%s", ndb.ModelClient.GetModelID().String())
	expiry := "2026-01-31T23:59:59Z"

	ndbApiKey, err := c.CreateAPIKey(ndbModelID, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create NDB API key: %v", err)
	}

	if ndbApiKey == "" {
		t.Fatal("Expected a valid NDB API key, but got an empty string")
	}

	apiKeys, err := c.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys after creating NDB API key: %v", err)
	}

	var ndbApiKeyID uuid.UUID
	found := false
	for _, key := range apiKeys {
		if key.Name == apiKeyName {
			ndbApiKeyID = key.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("NDB API key with name '%s' not found in the list of API keys", apiKeyName)
	}

	// Register cleanup to delete the NDB API key
	t.Cleanup(func() {
		err := c.DeleteAPIKey(ndbApiKeyID)
		if err != nil {
			t.Errorf("Failed to delete NDB API key (ID: %d): %v", ndbApiKeyID, err)
		}
	})

	ndb.ModelClient.UseApiKey(ndbApiKey)

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
		t.Fatal("expected an error because the API key does not allow the NLP model, but got none")
	}

	nlpModelID := []string{nlp.ModelClient.GetModelID().String()}
	nlpKeyName := fmt.Sprintf("test-api-key-nlp-%s", nlp.ModelClient.GetModelID().String())

	nlpApiKey, err := c.CreateAPIKey(nlpModelID, nlpKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create NLP API key: %v", err)
	}

	if nlpApiKey == "" {
		t.Fatal("Expected a valid NLP API key, but got an empty string")
	}

	apiKeys, err = c.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys after creating NLP API key: %v", err)
	}

	var nlpApiKeyID uuid.UUID
	found = false
	for _, key := range apiKeys {
		if key.Name == nlpKeyName {
			nlpApiKeyID = key.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("NLP API key with name '%s' not found in the list of API keys", nlpKeyName)
	}

	t.Cleanup(func() {
		err := c.DeleteAPIKey(nlpApiKeyID)
		if err != nil {
			t.Errorf("Failed to delete NLP API key (ID: %d): %v", nlpApiKeyID, err)
		}
	})

	nlp.ModelClient.UseApiKey(nlpApiKey)

	err = nlp.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nlp.Predict("I really like to eat apples", 3)
	if err != nil {
		t.Fatal(err)
	}

	ndb.ModelClient.UseApiKey(nlpApiKey)
	err = ndb.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key does not allow the NDB model, but got none")
	}

	ndb.ModelClient.UseApiKey(ndbApiKey)
	err = ndb.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Errorf("Failed to undeploy NDB model: %v", err)
		}
	})

	t.Cleanup(func() {
		err := nlp.Undeploy()
		if err != nil {
			t.Errorf("Failed to undeploy NLP model: %v", err)
		}
	})
}
