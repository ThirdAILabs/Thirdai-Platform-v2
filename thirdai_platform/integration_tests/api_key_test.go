package integrationtests

import (
	"fmt"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"

	"github.com/google/uuid"
)

func createAPIKey(t *testing.T, c *client.PlatformClient, modelID uuid.UUID, prefix string, expiry time.Time) (string, uuid.UUID) {
	apiKeyName := fmt.Sprintf("%s-%s", prefix, modelID.String())

	apiKey, err := c.CreateAPIKey([]uuid.UUID{modelID}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("Failed to create API key for model %s: %v", modelID, err)
	}

	if apiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	apiKeys, err := c.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys after creating API key: %v", err)
	}

	var apiKeyID uuid.UUID
	found := false
	for _, key := range apiKeys {
		if key.Name == apiKeyName {
			apiKeyID = key.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("API key with name '%s' not found in the list of API keys", apiKeyName)
	}

	t.Cleanup(func() {
		err := c.DeleteAPIKey(apiKeyID)
		if err != nil {
			t.Errorf("Failed to delete API key (ID: %s): %v", apiKeyID, err)
		}
	})

	return apiKey, apiKeyID
}

func deployModelApiKey(t *testing.T, modelClient client.ModelClient, force bool, timeout time.Duration) {
	err := modelClient.Deploy(force)
	if err != nil {
		t.Fatal(err)
	}

	err = modelClient.AwaitDeploy(timeout)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := modelClient.Undeploy()
		if err != nil {
			t.Errorf("Failed to undeploy model (ID: %s): %v", modelClient.GetModelID(), err)
		}
	})
}

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

	now := time.Now()
	expiry := now.Add(24 * time.Hour)

	ndbModelID := ndb.ModelClient.GetModelID()
	ndbApiKey, _ := createAPIKey(t, c, ndbModelID, "test-api-key-ndb", expiry)
	err = ndb.ModelClient.UseApiKey(ndbApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}

	deployModelApiKey(t, ndb.ModelClient, false, 100*time.Second)

	checkQuery(ndb, t, true)

	err = nlp.ModelClient.UseApiKey(ndbApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}
	err = nlp.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key does not allow the NLP model, but got none")
	}

	nlpModelID := nlp.ModelClient.GetModelID()
	nlpApiKey, _ := createAPIKey(t, c, nlpModelID, "test-api-key-nlp", expiry)
	err = nlp.ModelClient.UseApiKey(nlpApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}

	deployModelApiKey(t, nlp.ModelClient, false, 100*time.Second)

	_, err = nlp.Predict("I really like to eat apples", 3)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.ModelClient.UseApiKey(nlpApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}
	err = ndb.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key does not allow the NDB model, but got none")
	}

	err = ndb.ModelClient.UseApiKey(ndbApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}
	deployModelApiKey(t, ndb.ModelClient, false, 100*time.Second)
}
