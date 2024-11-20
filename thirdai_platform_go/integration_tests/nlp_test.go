package integrationtests

import (
	"testing"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestNlpToken(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]config.FileInfo{{Path: "./data/ner.csv", Location: "local"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = model.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := model.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpText(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpText(
		randomName("nlp-text"),
		3,
		[]config.FileInfo{{Path: "./data/supervised.csv", Location: "local"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = model.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := model.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.Predict("what is the answer to my question", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTokenDatagen(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpTokenDatagen(
		randomName("nlp-token"),
		"i want to detect a person's name and email in text",
		config.NlpTokenDatagenOptions{
			Tags: []config.LabelEntity{
				{Name: "EMAIL", Examples: []string{"my email is bob@gmail.com"}, Description: "a person's email"},
				{Name: "NAME", Examples: []string{"my name is bob"}, Description: "a person's name"},
			},
			NumSentencesToGenerate: 40,
		},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = model.Deploy(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := model.Undeploy()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}
