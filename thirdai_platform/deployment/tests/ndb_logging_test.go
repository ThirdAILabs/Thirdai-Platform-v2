package tests

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"

	"github.com/google/uuid"
)

func getLastLogLine(t *testing.T, logPath string) map[string]interface{} {
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading log file: %v", err)
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lastLine), &logEntry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}

	return logEntry
}

func validateLogEntry(t *testing.T, logEntry map[string]interface{}, expectedMsg string, expectedCode string, userId string, modelId string) {
	// Check default fields
	if logEntry["model_type"] != "ndb" {
		t.Errorf("expected model_type 'model_type', got %v", logEntry["model_type"])
	}
	if logEntry["service_type"] != "deployment" {
		t.Errorf("expected service_type 'deployment', got %v", logEntry["service_type"])
	}
	if logEntry["_msg"] != expectedMsg {
		t.Errorf("expected message '%s', got %v", expectedMsg, logEntry["_msg"])
	}
	if logEntry["code"] != expectedCode {
		t.Errorf("expected code '%s', got %v", expectedCode, logEntry["code"])
	}
	if _, exists := logEntry["_time"]; !exists {
		t.Errorf("expected _time field to exist")
	}
	// fields specific to deployment type
	if logEntry["user_id"] != userId {
		t.Errorf("expected user_id '%s', got %v", userId, logEntry["user_id"])
	}
	if logEntry["model_id"] != modelId {
		t.Errorf("expected model_id '%s', got %v", modelId, logEntry["model_id"])
	}
}

func TestLogging(t *testing.T) {
	config := &config.DeployConfig{
		ModelId:        uuid.New(),
		UserId:         uuid.New(),
		ModelType:      "ndb",
		ModelBazaarDir: t.TempDir(),
	}
	modelID := config.ModelId.String()
	userID := config.UserId.String()

	logPath := filepath.Join(config.ModelBazaarDir, "deployment.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	deployment.InitLogging(logFile, config)
	testServer, _ := makeNdbServer(t, config)
	defer testServer.Close()

	// Test insert logging
	doInsert(t, testServer)
	logEntry := getLastLogLine(t, logPath)
	validateLogEntry(t, logEntry, "inserted document", "MODEL_INSERT", userID, modelID)

	// Test associate logging
	doAssociate(t, testServer, "source", "test line")
	logEntry = getLastLogLine(t, logPath)
	validateLogEntry(t, logEntry, "associated text pairs", "MODEL_RLHF", userID, modelID)

	// Test upvote logging
	doUpvote(t, testServer, "unrelated_query", 2)
	logEntry = getLastLogLine(t, logPath)
	validateLogEntry(t, logEntry, "upvoted document", "MODEL_RLHF", userID, modelID)

	// Test delete logging
	doDelete(t, testServer, []string{"doc_id_1"})
	logEntry = getLastLogLine(t, logPath)
	validateLogEntry(t, logEntry, "deleted documents", "MODEL_DELETE", userID, modelID)
}
