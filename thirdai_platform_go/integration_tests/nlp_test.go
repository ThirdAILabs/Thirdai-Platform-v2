package integrationtests

import (
	"testing"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestNlpTokenSupervised(t *testing.T) {
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
	t.Log("training started")

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("training complete")

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
	t.Log("deployment started")

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deployment complete")

	_, err = model.Predict("jonas is my name", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextSupervised(t *testing.T) {
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
	t.Log("training started")

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("training complete")

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
	t.Log("deployment started")

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deployment complete")

	_, err = model.Predict("what is the answer to my question", 3)
	if err != nil {
		t.Fatal(err)
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
	t.Log("training started")

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("training complete")

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
	t.Log("deployment started")

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deployment complete")

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
	t.Log("training started")

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("training complete")

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
	t.Log("deployment started")

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deployment complete")

	_, err = model.Predict("i really like to eat apples", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTokenRetrain(t *testing.T) {
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
	t.Log("training started")

	err = model.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("training complete")

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
	t.Log("deployment started")

	err = model.AwaitDeploy(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deployment complete")

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
	client := getClient(t)

	nlp, err := client.TrainNlpToken(
		randomName("nlp"),
		[]string{"EMAIL", "NAME"},
		[]config.FileInfo{{Path: "./data/ner.csv", Location: "local"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nlp2, err := client.TrainNlpTokenWithBaseModel(
		randomName("nlp2"),
		nlp,
		[]config.FileInfo{{Path: "./data/ner.csv", Location: "local"}},
		config.NlpTrainOptions{},
	)

	err = nlp2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNlpTextTrainingFromBaseModel(t *testing.T) {
	client := getClient(t)

	nlp, err := client.TrainNlpText(
		randomName("nlp"),
		3,
		[]config.FileInfo{{Path: "./data/supervised.csv", Location: "local"}},
		config.NlpTrainOptions{Epochs: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = nlp.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	nlp2, err := client.TrainNlpTextWithBaseModel(
		randomName("nlp2"),
		nlp,
		[]config.FileInfo{{Path: "./data/supervised.csv", Location: "local"}},
		config.NlpTrainOptions{},
	)

	err = nlp2.AwaitTrain(100 * time.Second)
	if err != nil {
		t.Fatal(err)
		
	}
}
