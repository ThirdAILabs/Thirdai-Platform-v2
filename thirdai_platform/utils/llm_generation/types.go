package llm_generation

type Reference struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

type GenerateRequest struct {
	Query      string      `json:"query"`
	TaskPrompt string      `json:"task_prompt"`
	References []Reference `json:"references,omitempty"`
	Model      string      `json:"model"`
}