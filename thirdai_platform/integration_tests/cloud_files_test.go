package integrationtests

import (
	"net/http"
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func verifyDownload(t *testing.T, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got status %d from url '%s'", res.StatusCode, url)
	}
}

func verifyNdbWorksWithFile(t *testing.T, url, provider, query string) {
	c := getClient(t)

	ndb, err := c.TrainNdb(
		randomName(provider), []client.FileInfo{{Path: url, Location: provider}}, nil, config.JobOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &ndb.ModelClient, false)

	res, err := ndb.Search(query, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) < 1 {
		t.Fatal("missing results for query")
	}

	sources, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.DeleteDocs([]string{sources[0].SourceId})
	if err != nil {
		t.Fatal(err)
	}

	sourcesAfter, err := ndb.Sources()
	if err != nil {
		t.Fatal(err)
	}
	if len(sourcesAfter)+1 != len(sources) {
		t.Fatal("incorrect sources after deletion")
	}

	signedUrl, err := ndb.GetSignedUrl(res[0].Source, provider)
	if err != nil {
		t.Fatal(err)
	}

	verifyDownload(t, signedUrl)
}

func TestS3Public(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"s3://thirdai-corp-public/ThirdAI-Enterprise-Test-Data/scifact/",
		"s3",
		"sample query",
	)
}

func TestS3Private(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"s3://thirdai-datasets/insert.pdf",
		"s3",
		"Alice in wonderland",
	)
}

func TestAzurePublic(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"https://csg100320028d93f3bc.blob.core.windows.net/test/insert.pdf",
		"azure",
		"Alice in wonderland",
	)
}

func TestAzurePrivate(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"https://csg100320028d93f3bc.blob.core.windows.net/private-platform/test_folder/",
		"azure",
		"Alice in wonderland",
	)
}

func TestGcpPublic(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"gs://public-training-platform/sample_nda.pdf",
		"gcp",
		"confidentiality agreement",
	)
}

func TestGcpPrivate(t *testing.T) {
	verifyNdbWorksWithFile(
		t,
		"gs://private-thirdai-platform/test_folder/",
		"gcp",
		"confidentiality agreement",
	)
}
