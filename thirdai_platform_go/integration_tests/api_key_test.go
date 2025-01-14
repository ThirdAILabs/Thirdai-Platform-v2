package integrationtests

import (
	"testing"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestCreateUserModelAPIKeyDeployAndQuery(t *testing.T) {
	client := getClient(t)

	ndb, err := client.TrainNdb(
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

	nlp, err := client.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]config.FileInfo{{Path: "./data/ner.csv", Location: "local"}},
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
	selectedModelIDs := []string{ndb.modelId.string()}

	apiKeyName := "test-api-key"
	expiry := "2025-01-31T23:59:59Z"

	apiKey, err := client.createAPIKey(selectedModelIDs, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if apiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	client.UseApiKey(apiKey)

	// Now model client should use api key
	err = ndb.Deploy(false)
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

	checkQuery(ndb, t, true)

	// Now deploy nlp, which should fail beacause apiKey doesnot include the required Permission
	err = nlp.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key doesnot allows the nlp model, but got none")
	}

	expireApiKeyName := "expire-test-api-key"
	oldExpiry := "2025-01-00T23:59:59Z"

	expiredApiKey, err := client.createAPIKey(selectedModelIDs, expireApiKeyName, oldExpiry)

	client.UseApiKey(expiredApiKey)

	// Now model client should use api key
	err = ndb.Deploy(false)
	if err == nil {
		t.Fatal("expected an error because the API key is expired, but got none")
	}
}
