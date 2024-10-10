package registry

import (
	"encoding/json"
	"fmt"
	"model_registry/schema"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ModelRegistry struct {
	db *gorm.DB

	adminAuth  *jwtauth.JWTAuth
	uploadAuth *jwtauth.JWTAuth

	storage Storage
}

func New(db *gorm.DB, storage Storage) *ModelRegistry {
	return &ModelRegistry{
		db:         db,
		adminAuth:  jwtauth.New("HS256", []byte("secret-249024"), nil),
		uploadAuth: jwtauth.New("HS256", []byte("secret-982409"), nil),
		storage:    storage,
	}
}

func (registry *ModelRegistry) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(registry.adminAuth))
		r.Use(jwtauth.Authenticator(registry.adminAuth))

		r.Post("/generate-access-token", registry.GenerateAccessToken)
		r.Post("/delete-model", registry.DeleteModel)
		r.Post("/upload-start", registry.StartUpload)
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(registry.uploadAuth))
		r.Use(jwtauth.Authenticator(registry.uploadAuth))

		r.Post("/upload-chunk", registry.UploadChunk)
		r.Post("/upload-commit", registry.CommitUpload)
	})

	r.Group(func(r chi.Router) {
		r.Post("/login", registry.Login)
		r.Get("/list-models", registry.ListModels)
		r.Get("/download-link", registry.DownloadLink)
		r.Get("/download", registry.storage.DownloadHandler())
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

	_, token, err := registry.adminAuth.Encode(map[string]interface{}{"email": admin.Email, "exp": time.Now().Add(time.Minute * 15)})
	if err != nil {
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

	err = registry.storage.DeleteModel(model.ID)
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
	Size         int64     `json:"size"`
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
	result := params.applyFilters(registry.db.Where("access = ? and status = ?", schema.Public, schema.Commited)).Find(&models)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if len(params.AccessTokens) > 0 {
		var privateModels []schema.Model
		query := registry.db.Joins("AccessToken").Where("access = ? and status = ?", schema.Private, schema.Commited).Where("access_token IN ?", params.AccessTokens)
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
			Size:         model.Size,
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

type downloadResponse struct {
	DownloadLink string `json:"download_link"`
}

func (registry *ModelRegistry) DownloadLink(w http.ResponseWriter, r *http.Request) {
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

	link, err := registry.storage.GetDownloadLink(model.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
		return
	}

	res := downloadResponse{DownloadLink: link}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// func (registry *ModelRegistry) DownloadModel(w http.ResponseWriter, r *http.Request) {
// 	var params downloadRequest
// 	dec := json.NewDecoder(r.Body)
// 	err := dec.Decode(&params)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	model, err := registry.getModel(params.ModelName)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	if model.Access != schema.Public {
// 		var accessToken schema.AccessToken
// 		result := registry.db.Find("access_token = ?", params.AccessToken).Take(&accessToken)
// 		if result.Error != nil {
// 			http.Error(w, result.Error.Error(), http.StatusBadRequest)
// 			return
// 		}
// 		if accessToken.ModelID != model.ID {
// 			http.Error(w, fmt.Sprintf("Provided access token does not match model %v'.", params.ModelName), http.StatusUnauthorized)
// 			return
// 		}
// 	}

// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, fmt.Sprintf("Http response does not support chunked response."), http.StatusInternalServerError)
// 		return
// 	}

// 	file, err := os.Open(model.Path)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Unable to open file for model '%v'", params.ModelName), http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()

// 	buffer := bufio.NewReader(file)
// 	chunk := make([]byte, 1024*1024)

// 	for {
// 		readN, err := buffer.Read(chunk)
// 		if err != nil && err != io.EOF {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		writeN, err := w.Write(chunk[:readN])
// 		if writeN != readN {
// 			http.Error(w, fmt.Sprintf("Expected to write %d bytes to stream, wrote %d", readN, writeN), http.StatusInternalServerError)
// 			return
// 		}
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		flusher.Flush() // Sends chunk

// 		if err == io.EOF {
// 			break
// 		}
// 	}
// }

type uploadRequest struct {
	ModelName    string `json:"model_name"`
	Description  string `json:"description"`
	ModelType    string `json:"model_type"`
	ModelSubtype string `json:"model_subtype"`
	Access       string `json:"access"`
	Metadata     string `json:"metadata"`
	Size         int64  `json:"size"`
}

type uploadResponse struct {
	SessionToken string `json:"session_token"`
}

func (registry *ModelRegistry) StartUpload(w http.ResponseWriter, r *http.Request) {
	var params uploadRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model *schema.Model

	err = registry.db.Transaction(func(txn *gorm.DB) error {
		var count int64
		result := txn.Model(&schema.Model{}).Where("name = ?", params.ModelName).Count(&count)
		if result.Error != nil {
			return result.Error
		}
		if count != 0 {
			return fmt.Errorf("Model with name '%v' already exists", params.ModelName)
		}

		model = &schema.Model{
			Name:         params.ModelName,
			Description:  params.Description,
			ModelType:    params.ModelType,
			ModelSubtype: params.ModelSubtype,
			Access:       params.Access,
			Metadata:     params.Metadata,
			Size:         params.Size,
			Status:       schema.Pending,
		}

		result = txn.Create(model)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create model: %v", err), http.StatusBadRequest)
		return
	}

	err = registry.storage.StartUpload(model.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
		return
	}

	_, token, err := registry.uploadAuth.Encode(map[string]interface{}{"model_id": model.ID, "exp": time.Now().Add(time.Minute * 10)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := uploadResponse{SessionToken: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type contentRange struct {
	start, end, size int64
}

func parseContentRangeHeader(value string) (contentRange, error) {
	expr := regexp.MustCompile("^bytes (\\d+)-(\\d+)/(\\d+)$")
	match := expr.FindStringSubmatch(value)
	if len(match) != 4 {
		return contentRange{}, fmt.Errorf("Invalid/missing Content-Range header")
	}

	start, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return contentRange{}, err
	}
	end, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		return contentRange{}, err
	}
	size, err := strconv.ParseInt(match[3], 10, 64)
	if err != nil {
		return contentRange{}, err
	}

	if start > end || end > size {
		return contentRange{}, fmt.Errorf("Invalid Content-Range header parameters, start must <= end, and end must be <= size")
	}

	return contentRange{start: start, end: end, size: size}, nil
}

func getModelIdFromClaims(r *http.Request) (uint, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return 0, fmt.Errorf("Error retrieving session token: %v", err)
	}

	modelIdStr, ok := claims["model_id"]
	if !ok {
		return 0, fmt.Errorf("Invalid session token, missing claims")
	}

	modelId, ok := modelIdStr.(uint)
	if !ok {
		return 0, fmt.Errorf("Invalid session token, invalid claim format")
	}

	return modelId, nil
}

func (registry *ModelRegistry) UploadChunk(w http.ResponseWriter, r *http.Request) {
	contentRange, err := parseContentRangeHeader(r.Header.Get("Content-Range"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing Content-Range header: %v", err), http.StatusBadRequest)
		return
	}

	modelId, err := getModelIdFromClaims(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO(Nicholas): Should we validate content-range size against model size?

	expectedBytes := contentRange.end - contentRange.start
	chunk := make([]byte, 0, expectedBytes)
	n, err := r.Body.Read(chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	if n != int(expectedBytes) {
		http.Error(w, fmt.Sprintf("Expected %d bytes in request, received %d", expectedBytes, n), http.StatusBadRequest)
		return
	}

	err = registry.storage.UploadChunk(modelId, contentRange.start, chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (registry *ModelRegistry) CommitUpload(w http.ResponseWriter, r *http.Request) {
	modelId, err := getModelIdFromClaims(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model schema.Model

	result := registry.db.First(&model, modelId)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	model.Status = schema.Commited

	result = registry.db.Save(&model)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
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
