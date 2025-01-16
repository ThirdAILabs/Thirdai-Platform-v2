package integrationtests

import (
	"testing"
	"thirdai_platform/model_bazaar/services"
	"time"
)

func TestBackup(t *testing.T) {
	c := getClient(t)

	initialBackups, err := c.LocalBackups()
	if err != nil {
		t.Fatal(err)
	}

	err = c.Backup(services.BackupRequest{Provider: "local"})
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.Tick(20 * time.Second)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-tick:
			backups, err := c.LocalBackups()
			if err != nil {
				t.Fatal(err)
			}
			if len(backups) > len(initialBackups) {

				return
			}
		case <-timeout:
			t.Fatal("timeout reached before backup completed")
		}
	}
}
