package chat

import (
	"context"
	"fmt"
	"log/slog"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory/sqlite3"
)

type Chatbot struct {
	qaChain   chains.Chain
	sqliteUri string
}

func NewChatbot(ndb ndb.NeuralDB, llm llms.Model, sqliteUri string) (*Chatbot, error) {
	qaChain := NewReformulatedRetrievalQA(ndb, llm)

	return &Chatbot{
		qaChain:   qaChain,
		sqliteUri: sqliteUri,
	}, nil
}

func NewOpenAIChatbot(ndb ndb.NeuralDB, apiKey, sqliteUri string) (*Chatbot, error) {
	openai, err := openai.New(
		openai.WithModel("gpt-4o-mini"),
		openai.WithToken(apiKey),
		// Note: openai.WithBaseURL(), can be used for on prem llm
	)
	if err != nil {
		return nil, fmt.Errorf("error creating openai client: %w", err)
	}

	return NewChatbot(ndb, openai, sqliteUri)
}

func (chat *Chatbot) getHistoryConn(ctx context.Context, session string) *sqlite3.SqliteChatMessageHistory {
	return sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithDBAddress(chat.sqliteUri),
		sqlite3.WithSession(session),
		sqlite3.WithContext(ctx),
	)
}

func (chat *Chatbot) GetHistory(session string) ([]llms.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := chat.getHistoryConn(ctx, session)

	msgs, err := conn.Messages(ctx)
	if err != nil {
		slog.Error("error getting chat history", "session", session, "error", err)
		return nil, fmt.Errorf("error getting chat history: %w", err)
	}

	return msgs, nil
}

func (chat *Chatbot) Chat(userInput string, session string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn := chat.getHistoryConn(ctx, session)

	if err := conn.AddUserMessage(ctx, userInput); err != nil {
		slog.Error("error adding user message to chat history", "session", session, "error", err)
		return "", fmt.Errorf("error adding user message to chat history: %w", err)
	}

	msgs, err := conn.Messages(ctx)
	if err != nil {
		slog.Error("error getting chat history", "session", session, "error", err)
		return "", fmt.Errorf("error getting chat history: %w", err)
	}

	generation, err := chains.Run(ctx, chat.qaChain, msgs)
	if err != nil {
		slog.Error("error running chatbot", "session", session, "error", err)
		return "", fmt.Errorf("error running chatbot: %w", err)
	}

	if err := conn.AddAIMessage(ctx, generation); err != nil {
		slog.Error("error adding ai message to chat history", "session", session, "error", err)
		return "", fmt.Errorf("error adding ai message to chat history: %w", err)
	}

	return generation, nil
}
