package main

import (
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

// import (
// 	"log"
// 	"model_registry/registry"
// 	"model_registry/schema"
// 	"net/http"

// 	"github.com/go-chi/chi/middleware"
// 	"github.com/go-chi/chi/v5"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// func main() {
// 	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("failed to connect database")
// 	}

// 	db.AutoMigrate(&schema.Model{})
// 	db.AutoMigrate(&schema.AccessToken{})
// 	db.AutoMigrate(&schema.Admin{})

// 	registry := registry.New(db)

// 	r := chi.NewRouter()
// 	r.Use(middleware.Logger)
// 	r.Use(middleware.Recoverer)

// 	r.Mount("/api/v1", registry.Routes())

// 	err = http.ListenAndServe(":3000", r)
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// }

func main() {
	tokenAuth := jwtauth.New("HS256", []byte("secret-249024"), nil)

	tok, _, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123, "exp": time.Now().Add(time.Minute * 5)})

	fmt.Println(tok.Expiration())

}
