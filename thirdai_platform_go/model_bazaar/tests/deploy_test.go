package tests

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestDeploy(t *testing.T) {
	env := setupTestEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	model, err := client.trainNdb("xyz")
	if err != nil {
		t.Fatal(err)
	}

	trainConfig, err := env.storage.Read(filepath.Join("models", model, "train_config.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer trainConfig.Close()

	var params map[string]interface{}
	err = json.NewDecoder(trainConfig).Decode(&params)
	if err != nil {
		t.Fatal(err)
	}

	token := params["job_auth_token"].(string)

	_, err = postWithToken[NoBody](&client, "/train/update-status", []byte(`{"status": "complete"}`), token)
	if err != nil {
		t.Fatal(err)
	}

	err = client.deploy(model)
	if err != nil {
		t.Fatal(err)
	}

	status, err := client.deployStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "starting" || len(status.Errors) != 0 || len(status.Warnings) != 0 {
		t.Fatalf("invalid status: %v", status)
	}

	_, err = postWithToken[NoBody](&client, "/deploy/log", []byte(`{"level": "warning", "message": "probably fine"}`), token)
	if err != nil {
		t.Fatal(err)
	}
	_, err = postWithToken[NoBody](&client, "/deploy/log", []byte(`{"level": "error", "message": "uh oh"}`), token)
	if err != nil {
		t.Fatal(err)
	}

	_, err = postWithToken[NoBody](&client, "/deploy/update-status", []byte(`{"status": "in_progress"}`), token)
	if err != nil {
		t.Fatal(err)
	}

	status, err = client.deployStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "in_progress" || len(status.Errors) != 1 || status.Errors[0] != "uh oh" || len(status.Warnings) != 1 || status.Warnings[0] != "probably fine" {
		t.Fatalf("invalid status: %v", status)
	}

	env.nomad.Clear() // Make it look like the job stopped

	go env.modelBazaar.StartStatusSync(100 * time.Millisecond)
	time.Sleep(300 * time.Millisecond) // Ensure status sync runs
	env.modelBazaar.StopStatusSync()
	time.Sleep(300 * time.Millisecond) // Ensure status sync stops

	status, err = client.deployStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "failed" || len(status.Errors) != 1 || status.Errors[0] != "uh oh" || len(status.Warnings) != 1 || status.Warnings[0] != "probably fine" {
		t.Fatalf("invalid status: %v", status)
	}

	err = client.undeploy(model)
	status, err = client.deployStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "stopped" || len(status.Errors) != 1 || status.Errors[0] != "uh oh" || len(status.Warnings) != 1 || status.Warnings[0] != "probably fine" {
		t.Fatalf("invalid status: %v", status)
	}
}
