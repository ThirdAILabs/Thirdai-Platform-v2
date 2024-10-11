package main

import (
	"flag"
	"fmt"
	"log"
	"model_registry/registry"
	"model_registry/schema"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := flag.String("db", "registry.db", "The sqlite db to create/use")
	storagePath := flag.String("storage", "storage", "The directory to use for local storage")
	adminEmail := flag.String("email", "", "The admin email to use")
	adminPassword := flag.String("password", "", "The admin password to use")
	port := flag.Int("port", 3040, "The port to run on")

	flag.Parse()

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database")
	}

	db.AutoMigrate(&schema.Model{})
	db.AutoMigrate(&schema.AccessToken{})
	db.AutoMigrate(&schema.Admin{})

	storage := registry.NewLocalStorage(*storagePath)

	registry := registry.New(db, storage)

	registry.AddAdmin(*adminEmail, *adminPassword)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/api/v1", registry.Routes())

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
