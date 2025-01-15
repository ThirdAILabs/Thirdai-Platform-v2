package integrationtests

import (
	"path/filepath"
	"strings"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func downloadAndUpload(t *testing.T, c *client.PlatformClient, model *client.ModelClient) interface{} {
	downloadFile := filepath.Join(t.TempDir(), "download.zip")

	if err := model.Download(downloadFile); err != nil {
		t.Fatal(err)
	}

	uploadModel, err := c.UploadModel(randomName("upload-ndb"), downloadFile)
	if err != nil {
		t.Fatal(err)
	}

	return uploadModel
}

func TestNdbDownloadUpload(t *testing.T) {
	c := getClient(t)

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

	if err := ndb.AwaitTrain(100 * time.Second); err != nil {
		t.Fatal(err)
	}

	uploadNdb := downloadAndUpload(t, c, &ndb.ModelClient).(*client.NdbClient)

	deployModel(t, &uploadNdb.ModelClient, false)

	checkQuery(uploadNdb, t, true)

	checkNoUpvote(uploadNdb, t)
	doUpvote(uploadNdb, t)
	checkUpvote(uploadNdb, t)

	checkNoAssociate(uploadNdb, t)
	doAssociate(uploadNdb, t)
	checkAssociate(uploadNdb, t)

	checkSources(uploadNdb, t, []string{"articles.csv"})
	doInsert(uploadNdb, t, []string{"four_english_words.docx", "mutual_nda.pdf", "articles.csv"})
	checkSources(uploadNdb, t, []string{"articles.csv", "articles.csv", "four_english_words.docx", "mutual_nda.pdf"})
}

func TestNlpTokenDownloadUpload(t *testing.T) {
	c := getClient(t)

	nlp, err := c.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := nlp.AwaitTrain(100 * time.Second); err != nil {
		t.Fatal(err)
	}

	uploadNlp := downloadAndUpload(t, c, &nlp.ModelClient).(*client.NlpTokenClient)

	deployModel(t, &uploadNlp.ModelClient, false)

	_, err = uploadNlp.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextDownloadUpload(t *testing.T) {
	c := getClient(t)

	nlp, err := c.TrainNlpText(
		randomName("nlp-text"),
		3,
		[]client.FileInfo{{Path: "./data/supervised.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := nlp.AwaitTrain(100 * time.Second); err != nil {
		t.Fatal(err)
	}

	uploadNlp := downloadAndUpload(t, c, &nlp.ModelClient).(*client.NlpTextClient)

	deployModel(t, &uploadNlp.ModelClient, false)

	_, err = uploadNlp.Predict("what is the answer to my question", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpDocDownloadUpload(t *testing.T) {
	c := getClient(t)

	nlp, err := c.TrainNlpDoc(
		randomName("nlp-doc"),
		"./data/doc_classification_data",
		config.NlpTrainOptions{Epochs: 20},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := nlp.AwaitTrain(100 * time.Second); err != nil {
		t.Fatal(err)
	}

	uploadNlp := downloadAndUpload(t, c, &nlp.ModelClient).(*client.NlpTextClient)

	deployModel(t, &uploadNlp.ModelClient, false)

	result, err := uploadNlp.Predict("The product exceeded my expectations!", 2)
	if err != nil {
		t.Fatal(err)
	}

	if result.PredictedClasses[0].Class != "positive" {
		t.Fatalf("invalid predicted class %v", result.PredictedClasses)
	}
}

func TestKnowledgeExtractionUploadDownload(t *testing.T) {
	c := getClient(t)

	ke, err := c.CreateKnowledgeExtractionWorkflow(
		randomName("ke"),
		[]string{"what are arsenal and manchester united", "manufacturing faster chips"},
		"",
		false, // no generation for test speed
	)
	if err != nil {
		t.Fatal(err)
	}

	uploadKe := downloadAndUpload(t, c, &ke.ModelClient).(*client.KnowledgeExtractionClient)

	deployModel(t, &uploadKe.ModelClient, false)

	reportId, err := uploadKe.CreateReport([]client.FileInfo{
		{Path: "./data/articles.csv", Location: "upload"},
	})
	if err != nil {
		t.Fatal(err)
	}

	report, err := uploadKe.AwaitReport(reportId, 100*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Content.Results) != 2 {
		t.Fatal("incorrect results returned")
	}

	expectedAnswerText := []string{
		"England boss Sven Goran Eriksson", "New Technology From AMD And IBM",
	}

	for i, result := range report.Content.Results {
		if len(result.References) < 1 || !strings.Contains(result.References[0].Text, expectedAnswerText[i]) {
			t.Fatal("incorrect result for question")
		}
		if result.Answer != "" {
			t.Fatal("no generation should be done")
		}
	}
}
