package config

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

	Tags                     []LabelEntity `json:"tags"`
	NumSentencesToGeneration int           `json:"num_sentences_to_generate"`
	NumSamplesPerTag         *int          `json:"num_samples_per_tag"`

	Samples            []NlpTokenSample `json:"samples"`
	TemplatesPerSample int              `json:"templates_per_sample"`
}

type NlpTextSample struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

type NlpTextDatagenOptions struct {
	ModelType string `json:"model_type"`

	SamplesPerlabel  int           `json:"samples_per_label"`
	Labels           []LabelEntity `json:"labels"`
	UserVocab        []string      `json:"user_vocab"`
	UserPrompts      []string      `json:"user_prompts"`
	VocabPerSentence int           `json:"vocab_per_sentence"`
}

type DatagenConfig struct {
	TaskPrompt  string `json:"task_prompt"`
	LlmProvider string `json:"llm_provider"`

	DatagenOptions interface{} `json:"datagen_options"`
}