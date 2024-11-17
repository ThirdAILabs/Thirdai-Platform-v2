package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTrain(t *testing.T) {
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

	status, err := client.trainStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "starting" || len(status.Errors) != 0 || len(status.Warnings) != 0 {
		t.Fatalf("invalid status: %v", status)
	}

	_, err = postWithToken[NoBody](&client, "/train/log", []byte(`{"level": "warning", "message": "probably fine"}`), token)
	if err != nil {
		t.Fatal(err)
	}
	_, err = postWithToken[NoBody](&client, "/train/log", []byte(`{"level": "error", "message": "uh oh"}`), token)
	if err != nil {
		t.Fatal(err)
	}

	_, err = postWithToken[NoBody](&client, "/train/update-status", []byte(`{"status": "in_progress"}`), token)
	if err != nil {
		t.Fatal(err)
	}

	status, err = client.trainStatus(model)
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

	status, err = client.trainStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "failed" || len(status.Errors) != 1 || status.Errors[0] != "uh oh" || len(status.Warnings) != 1 || status.Warnings[0] != "probably fine" {
		t.Fatalf("invalid status: %v", status)
	}
}

func TestFileUpload(t *testing.T) {
	env := setupTestEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	files := []struct {
		name, data string
	}{
		{"a.pdf", "this is some random content"},
		{"b.docx", "more random content"},
		{"c.csv", "a,b\n1,2\n3,4"},
	}

	for _, file := range files {
		part, err := writer.CreateFormFile("files", file.name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = part.Write([]byte(file.data))
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	res, err := postWithHeaders[map[string]string](
		&client,
		"/train/upload-data",
		body.Bytes(),
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", client.token),
			"Content-Type":  writer.FormDataContentType(),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	artifactPath := res["artifact_path"]

	for _, file := range files {
		obj, err := os.Open(filepath.Join(artifactPath, file.name))
		if err != nil {
			t.Fatal(err)
		}
		defer obj.Close()

		data, err := io.ReadAll(obj)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, []byte(file.data)) {
			t.Fatal("invalid file contents")
		}
	}

}
