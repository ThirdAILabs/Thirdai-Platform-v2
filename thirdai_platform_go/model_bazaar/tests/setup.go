package tests

import (
	"testing"
	"thirdai_platform/model_bazaar/services"
	"thirdai_platform/model_bazaar/schema"

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

	err = db.AutoMigrate(
		&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
		&schema.User{}, &schema.Team{}, &schema.UserTeam{},
	)
	if err != nil {
		t.Fatal(err)
	}

	modelBazaar := services.NewModelBazaar(db)

	modelBazaar.InitAdmin(adminUsername, adminEmail, adminPassword)

	return testEnv{api: modelBazaar.Routes()}
}

func (t *testEnv) newClient() client {
	return client{api: t.api}
}

func (t *testEnv) newUser(username string) (client, error) {
	c := t.newClient()
	login, err := c.signup(username, username+"@mail.com", username+"_password")
	if err != nil {
		return client{}, err
	}

	err = c.login(login)
	if err != nil {
		return client{}, err
	}

	return c, nil
}

func (t *testEnv) adminClient() (client, error) {
	c := t.newClient()
	err := c.login(loginInfo{Email: adminEmail, Password: adminPassword})
	return c, err
}
