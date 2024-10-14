package main

import (
	"fmt"
	"log"
	"model_registry/registry"
	"model_registry/schema"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/login.html")
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/dashboard.html")
}

func getEnvWithDefault(key, fallback string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return value
}

func main() {
	dbPath := getEnvWithDefault("registry_db", "registry.db")
	storagePath := getEnvWithDefault("registry_storage", "storage")
	port := getEnvWithDefault("registry_port", "3040")

	adminEmail := os.Getenv("admin_email")
	adminPassword := os.Getenv("admin_password")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database")
	}

	err = db.AutoMigrate(&schema.Model{}, &schema.AccessToken{}, &schema.Admin{})
	if err != nil {
		log.Fatalf("Failed to setup tables: %v", err)
	}

	storage := registry.NewLocalStorage(storagePath)

	registry := registry.New(db, storage)

	registry.AddAdmin(adminEmail, adminPassword)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/api/v1", registry.Routes())

	r.Get("/login", login)
	r.Get("/dashboard", dashboard)

	err = http.ListenAndServe(fmt.Sprintf(":%v", port), r)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
