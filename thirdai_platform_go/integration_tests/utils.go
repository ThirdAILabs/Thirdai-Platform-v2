package integrationtests

import (
	"testing"
	"thirdai_platform/client"

	"github.com/google/uuid"
)

func getClient(t *testing.T) client.PlatformClient {
	client := client.New("http://localhost:80")
	err := client.Login("admin@mail.com", "password")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func randomName(base string) string {
	return base + "-" + uuid.New().String()
}
