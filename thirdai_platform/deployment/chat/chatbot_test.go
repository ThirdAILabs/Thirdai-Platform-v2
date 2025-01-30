package chat

import (
	"log"
	"path/filepath"
	"testing"
	"thirdai_platform/search/ndb"
)

func TestChatbot(t *testing.T) {
	const licensePath = "../../../.test_license/thirdai.license"
	if err := ndb.SetLicensePath(licensePath); err != nil {
		panic(err)
	}

	ndb, err := ndb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	chunks := []string{
		"i like to eat cookies",
		"the capital of washington is olympia",
		"apples are a fruit",
		"this is a sentence",
	}
	if err := ndb.Insert("doc", "id", chunks, nil, nil); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(t.TempDir(), "history.db")

	key := "key here"
	chatbot, err := NewOpenAIChatbot(ndb, key, dbPath)
	if err != nil {
		t.Fatal(err)
	}

	{ // Session 1 chat
		res, err := chatbot.Chat("what is the capital of washington", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(res)
	}

	{ // session 2 chat
		res, err := chatbot.Chat("what is the capital of washington", "session_2")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(res)
	}

	{ // Session 1 chat
		res, err := chatbot.Chat("what is the capital of france", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(res)
	}

	{ // Session 1 history
		history, err := chatbot.GetHistory("session_1")
		if err != nil {
			t.Fatal(err)
		}

		if len(history) != 4 {
			t.Fatal("invalid history length")
		}
	}

	{ // Session 2 history
		history, err := chatbot.GetHistory("session_2")
		if err != nil {
			t.Fatal(err)
		}

		if len(history) != 2 {
			t.Fatalf("invalid history length: %v", history)
		}
	}

}
