package registry

import (
	"encoding/json"
	"fmt"
	"model_registry/schema"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ModelRegistry struct {
	db *gorm.DB

	tokenAuth *jwtauth.JWTAuth
}

func New(db *gorm.DB) *ModelRegistry {
	return &ModelRegistry{
		db:        db,
		tokenAuth: jwtauth.New("HS256", []byte("secret-249024"), nil),
	}
}

func (registry *ModelRegistry) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(registry.tokenAuth))
		r.Use(jwtauth.Authenticator(registry.tokenAuth))

		r.Post("/generate-access-token", registry.GenerateAccessToken)
		r.Post("/delete-model", registry.DeleteModel)
		r.Post("/upload-start", registry.StartUpload)
		r.Post("/upload-chunk", registry.UploadChunk)
		r.Post("/upload-commit", registry.CommitUpload)
	})

	r.Group(func(r chi.Router) {
		r.Post("/login", registry.Login)
		r.Get("/list-models", registry.ListModels)
		r.Get("/download-model", registry.DownloadModel)
	})

	return r
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (registry *ModelRegistry) Login(w http.ResponseWriter, r *http.Request) {
	var params loginRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var admin schema.Admin
	result := registry.db.Take(&admin, "username = ?", params.Email)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(params.Password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, token, err := registry.tokenAuth.Encode(map[string]interface{}{"email": admin.Email})
	if result.RowsAffected != 1 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := loginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (registry *ModelRegistry) GenerateAccessToken(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) DeleteModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.URL.Query().Get("name")
	if modelName == "" {
		http.Error(w, "Param 'name' is not present in request.", http.StatusBadRequest)
		return
	}

	var model schema.Model
	result := registry.db.Take(&model, "name = ?", modelName)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	if result.RowsAffected != 1 {
		http.Error(w, fmt.Sprintf("Unable to find model with name '%v'", modelName), http.StatusBadRequest)
		return
	}

	err := os.RemoveAll(model.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result = registry.db.Delete(&model)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (registry *ModelRegistry) ListModels(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) DownloadModel(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) StartUpload(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) UploadChunk(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) CommitUpload(w http.ResponseWriter, r *http.Request) {

}
