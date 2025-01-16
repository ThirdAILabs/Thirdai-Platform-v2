package integrationtests

import (
	"log/slog"
	"testing"
	"thirdai_platform/client"
	"time"

	"github.com/google/uuid"
)

func getClient(t *testing.T) *client.PlatformClient {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	client := client.New("http://localhost:80")
	err := client.Login("admin@mail.com", "password")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func deployModel(t *testing.T, model *client.ModelClient, autoscaling bool) {
	err := model.Deploy(autoscaling)
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
}

func randomName(base string) string {
	return base + "-" + uuid.New().String()
}
