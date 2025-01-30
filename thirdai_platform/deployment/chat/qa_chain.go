package chat

import (
	"context"
	"fmt"
	"thirdai_platform/search/ndb"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type ndbRetriever struct {
	ndb.NeuralDB
}

func (ndb *ndbRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	chunks, err := ndb.Query(query /*topk=*/, 5 /*constraints=*/, nil)
	if err != nil {
		return nil, fmt.Errorf("ndb query failed: %w", err)
	}

	docs := make([]schema.Document, 0, len(chunks))
	for _, chunk := range chunks {
		docs = append(docs, schema.Document{
			PageContent: chunk.Text,
			Metadata:    nil,
			Score:       chunk.Score,
		})
	}

	return docs, nil
}

var _ schema.Retriever = &ndbRetriever{}

const (
	defaultChatPrompt               = "Answer the user's questions based on the below context:\n\n{{.references}}"
	defaultQueryReformulationPrompt = "Given the above conversation, generate a search query that would help retrieve relevant sources for responding to the last message."
)

const (
	inputKey             = "messages"
	outputKey            = "answer"
	reformulatedQueryKey = "reformulated_query"
	referencesKey        = "references"
)

type ReformulatedRetrievalQA struct {
	reformulatedQuery chains.Chain
	retriever         schema.Retriever
	llmAnswer         *chains.LLMChain
}

func (qa *ReformulatedRetrievalQA) Call(ctx context.Context, inputs map[string]any, options ...chains.ChainCallOption) (map[string]any, error) {
	messages, ok := inputs[inputKey]
	if !ok {
		return nil, fmt.Errorf("%w: %w", chains.ErrInvalidInputValues, chains.ErrMissingInputValues)
	}

	query, err := chains.Predict(ctx, qa.reformulatedQuery, inputs, options...)
	if err != nil {
		return nil, fmt.Errorf("error reformulating query: %w", err)
	}

	docs, err := qa.retriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	result, err := chains.Predict(ctx, qa.llmAnswer, map[string]any{
		inputKey:      messages,
		referencesKey: docs,
	}, options...)
	if err != nil {
		return nil, err
	}

	return map[string]any{outputKey: result}, nil
}

func (qa *ReformulatedRetrievalQA) GetMemory() schema.Memory {
	return memory.NewSimple()
}

func (qa *ReformulatedRetrievalQA) GetInputKeys() []string {
	return []string{inputKey}
}

func (qa *ReformulatedRetrievalQA) GetOutputKeys() []string {
	return []string{outputKey}
}

func NewReformulatedRetrievalQA(ndb ndb.NeuralDB, llm llms.Model) *ReformulatedRetrievalQA {
	queryReformulationPrompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.MessagesPlaceholder{VariableName: inputKey},
		prompts.NewHumanMessagePromptTemplate(defaultQueryReformulationPrompt, nil),
	})

	reformulatedQuery := chains.NewLLMChain(llm, queryReformulationPrompt)
	// Ensure the output of this is passed to the Retriever correctly.
	reformulatedQuery.OutputKey = reformulatedQueryKey

	qaPrompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(defaultChatPrompt, []string{referencesKey}),
		// The RetrieverQA chain passes the results from the retriever to the llm using
		// the key 'input_documents'
		prompts.MessagesPlaceholder{VariableName: inputKey},
	})

	llmAnswer := chains.NewLLMChain(llm, qaPrompt)
	llmAnswer.OutputKey = outputKey

	return &ReformulatedRetrievalQA{
		reformulatedQuery: reformulatedQuery,
		retriever:         &ndbRetriever{ndb},
		llmAnswer:         llmAnswer,
	}
}
