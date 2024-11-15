package integrationtests

import (
	"fmt"
	"os"
	"testing"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestNdbTrainingAndDeployment(t *testing.T) {
	client := getClient(t)

	d, _ := os.Getwd()

	fmt.Println(d)

	ndb, err := client.TrainNdb(
		randomName("ndb"), []config.FileInfo{{
			Path: "./data/articles.csv", Location: "local",
		}},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = ndb.Deploy()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := ndb.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	}()

	err = ndb.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	results, err := ndb.Search("manufacturing faster chips", 4)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) < 1 || results[0].Id != 27 {
		t.Fatal("incorrect results returned")
	}

}
