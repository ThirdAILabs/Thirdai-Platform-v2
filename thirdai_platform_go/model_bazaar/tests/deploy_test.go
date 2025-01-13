package tests

import (
	"slices"
	"testing"
	"thirdai_platform/model_bazaar/services"
	"time"
)

func TestDeploy(t *testing.T) {
	env := setupTestEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	model, err := client.trainNdbDummyFile("xyz")
	if err != nil {
		t.Fatal(err)
	}

	jobToken := getJobAuthToken(env, t, model)

	err = updatTrainStatus(client, jobToken, "complete")
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

	err = client.Post("/deploy/log").Auth(jobToken).Json(map[string]string{"level": "warning", "message": "probably fine"}).Do(nil)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Post("/deploy/log").Auth(jobToken).Json(map[string]string{"level": "error", "message": "uh oh"}).Do(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Post("/deploy/update-status").Auth(jobToken).Json(map[string]string{"status": "in_progress"}).Do(nil)
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

	var internalStatus services.StatusResponse
	err = client.Get("/deploy/status-internal").Auth(jobToken).Do(&internalStatus)
	if err != nil {
		t.Fatal(err)
	}
	if internalStatus.Status != status.Status || !slices.Equal(internalStatus.Errors, status.Errors) || !slices.Equal(internalStatus.Warnings, status.Warnings) {
		t.Fatalf("invalid internal status: %v", internalStatus)
	}

	env.nomad.Clear() // Make it look like the job stopped

	go env.modelBazaar.JobStatusSync(100 * time.Millisecond)
	time.Sleep(300 * time.Millisecond) // Ensure status sync runs
	env.modelBazaar.StopJobStatusSync()
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
