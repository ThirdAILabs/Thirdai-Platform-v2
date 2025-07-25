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

	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, keyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}
	if apiKeyVal == "" {
		t.Fatal("expected a valid API key, but got an empty string")
	}

	userKeys, err := user.ListAPIKeys()
	if err != nil {
		t.Fatalf("failed to list API keys: %v", err)
	}

	var createdKey *services.APIKeyResponse
	for _, key := range userKeys {
		if key.Name == keyName {
			createdKey = &key
			break
		}
	}
	if createdKey == nil {
		t.Fatalf("expected to find API key '%s' in the list, but did not", keyName)
	}
	if createdKey.Expiry.IsZero() || createdKey.Expiry.Before(time.Now()) {
		t.Fatalf("invalid expiry time on created key: %v", createdKey.Expiry)
	}

	err = user.DeleteAPIKey(createdKey.ID)
	if err != nil {
		t.Fatalf("failed to delete API key: %v", err)
	}

	userKeys, err = user.ListAPIKeys()
	if err != nil {
		t.Fatalf("failed to list API keys after deletion: %v", err)
	}
	for _, key := range userKeys {
		if key.Name == keyName {
			t.Fatalf("api key '%s' was not deleted properly", keyName)
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
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	checkPermissions(user, t, modelID, true, true, true)
	if err != nil {
		t.Fatalf("test failed: %v", err)
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
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	info, err := user.modelInfo(modelID)
	if err != nil {
		t.Fatalf("failed to access model info with API key: %v", err)
	}

	if info.ModelId.String() != modelID {
		t.Fatalf("expected model_id %s, got %s", modelID, info.ModelId.String())
	}
	if info.ModelName != "key-info-model" {
		t.Fatalf("expected model name 'key-info-model', got %s", info.ModelName)
	}

	_, err = user.trainNdbDummyFile("other-model")
	if err == nil {
		t.Fatalf("cannot train a model using API-KEY")
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
	_, err = user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry, false)
	if err == nil {
		t.Fatal("expected error when creating an already expired API key, but got success.")
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
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err != nil {
		t.Fatalf("api key should have worked, but got error: %v", err)
	}

	err = user.login(loginInfo{Email: "delete-user@mail.com", Password: "delete-user_password"})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	userKeys, err := user.ListAPIKeys()
	if err != nil {
		t.Fatalf("failed to list API keys: %v", err)
	}

	var keyToDelete uuid.UUID
	for _, k := range userKeys {
		if k.Name == apiKeyName {
			keyToDelete = k.ID
			break
		}
	}
	if keyToDelete == uuid.Nil {
		t.Fatalf("created key '%s' not found in list", apiKeyName)
	}

	err = user.DeleteAPIKey(keyToDelete)
	if err != nil {
		t.Fatalf("failed to delete API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err == nil {
		t.Fatal("expected an error when using a deleted API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
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
		t.Fatalf("failed to train NDB model: %v", err)
	}

	nlpID, err := user.trainNlpToken("my-nlp-dep")
	if err != nil {
		t.Fatalf("failed to train NLP model: %v", err)
	}

	err = user.updateAccess(nlpID, schema.Public, nil)
	if err != nil {
		t.Fatal(err)
	}

	esID, err := user.createEnterpriseSearch("my-es-model", ndbID, nlpID)
	if err != nil {
		t.Fatalf("failed to create Enterprise Search model: %v", err)
	}

	apiKeyName := "dep-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(esID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	checkPermissions(user, t, ndbID, true, true, true)
	checkPermissions(user, t, nlpID, true, true, true)
	checkPermissions(user, t, esID, true, true, true)
}

func TestAPIKeyUsageAfterExpiration(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("api-exp-user")
	if err != nil {
		t.Fatal(err)
	}

	modelID, err := user.trainNdbDummyFile("expire-key-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "temp-key"
	expiry := time.Now().Add(2 * time.Second)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(modelID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(modelID)
	if err != nil {
		t.Fatalf("api key should have worked before expiration, but got error: %v", err)
	}

	time.Sleep(3 * time.Second)

	_, err = user.modelPermissions(modelID)
	if err == nil {
		t.Fatal("expected an error when using an expired API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized after expiration, got: %v", err)
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
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{uuid.MustParse(model1ID)}, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("failed to create API key for model1: %v", err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	_, err = user.modelPermissions(model1ID)
	if err != nil {
		t.Fatalf("api key should have access to model1, but got error: %v", err)
	}

	_, err = user.modelPermissions(model2ID)
	if err == nil {
		t.Fatal("expected an error when accessing model2 with model1's API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized when accessing model2, got: %v", err)
	}

	_, err = user.modelInfo(model2ID)
	if err == nil {
		t.Fatal("expected an error when accessing model2's info with model1's API key, but got success.")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized when accessing model2's info, got: %v", err)
	}
}

func TestAPIKeyAccessAllModels(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.newUser("all-models-user")
	if err != nil {
		t.Fatal(err)
	}

	initialModelID, err := user.trainNdbDummyFile("initial-model")
	if err != nil {
		t.Fatal(err)
	}

	apiKeyName := "all-models-key"
	expiry := time.Now().Add(24 * time.Hour)
	apiKeyVal, err := user.createAPIKey([]uuid.UUID{}, apiKeyName, expiry, true)
	if err != nil {
		t.Fatalf("failed to create API key with AllModels access: %v", err)
	}
	if apiKeyVal == "" {
		t.Fatal("expected a valid API key, but got an empty string")
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	initialInfo, err := user.modelInfo(initialModelID)
	if err != nil {
		t.Fatalf("failed to access initial model with API key: %v", err)
	}
	if initialInfo.ModelId.String() != initialModelID {
		t.Fatalf("expected model_id %s, got %s", initialModelID, initialInfo.ModelId.String())
	}

	err = user.login(loginInfo{Email: "all-models-user@mail.com", Password: "all-models-user_password"})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	newModelID, err := user.trainNdbDummyFile("new-model-after-key")
	if err != nil {
		t.Fatal(err)
	}

	err = user.UseApiKey(apiKeyVal)
	if err != nil {
		t.Fatalf("failed to set API key: %v", err)
	}

	newInfo, err := user.modelInfo(newModelID)
	if err != nil {
		t.Fatalf("failed to access new model with API key: %v", err)
	}
	if newInfo.ModelId.String() != newModelID {
		t.Fatalf("expected model_id %s, got %s", newModelID, newInfo.ModelId.String())
	}

	_, err = user.modelPermissions(initialModelID)
	if err != nil {
		t.Fatalf("API key should have access to initial model, but got error: %v", err)
	}

	_, err = user.modelPermissions(newModelID)
	if err != nil {
		t.Fatalf("API key should have access to new model, but got error: %v", err)
	}
}
