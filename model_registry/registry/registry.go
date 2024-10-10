package registry

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"model_registry/schema"
	"net/http"
	"os"
	"time"

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

type generateAccessTokenRequest struct {
	ModelName string `json:"model_name"`
	TokenName string `json:"token_name"`
}

type generateAccessTokenResponse struct {
	ModelName   string `json:"model_name"`
	AccessToken string `json:"access_token"`
}

func (registry *ModelRegistry) GenerateAccessToken(w http.ResponseWriter, r *http.Request) {
	var params generateAccessTokenRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := registry.getModel(params.ModelName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, err := createAccessToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := registry.db.Create(&schema.AccessToken{AccessToken: accessToken, Name: params.TokenName, ModelID: model.ID})
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	res := generateAccessTokenResponse{AccessToken: accessToken, ModelName: params.ModelName}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (registry *ModelRegistry) DeleteModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.URL.Query().Get("model_name")
	if modelName == "" {
		http.Error(w, "Param 'name' is not present in request.", http.StatusBadRequest)
		return
	}

	model, err := registry.getModel(modelName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = os.RemoveAll(model.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := registry.db.Delete(&model)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type listModelsRequest struct {
	NameFilter    string   `json:"name_filter"`
	TypeFilter    string   `json:"type_filter"`
	SubtypeFilter string   `json:"subtype_filter"`
	AccessTokens  []string `json:"access_tokens"`
}

func (filters *listModelsRequest) applyFilters(query *gorm.DB) *gorm.DB {
	if filters.NameFilter != "" {
		query = query.Where("name ILIKE ?", fmt.Sprintf("%%%v%%", filters.NameFilter))
	}
	if filters.TypeFilter != "" {
		query = query.Where("model_type = ?", filters.TypeFilter)
	}
	if filters.SubtypeFilter != "" {
		query = query.Where("model_subtype = ?", filters.SubtypeFilter)
	}
	return query
}

type ModelInfo struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ModelType    string    `json:"model_type"`
	ModelSubtype string    `json:"model_subtype"`
	Metadata     string    `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
}

type listModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

func (registry *ModelRegistry) ListModels(w http.ResponseWriter, r *http.Request) {
	var params listModelsRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var models []schema.Model
	result := params.applyFilters(registry.db.Where("access = ?", schema.Public)).Find(&models)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if len(params.AccessTokens) > 0 {
		var privateModels []schema.Model
		query := registry.db.Joins("AccessToken").Where("access = ?", schema.Private).Where("access_token IN ?", params.AccessTokens)
		result := params.applyFilters(query).Find(&privateModels)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
		models = append(models, privateModels...)
	}

	modelInfos := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		modelInfos = append(modelInfos, ModelInfo{
			Name:         model.Name,
			Description:  model.Description,
			ModelType:    model.ModelType,
			ModelSubtype: model.ModelSubtype,
			Metadata:     model.Metadata,
			CreatedAt:    model.CreatedAt,
		})
	}

	res := listModelsResponse{Models: modelInfos}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type downloadRequest struct {
	ModelName   string `json:"model_name"`
	AccessToken string `json:"access_token"`
}

func (registry *ModelRegistry) DownloadModel(w http.ResponseWriter, r *http.Request) {
	var params downloadRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := registry.getModel(params.ModelName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if model.Access != schema.Public {
		var accessToken schema.AccessToken
		result := registry.db.Find("access_token = ?", params.AccessToken).Take(&accessToken)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusBadRequest)
			return
		}
		if accessToken.ModelID != model.ID {
			http.Error(w, fmt.Sprintf("Provided access token does not match model %v'.", params.ModelName), http.StatusUnauthorized)
			return
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, fmt.Sprintf("Http response does not support chunked response."), http.StatusInternalServerError)
		return
	}

	file, err := os.Open(model.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to open file for model '%v'", params.ModelName), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	buffer := bufio.NewReader(file)
	chunk := make([]byte, 1024*1024)

	for {
		readN, err := buffer.Read(chunk)
		if err != nil && err != io.EOF {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeN, err := w.Write(chunk[:readN])
		if writeN != readN {
			http.Error(w, fmt.Sprintf("Expected to write %d bytes to stream, wrote %d", readN, writeN), http.StatusInternalServerError)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flusher.Flush() // Sends chunk

		if err == io.EOF {
			break
		}
	}
}

func (registry *ModelRegistry) StartUpload(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) UploadChunk(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) CommitUpload(w http.ResponseWriter, r *http.Request) {

}

func (registry *ModelRegistry) getModel(name string) (schema.Model, error) {
	var model schema.Model
	result := registry.db.Take(&model, "name = ?", name)
	if result.Error != nil {
		return model, result.Error
	}
	if result.RowsAffected != 1 {
		return model, fmt.Errorf("Unable to find model with name '%v'", name)
	}

	return model, nil
}
