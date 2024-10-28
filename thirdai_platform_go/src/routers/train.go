package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"thirdai_platform/src/auth"
	"thirdai_platform/src/config"
	"thirdai_platform/src/licensing"
	"thirdai_platform/src/nomad"
	"thirdai_platform/src/schema"
	"thirdai_platform/src/storage"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainRouter struct {
	db      *gorm.DB
	nomad   nomad.NomadClient
	storage storage.Storage
	license licensing.LicenseVerifier
}

func parseMultipartRequest(r *http.Request, saveDir string, storage storage.Storage) (map[string][]byte, map[string]string, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return nil, nil, fmt.Errorf("missing 'Content-Type' header")
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing media type in request: %v", err)
	}
	if mediaType != "multipart/form-data" {
		return nil, nil, fmt.Errorf("expected media type to be 'multipart/form-data'")
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, nil, fmt.Errorf("missing 'boundary' parameter in 'Content-Type' header")
	}

	reader := multipart.NewReader(r.Body, boundary)

	forms := make(map[string][]byte)
	files := make(map[string]string)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing multipart request: %v", err)
		}
		defer part.Close()

		if part.FormName() == "files" {
			newFilepath := filepath.Join(saveDir, part.FileName())
			err := storage.Write(newFilepath, part)
			if err != nil {
				return nil, nil, fmt.Errorf("error saving file '%v': %v", part.FileName(), err)
			}
			files[part.FileName()] = newFilepath
		} else {
			data, err := io.ReadAll(part)
			if err != nil {
				return nil, nil, fmt.Errorf("error reading part '%v': %v", part.FormName(), err)
			}
			forms[part.FormName()] = data
		}
	}

	return forms, files, nil
}

func updatePaths(files []config.FileInfo, requestFiles map[string]string) ([]config.FileInfo, error) {
	output := make([]config.FileInfo, 0, len(files))

	for _, file := range files {
		if file.Location == config.FileLocLocal {
			newPath, ok := requestFiles[file.Path]
			if !ok {
				return nil, fmt.Errorf("file %v is not found in request files")
			}
			output = append(output, config.FileInfo{
				Path:     newPath,
				Location: file.Location,
				DocId:    file.DocId,
				Options:  file.Options,
				Metadata: file.Metadata,
			})
		} else {
			output = append(output, file)
		}
	}

	return output, nil
}

type ndbTrainOptions struct {
	ModelName    string             `json:"model_name"`
	BaseModelId  *string            `json:"base_model_id"`
	ModelOptions *config.NdbOptions `json:"model_options"`
	Data         config.NDBData     `json:"data"`
	JobOptions   config.JobOptions  `json:"job_options"`
}

func (t *TrainRouter) TrainNdb(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	formData, files, err := parseMultipartRequest(
		r, filepath.Join(storage.DataPath(modelId), "train"), t.storage,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	optionsData, ok := formData["options"]
	if !ok {
		http.Error(w, fmt.Sprintf("missing options in train request"), http.StatusBadRequest)
		return
	}

	var options ndbTrainOptions
	err = json.Unmarshal(optionsData, &options)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing train options: %v", err), http.StatusBadRequest)
		return
	}

	if options.ModelOptions == nil {
		// TODO(Nicholas): set default options
	}

	if len(options.Data.SupervisedFiles)+len(options.Data.UnsupervisedFiles) == 0 {
		http.Error(w, "unsupervised or supervised data must be specified for training", http.StatusBadRequest)
		return
	}

	options.JobOptions.EnsureValid()

	unsup, err := updatePaths(options.Data.UnsupervisedFiles, files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	options.Data.UnsupervisedFiles = unsup

	sup, err := updatePaths(options.Data.SupervisedFiles, files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	options.Data.SupervisedFiles = sup

	test, err := updatePaths(options.Data.TestFiles, files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	options.Data.TestFiles = test

	// TODO(Nicholas): Get total cpu usage
	license, err := t.license.Verify(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      t.storage.Location(),
		LicenseKey:          license.BoltLicenseKey,
		ModelBazaarEndpoint: "TODO(Nicholas)",
		ModelId:             modelId,
		BaseModelId:         options.BaseModelId,
		ModelOptions:        options.ModelOptions,
		Data:                options.Data,
		JobOptions:          options.JobOptions,
		IsRetraining:        false,
	}

	trainConfigData, err := json.Marshal(trainConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("error saving train config: %v", err), http.StatusBadRequest)
		return
	}

	configPath := filepath.Join(storage.ModelPath(modelId), "train_config.json")
	err = t.storage.Write(configPath, bytes.NewReader(trainConfigData))
	if err != nil {
		http.Error(w, fmt.Sprintf("error saving train config: %v", err), http.StatusBadRequest)
		return
	}

	subtype, ok := options.ModelOptions.NdbOptions["ndb_subtype"]
	if !ok {
		http.Error(w, fmt.Sprintf("invalid/missing ndb subtype"), http.StatusBadRequest)
		return
	}

	model := schema.Model{
		Id:                modelId,
		Name:              options.ModelName,
		Type:              schema.NdbType,
		Subtype:           subtype.(string), // TODO(Nicholas) : check cast
		PublishedDate:     time.Now(),
		TrainStatus:       schema.NotStarted,
		DeployStatus:      schema.NotStarted,
		Access:            schema.Private,
		DefaultPermission: schema.ReadPerm,
		ParentId:          options.BaseModelId,
		UserId:            userId,
	}

	err = t.db.Transaction(func(db *gorm.DB) error {
		if options.BaseModelId != nil {
			exists, err := schema.ModelExists(db, *options.BaseModelId)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("base model %v does not exist", *options.BaseModelId)
			}
		}

		//TODO(Nicholas): check for duplicate model name for user

		result := db.Create(&model)
		if result.Error != nil {
			return fmt.Errorf("database error: %v", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = t.nomad.StartJob(
		"train_job.hcl", // TODO(Nicholas): template dir path?
		nomad.TrainJob{
			JobName:      nomad.TrainJobName(model),
			TrainScript:  "TODO(Nicholas).py",
			ConfigPath:   configPath,
			PythonPath:   "TODO(Nicholas)",
			PlatformType: "TODO(Nicholas)",
			Platform:     nomad.DockerPlatform{}, // TODO(Nicholas)
			Resources: nomad.Resource{
				AllocationMhz:       options.JobOptions.CpuUsageMhz(),
				AllocationMemory:    options.JobOptions.AllocationMemory,
				AllocationMemoryMax: 60000,
			},
		},
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start training job: %v", err), http.StatusBadRequest)
		return
	}

	model.TrainStatus = schema.Starting

	result := t.db.Save(&model)
	if result.Error != nil {
		dbError(w, result.Error)
		return
	}

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}
