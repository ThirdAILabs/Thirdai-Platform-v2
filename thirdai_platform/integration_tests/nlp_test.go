package integrationtests

import (
	"testing"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestNlpTokenSupervised(t *testing.T) {
	c := getClient(t)

	model, err := c.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &model.ModelClient, false)

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextSupervised(t *testing.T) {
	c := getClient(t)

	model, err := c.TrainNlpText(
		randomName("nlp-text"),
		3,
		[]client.FileInfo{{Path: "./data/supervised.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &model.ModelClient, false)

	_, err = model.Predict("what is the answer to my question", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpDocSupervised(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpDoc(
		randomName("nlp-doc"),
		"./data/doc_classification_data",
		config.NlpTrainOptions{Epochs: 20},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &model.ModelClient, false)

	result, err := model.Predict("The product exceeded my expectations!", 2)
	if err != nil {
		t.Fatal(err)
	}

	if result.PredictedClasses[0].Class != "positive" {
		t.Fatalf("invalid predicted class %v", result.PredictedClasses)
	}
}

func TestNlpTokenDatagen(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpTokenDatagen(
		randomName("nlp-token"),
		"i want to detect entities in text",
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

	deployModel(t, &model.ModelClient, false)

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextDatagen(t *testing.T) {
	client := getClient(t)

	model, err := client.TrainNlpTextDatagen(
		randomName("nlp-text"),
		"i want to determine if text has a postive or negative sentiment",
		config.NlpTextDatagenOptions{
			Labels: []config.LabelEntity{
				{Name: "POSTIVE", Examples: []string{"i love apples"}, Description: "positive words or phrases"},
				{Name: "NEGATIVE", Examples: []string{"i hate everything"}, Description: "negative words or phrases"},
			},
			SamplesPerlabel: 20,
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

	deployModel(t, &model.ModelClient, false)

	_, err = model.Predict("i really like to eat apples", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTokenRetrain(t *testing.T) {
	c := getClient(t)

	model, err := c.TrainNlpToken(
		randomName("nlp-token"),
		[]string{"EMAIL", "NAME"},
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &model.ModelClient, false)

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AddSample(
		[]string{"jane", "and", "rick", "are", "friends"},
		[]string{"NAME", "O", "NAME", "O", "O"},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AddLabel("PHONE", "a phone number", []string{"123-459-1049"})
	if err != nil {
		t.Fatal(err)
	}

	err = model.AddSample(
		[]string{"call me at 309-248-1094"},
		[]string{"O", "O", "O", "PHONE"},
	)
	if err != nil {
		t.Fatal(err)
	}

	newModel, err := model.Retrain(randomName("nlp-token-retrain"))
	if err != nil {
		t.Fatal(err)
	}

	err = newModel.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTokenTrainingFromBaseModel(t *testing.T) {
	c := getClient(t)

	nlp, err := c.TrainNlpToken(
		randomName("nlp"),
		[]string{"EMAIL", "NAME"},
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nlp2, err := c.TrainNlpTokenWithBaseModel(
		randomName("nlp2"),
		nlp,
		[]client.FileInfo{{Path: "./data/ner.csv", Location: "upload"}},
		config.NlpTrainOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextTrainingFromBaseModel(t *testing.T) {
	c := getClient(t)

	nlp, err := c.TrainNlpText(
		randomName("nlp"),
		3,
		[]client.FileInfo{{Path: "./data/supervised.csv", Location: "upload"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nlp2, err := c.TrainNlpTextWithBaseModel(
		randomName("nlp2"),
		nlp,
		[]client.FileInfo{{Path: "./data/supervised.csv", Location: "upload"}},
		config.NlpTrainOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
}
