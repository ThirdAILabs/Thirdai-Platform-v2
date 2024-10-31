package tests

import (
	"testing"
	"thirdai_platform/model_bazaar/routers"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testEnv struct {
	api chi.Router
}

const (
	adminUsername = "admin123"
	adminEmail    = "admin123@mail.com"
	adminPassword = "admin_password123"
)

func setupTestEnv(t *testing.T) testEnv {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	userRouter := routers.NewUserRouter(db)
	_, err = userRouter.CreateUser(adminUsername, adminEmail, adminPassword)
	if err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Mount("/user", userRouter.Routes())

	return testEnv{api: r}
}

func (t *testEnv) newClient() client {
	return client{api: t.api}
}

func (t *testEnv) adminClient() (client, error) {
	c := t.newClient()
	err := c.login(login{Email: adminEmail, Password: adminPassword})
	return c, err
}
