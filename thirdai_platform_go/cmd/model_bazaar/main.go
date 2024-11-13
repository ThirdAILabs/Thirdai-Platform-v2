package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/services"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Dsn                 string
	NomadAddr           string
	NomadToken          string
	ShareDir            string
	LicensePath         string
	ModelBazaarEndpoint string

	LocalDriver  *nomad.LocalDriver
	DockerDriver *nomad.DockerDriver

	AdminUsername string
	AdminEmail    string
	AdminPassword string

	Port int
}

func (c *Config) Driver() nomad.Driver {
	if c.LocalDriver != nil {
		return *c.LocalDriver
	}
	if c.DockerDriver != nil {
		return *c.DockerDriver
	}
	panic("either LocalDriver or DockerDriver must be specified in config")
}

func initLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

func initDb(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("error opening database connection: %v", err)
	}

	err = db.AutoMigrate(
		&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
		&schema.User{}, &schema.Team{}, &schema.UserTeam{},
	)
	if err != nil {
		log.Panicf("error migrating db schema: %v", err)
	}

	return db
}

func main() {
	var c Config

	logFile, err := os.OpenFile(filepath.Join(c.ShareDir, "logs/model_bazaar.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Panicf("error opening log file: %v", err)
	}
	defer logFile.Close()

	initLogging(logFile)

	model_bazaar := services.NewModelBazaar(
		initDb(c.Dsn),
		nomad.NewHttpClient(c.NomadAddr, c.NomadToken),
		storage.NewSharedDisk(c.ShareDir),
		licensing.NewVerifier(c.LicensePath),
		services.Variables{
			Driver:              c.Driver(),
			ModelBazaarEndpoint: c.ModelBazaarEndpoint,
		},
	)

	go model_bazaar.StartStatusSync(5 * time.Second)

	model_bazaar.InitAdmin(c.AdminUsername, c.AdminEmail, c.AdminPassword)

	r := chi.NewRouter()
	r.Mount("/api/v1", model_bazaar.Routes())

	slog.Info("starting serrver", "port", c.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", c.Port), r)
	if err != nil {
		log.Fatalf(err.Error())
	}
	model_bazaar.StopStatusSync()
}
