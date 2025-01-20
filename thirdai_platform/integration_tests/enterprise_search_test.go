package integrationtests

import (
	"fmt"
	"strings"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestEnterpriseSearchWithGuardrail(t *testing.T) {
	c := getClient(t)

	model, err := c.UploadModel(randomName("basic_guardrail"), "./models/phone_guardrail.zip")
	if err != nil {
		t.Fatal(err)
	}

	guardrail := model.(*client.NlpTokenClient)

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

	es, err := c.CreateEnterpriseSearchWorkflow(randomName("es"), ndb, guardrail)
	if err != nil {
		t.Fatal(err)
	}

	err = es.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Cleanup(func() {
		err := guardrail.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Cleanup(func() {
		err := es.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = es.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	query := "American Express Profit Rises 14. my phone number is 123-457-2490"

	results, err := es.Search(query, 5)
	if err != nil {
		t.Fatal(err)
	}

	if results.QueryText == query {
		t.Fatal("query should be redacted")
	}

	if results.QueryText != strings.Replace(query, "123-457-2490", "[PHONENUMBER#0]", -1) {
		t.Fatalf("invalid redacted query: %v", results.QueryText)
	}

	unredacted, err := es.Unredact(results.QueryText, results.PiiEntities)
	if err != nil {
		t.Fatal(err)
	}

	if unredacted != query {
		t.Fatal("unredacted query should match original")
	}

	// use API Key to deploy query
	enterpriseSearchModelIds := []string{es.GetModelID().String()}

	apiKeyName := fmt.Sprintf("test-api-key-%s", es.GetModelID().String())
	expiry := "2026-01-31T23:59:59Z"

	eskApiKey, err := c.CreateAPIKey(enterpriseSearchModelIds, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if eskApiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	es.ModelClient.UseApiKey(eskApiKey)

	results, err = es.Search(query, 5)
	if err != nil {
		t.Fatal(err)
	}

	if results.QueryText != strings.Replace(query, "123-457-2490", "[PHONENUMBER#0]", -1) {
		t.Fatalf("invalid redacted query: %v", results.QueryText)
	}

	unredacted, err = es.Unredact(results.QueryText, results.PiiEntities)
	if err != nil {
		t.Fatal(err)
	}

	if unredacted != query {
		t.Fatal("unredacted query should match original")
	}
}
