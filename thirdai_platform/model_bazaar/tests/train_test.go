package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/google/uuid"
)

func TestTrain(t *testing.T) {
	env := setupPrivateDBEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	model, err := client.trainNdbDummyFile("xyz")
	if err != nil {
		t.Fatal(err)
	}

	jobToken := getJobAuthToken(env, t, model)

	status, err := client.trainStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "starting" || len(status.Errors) != 0 || len(status.Warnings) != 0 {
		t.Fatalf("invalid status: %v", status)
	}

	err = client.Post("/train/log").Auth(jobToken).Json(map[string]string{"level": "warning", "message": "probably fine"}).Do(nil)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Post("/train/log").Auth(jobToken).Json(map[string]string{"level": "error", "message": "uh oh"}).Do(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = updateTrainStatus(client, jobToken, "in_progress")
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

	go env.modelBazaar.JobStatusSync(100 * time.Millisecond)
	time.Sleep(300 * time.Millisecond) // Ensure status sync runs
	env.modelBazaar.StopJobStatusSync()
	time.Sleep(300 * time.Millisecond) // Ensure status sync stops

	status, err = client.trainStatus(model)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "failed" || len(status.Errors) != 1 || status.Errors[0] != "uh oh" || len(status.Warnings) != 1 || status.Warnings[0] != "probably fine" {
		t.Fatalf("invalid status: %v", status)
	}
}

func createUploadBody(t *testing.T, files []struct{ name, data string }) (io.Reader, string) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for _, file := range files {
		part, err := writer.CreateFormFile("files", file.name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := part.Write([]byte(file.data)); err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	return body, writer.FormDataContentType()
}

func TestFileUpload(t *testing.T) {
	env := setupPrivateDBEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	files := []struct {
		name, data string
	}{
		{"a.pdf", "this is some random content"},
		{"b.docx", "more random content"},
		{"c.csv", "a,b\n1,2\n3,4"},
	}

	checkFiles := func(dir string) {
		for _, file := range files {
			obj, err := os.Open(filepath.Join(dir, file.name))
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

	{ // Basic upload
		body, contentType := createUploadBody(t, files)
		var res map[string]string
		err := client.Post("/train/upload-data").Header("Content-Type", contentType).Body(body).Do(&res)
		if err != nil {
			t.Fatal(err)
		}
		checkFiles(filepath.Join(env.storage.Location(), "uploads", res["upload_id"]))
	}

	{ // Sub dir
		body, contentType := createUploadBody(t, files)
		var res map[string]string
		err := client.Post("/train/upload-data?sub_dir=abc").Header("Content-Type", contentType).Body(body).Do(&res)
		if err != nil {
			t.Fatal(err)
		}
		checkFiles(filepath.Join(env.storage.Location(), "uploads", res["upload_id"], "abc"))
	}

	{ // Invalid sub dir name
		body, contentType := createUploadBody(t, files)
		var res map[string]string
		err := client.Post("/train/upload-data?sub_dir=a/b").Header("Content-Type", contentType).Body(body).Do(&res)
		if err == nil || !strings.Contains(err.Error(), "status 422") {
			t.Fatal(err)
		}
	}
}

func TestUseUploadInTrain(t *testing.T) {
	env := setupPrivateDBEnv(t)

	user1, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := env.newUser("xyz")
	if err != nil {
		t.Fatal(err)
	}

	_, err = user1.trainNdb("bad-upload", config.TrainFile{Location: "upload", Path: uuid.NewString()})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatal("upload should not exist")
	}

	body, contentType := createUploadBody(t, []struct{ name, data string }{{"a.pdf", "some random text"}})
	var uploadRes map[string]string
	err = user1.Post("/train/upload-data").Header("Content-Type", contentType).Body(body).Do(&uploadRes)
	if err != nil {
		t.Fatal(err)
	}

	_, err = user2.trainNdb("bad-upload", config.TrainFile{Location: "upload", Path: uploadRes["upload_id"]})
	if err == nil || !strings.Contains(err.Error(), "does not have permission to access upload") {
		t.Fatalf("user cannot access another user's upload: %v", err)
	}

	_, err = user1.trainNdb("bad-upload", config.TrainFile{Location: "upload", Path: uploadRes["upload_id"]})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTrainReport(t *testing.T) {
	env := setupPrivateDBEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	model, err := client.trainNdbDummyFile("xyz")
	if err != nil {
		t.Fatal(err)
	}

	jobToken := getJobAuthToken(env, t, model)

	err = updateTrainStatus(client, jobToken, "complete")
	if err != nil {
		t.Fatal(err)
	}

	err = env.storage.Write(filepath.Join(storage.ModelPath(uuid.MustParse(model)), "train_reports", "1.json"), strings.NewReader(`"the first report"`))
	if err != nil {
		t.Fatal(err)
	}

	report, err := client.trainReport(model)
	if err != nil {
		t.Fatal(err)
	}
	if report.(string) != "the first report" {
		t.Fatal("invalid report data")
	}

	err = env.storage.Write(filepath.Join(storage.ModelPath(uuid.MustParse(model)), "train_reports", "1.json"), strings.NewReader(`"the second report"`))
	if err != nil {
		t.Fatal(err)
	}

	report, err = client.trainReport(model)
	if err != nil {
		t.Fatal(err)
	}
	if report.(string) != "the second report" {
		t.Fatal("invalid report data")
	}
}
