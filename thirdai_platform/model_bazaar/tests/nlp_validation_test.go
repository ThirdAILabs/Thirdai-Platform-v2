package tests

import (
	"slices"
	"strings"
	"testing"
	"thirdai_platform/model_bazaar/services"

	"github.com/google/uuid"
)

func TestTrainableCsvValidation(t *testing.T) {
	env := setupTestEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	uploadFunc := func(fileContent []struct{ name, data string }) string {
		body, contentType := createUploadBody(t, fileContent)
		var res map[string]uuid.UUID
		err := client.Post("/train/upload-data").Header("Content-Type", contentType).Body(body).Do(&res)
		if err != nil {
			t.Fatal(err)
		}
		return res["upload_id"].String()
	}

	{ // Non-csv file
		non_csv_file := []struct {
			name, data string
		}{
			{"a.pdf", "this is some random content"},
		}

		uploadID := uploadFunc(non_csv_file)
		var response map[string][]string

		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "text"}).Do(&response)
		if err == nil {
			t.Fatal("Only CSV file should've been supported")
		}
		if !strings.Contains(err.Error(), "only CSV file is supported") {
			t.Fatal(err)
		}
	}

	{
		//text-csv file with invalid headers
		textFile1 := []struct {
			name, data string
		}{
			{"textFile1.csv", "text,target\nNormal text,label1\nDifferent text,label2"},
		}
		var response map[string][]string

		uploadID := uploadFunc(textFile1)
		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "text"}).Do(&response)
		if err == nil {
			t.Fatal("text-csv file with invalid headers shouldn't get validated")
		}
		if !strings.Contains(err.Error(), "invalid column") {
			t.Fatal(err)
		}
	}

	{
		//Text-csv file with wrong number of fields
		textFile2 := []struct {
			name, data string
		}{
			{"textFile2.csv", "text,labels\nNormal text,label1\nDifferent text,label2,extra-entry"},
		}
		var response map[string][]string

		uploadID := uploadFunc(textFile2)
		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "text"}).Do(&response)
		if err == nil {
			t.Fatal("text-csv file with wrong number of fields shouldn't get validated")
		}
		if !strings.Contains(err.Error(), "wrong number of fields") {
			t.Fatal(err)
		}
	}

	{
		//Token-csv file with incorrect source and target length
		tokenFile2 := []struct {
			name, data string
		}{
			{"tokenFile2.csv", "source,target\nTexas is the address,LOCATION O O O\nHe saw Dr Liam yesterday,O O O NAME O O"},
		}
		var response map[string][]string

		uploadID := uploadFunc(tokenFile2)
		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "token"}).Do(&response)
		if err == nil {
			t.Fatal("token-csv file with with incorrect source and target length shouldn't get validated")
		}
		if !strings.Contains(err.Error(), "number of source tokens: 5 ≠ number of target tokens: 6") {
			t.Fatal(err)
		}
	}

	{
		//Correct text-csv file
		correctTextFile := []struct {
			name, data string
		}{
			{"correctTextFile.csv", "text,labels\nNormal text,label1\nDifferent text,label2"},
		}
		var response map[string][]string

		uploadID := uploadFunc(correctTextFile)
		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "text"}).Do(&response)
		if err != nil {
			t.Fatal(err)
		}

		labels := response["labels"]
		if !slices.Contains(labels, "label1") || !slices.Contains(labels, "label2") {
			t.Fatalf("Invalid labels: %v parsed", labels)
		}
	}

	{
		//Correct text-csv file
		correctTokenFile := []struct {
			name, data string
		}{
			{"correctTokenFile.csv", "source,target\nTexas is the address,LOCATION O O O\nHe saw Dr Liam yesterday,O O O NAME O"},
		}
		var response map[string][]string

		uploadID := uploadFunc(correctTokenFile)
		err := client.Post("/train/validate-trainable-csv").Json(services.TrainableCSVRequest{UploadId: uploadID, FileType: "token"}).Do(&response)
		if err != nil {
			t.Fatal(err)
		}

		labels := response["labels"]
		if !slices.Contains(labels, "NAME") || !slices.Contains(labels, "LOCATION") {
			t.Fatalf("Invalid labels: %v parsed", labels)
		}
	}
}
