package tests

import (
	"errors"
	"testing"
	"time"

	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/services"

	"github.com/google/uuid"
)

func TestAPIKeyBasicCRUD(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("user1")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("test-api-key-model")
	if err != nil {
		t.Fatal(err)
	}

	keyName := "my-api-key"
	expiry := time.Now().Add(24 * time.Hour)

	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, keyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}
	if apiKeyVal == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	userKeys, err := user.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys: %v", err)
	}

	var createdKey *services.APIKeyResponse
	for _, key := range userKeys {
		if key.Name == keyName {
			createdKey = &key
			break
		}
	}
	if createdKey == nil {
		t.Fatalf("Expected to find API key '%s' in the list, but did not", keyName)
	}
	if createdKey.Expiry.IsZero() || createdKey.Expiry.Before(time.Now()) {
		t.Fatalf("Invalid expiry time on created key: %v", createdKey.Expiry)
	}

	err = user.DeleteAPIKey(createdKey.ID)
	if err != nil {
		t.Fatalf("Failed to delete API key: %v", err)
	}

	userKeys, err = user.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys after deletion: %v", err)
	}
	for _, key := range userKeys {
		if key.Name == keyName {
			t.Fatalf("API key '%s' was not deleted properly", keyName)
		}
	}
}

func TestAPIKeyPermissions(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("userA")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("key-perm-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "perm-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	checkPermissions(user, t, modelID, true, true, true)
	if err != nil {
		t.Fatalf("Test Failed: %v", err)
	}

}

func TestAPIKeyModelInfo(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("userB")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("key-info-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "info-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	info, err := user.modelInfo(modelID)
	if err != nil {
		t.Fatalf("Failed to access model info with API key: %v", err)
	}

	if info.ModelId.String() != modelID {
		t.Fatalf("Expected model_id %s, got %s", modelID, info.ModelId.String())
	}
	if info.ModelName != "key-info-model" {
		t.Fatalf("Expected model name 'key-info-model', got %s", info.ModelName)
	}

	_, err = user.trainNdbDummyFile("other-model")
	if err == nil {
		t.Fatalf("Cannot train a model using API-KEY")
	}

}

func TestAPIKeyExpiration(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("exp-user")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("expire-key-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "expired-key"
	expiry := time.Now().Add(-1 * time.Hour)
	_, err = user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry)
	if err == nil {
		t.Fatal("Expected error when creating an already expired API key, but got success.")
	}
}

func TestAPIKeyDeletedNoAccess(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("delete-user")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("del-key-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "delete-me-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err != nil {
		t.Fatalf("API key should have worked, but got error: %v", err)
	}

	user.login(loginInfo{Email: "delete-user@mail.com", Password: "delete-user_password"})
	userKeys, err := user.ListAPIKeys()
	if err != nil {
		t.Fatalf("Failed to list API keys: %v", err)
	}

	var keyToDelete uuid.UUID
	for _, k := range userKeys {
		if k.Name == apiKeyName {
			keyToDelete = k.ID
			break
		}
	}
	if keyToDelete == uuid.Nil {
		t.Fatalf("Created key '%s' not found in list", apiKeyName)
	}

	err = user.DeleteAPIKey(keyToDelete)
	if err != nil {
		t.Fatalf("Failed to delete API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err == nil {
		t.Fatal("Expected an error when using a deleted API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized, got: %v", err)
	}
}

func TestAPIKeyDependencies(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("dep-user")
	if err != nil {
		t.Fatal(err)
	}

	ndbID, err := user.trainNdbDummyFile("my-ndb-dep")
	if err != nil {
		t.Fatalf("Failed to train NDB model: %v", err)
	}

	nlpID, err := user.trainNlpToken("my-nlp-dep")
	if err != nil {
		t.Fatalf("Failed to train NLP model: %v", err)
	}

	err = user.updateAccess(nlpID, schema.Public)
	if err != nil {
		t.Fatal(err)
	}

	esID, err := user.createEnterpriseSearch("my-es-model", ndbID, nlpID)
	if err != nil {
		t.Fatalf("Failed to create Enterprise Search model: %v", err)
	}

	apiKeyName := "dep-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(esID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	checkPermissions(user, t, ndbID, true, true, true)
	checkPermissions(user, t, nlpID, true, true, true)
	checkPermissions(user, t, esID, true, true, true)
}

func TestAPIKeyUsageAfterExpiration(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("exp-user")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("expire-key-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "temp-key"
	expiry := time.Now().Add(2 * time.Second)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err != nil {
		t.Fatalf("API key should have worked before expiration, but got error: %v", err)
	}

	time.Sleep(3 * time.Second)

	_, err = user.modelPermissions(modelID)
	if err == nil {
		t.Fatal("Expected an error when using an expired API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized after expiration, got: %v", err)
	}
}

func TestAPIKeyAccessDifferentModel(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("access-user")
	if err != nil {
		t.Fatal(err)
	}

	model1ID, err := user.trainNdbDummyFile("model1")
	if err != nil {
		t.Fatal(err)
	}

	model2ID, err := user.trainNdbDummyFile("model2")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "model1-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(model1ID)}, apiKeyName, expiry)
	if err != nil {
		t.Fatalf("Failed to create API key for model1: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(model1ID)
	if err != nil {
		t.Fatalf("API key should have access to model1, but got error: %v", err)
	}

	_, err = user.modelPermissions(model2ID)
	if err == nil {
		t.Fatal("Expected an error when accessing model2 with model1's API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized when accessing model2, got: %v", err)
	}

	_, err = user.modelInfo(model2ID)
	if err == nil {
		t.Fatal("Expected an error when accessing model2's info with model1's API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Expected ErrUnauthorized when accessing model2's info, got: %v", err)
	}
}
