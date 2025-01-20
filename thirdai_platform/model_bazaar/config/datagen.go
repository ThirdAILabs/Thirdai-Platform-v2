package config

import (
	"fmt"
	"thirdai_platform/model_bazaar/schema"

	"github.com/google/uuid"
)

type LabelEntity struct {
	Name        string   `json:"name"`
	Examples    []string `json:"examples"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
}

type NlpTokenSample struct {
	Tokens []string `json:"tokens"`
	Tags   []string `json:"tags"`
}

type NlpTokenDatagenOptions struct {
	ModelType string `json:"model_type"`

	Tags                   []LabelEntity `json:"tags"`
	NumSentencesToGenerate int           `json:"num_sentences_to_generate"`
	NumSamplesPerTag       *int          `json:"num_samples_per_tag"`

	Samples            []NlpTokenSample `json:"samples"`
	TemplatesPerSample int              `json:"templates_per_sample"`

	LoadFromStorage bool `json:"load_from_storage"`
}

func (opts *NlpTokenDatagenOptions) Validate() error {
	opts.ModelType = schema.NlpTokenModel

	if opts.Tags == nil {
		return fmt.Errorf("'tags' must be specified in token datagen options")
	}

	for i, tag := range opts.Tags {
		if tag.Name == "" {
			return fmt.Errorf("each tag must have a specified name")
		}
		if tag.Status == "" {
			opts.Tags[i].Status = "uninserted"
		}
	}

	if opts.NumSentencesToGenerate == 0 {
		opts.NumSentencesToGenerate = 1000
	}

	if opts.TemplatesPerSample == 0 {
		opts.TemplatesPerSample = 10
	}

	return nil
}

func (opts *NlpTokenDatagenOptions) GetModelOptions() interface{} {
	tags := make([]string, 0, len(opts.Tags))
	for _, tag := range opts.Tags {
		tags = append(tags, tag.Name)
	}
	return NlpTokenOptions{
		ModelType:    schema.NlpTokenModel,
		TargetLabels: tags,
		SourceColumn: "source",
		TargetColumn: "target",
		DefaultTag:   "O",
	}
}

type NlpTextSample struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

type NlpTextDatagenOptions struct {
	ModelType string `json:"model_type"`

	Labels           []LabelEntity `json:"labels"`
	SamplesPerlabel  int           `json:"samples_per_label"`
	UserVocab        []string      `json:"user_vocab"`
	UserPrompts      []string      `json:"user_prompts"`
	VocabPerSentence int           `json:"vocab_per_sentence"`
}

func (opts *NlpTextDatagenOptions) Validate() error {
	opts.ModelType = schema.NlpTextModel

	if opts.Labels == nil {
		return fmt.Errorf("'labels' must be specified in text datagen options")
	}

	for i, tag := range opts.Labels {
		if tag.Name == "" {
			return fmt.Errorf("each label must have a specified name")
		}
		if tag.Status == "" {
			opts.Labels[i].Status = "uninserted"
		}
	}

	if opts.SamplesPerlabel == 0 {
		return fmt.Errorf("'samples_per_label' must be specified in datagen options")
	}

	if opts.VocabPerSentence == 0 {
		opts.VocabPerSentence = 4
	}

	return nil
}

func (opts *NlpTextDatagenOptions) GetModelOptions() interface{} {
	return NlpTextOptions{
		ModelType:      schema.NlpTextModel,
		NTargetClasses: len(opts.Labels),
		TextColumn:     "text",
		LabelColumn:    "label",
		Delimiter:      ",",
	}
}

type DatagenConfig struct {
	ModelId        uuid.UUID  `json:"model_id"`
	BaseModelId    *uuid.UUID `json:"base_model_id"`
	ModelBazaarDir string     `json:"model_bazaar_dir"`
	StorageDir     string     `json:"storage_dir"`

	ModelBazaarEndpoint string `json:"model_bazaar_endpoint"`

	TaskPrompt  string  `json:"task_prompt"`
	LlmProvider string  `json:"llm_provider"`
	TestSize    float32 `json:"test_size"`

	TaskOptions interface{} `json:"task_options"`
}
