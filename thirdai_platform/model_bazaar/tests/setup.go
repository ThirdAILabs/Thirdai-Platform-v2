package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/services"
	"thirdai_platform/model_bazaar/storage"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testEnv struct {
	modelBazaar services.ModelBazaar
	api         chi.Router
	storage     storage.Storage
	nomad       *NomadStub
}

const (
	adminUsername = "admin123"
	adminEmail    = "admin123@mail.com"
	adminPassword = "admin_password123"
)

func setupTestEnv(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(
		&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
		&schema.User{}, &schema.Team{}, &schema.UserTeam{}, &schema.JobLog{},
		&schema.Upload{}, &schema.UserAPIKey{},
	)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	licensePath := filepath.Join(tmpDir, "/platform_license")
	file, err := os.Create(licensePath)
	if err != nil {
		t.Fatal(err)
	}
	err = json.NewEncoder(file).Encode(TEST_LICENSE)
	if err != nil {
		t.Fatal(err)
	}

	storagePath := filepath.Join(tmpDir, "/storage")
	err = os.MkdirAll(storagePath, 0777)
	if err != nil {
		t.Fatalf("error creating storate directory: %v", err)
	}

	store := storage.NewSharedDisk(storagePath)
	nomadStub := newNomadStub()

	secret := []byte("290zcv02ai249")

	userAuth, err := auth.NewBasicIdentityProvider(
		db,
		auth.NewAuditLogger(new(bytes.Buffer)),
		auth.BasicProviderArgs{
			Secret:        secret,
			AdminUsername: adminUsername,
			AdminEmail:    adminEmail,
			AdminPassword: adminPassword,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	modelBazaar := services.NewModelBazaar(
		db, nomadStub, store,
		licensing.NewVerifier(licensePath),
		userAuth,
		services.Variables{
			BackendDriver: &orchestrator.LocalDriver{},
		},
		secret,
	)

	return &testEnv{modelBazaar: modelBazaar, api: modelBazaar.Routes(), storage: store, nomad: nomadStub}
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

var TEST_LICENSE = map[string]interface{}{
	"license": map[string]string{
		"cpuMhzLimit":    "100000000",
		"expiryDate":     "2030-04-03T00:00:00+00:00",
		"boltLicenseKey": "236C00-47457C-4641C5-52E3BB-3D1F34-V3",
	},
	"signature": "SM8NMmVhdW23u9g97LnmkbG1lqiG7U07RkUdXIVll9XYI6qVYfRPbLZVNYJiYoo/iY6Jrpom/ga+NRYGDz8P+9cfpwF3CfAsdjlH41CkBTB3aZr/0t1JC/M4J3IQe5DXMF30DDjmhrrTsYsSfFcvtq8J4GG9QnMiveoB2nozuwA8Xz7XlSCujJcTFwpqFvEsJ5RGH6OuJaNXT2auuCO0EdAsNxyDOxmYxnTlKH9NdeZT9DLoEYjSfmfk4b3gLxNpmMoXPk8MJWGeoSdM99TR1wtb1JbGg/KtJSKkFkmzdCNz2dXc2ol28AkIq3eqGiU7VLh/fVZ8hvUqe7yw+FTUEw==",
}
