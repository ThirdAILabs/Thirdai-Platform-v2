package registry

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Storage interface {
	Type() string

	GetDownloadLink(storageUrl string, modelId uint) (string, error)

	StartUpload(modelId uint) error

	UploadChunk(modelId uint, offset int64, chunk []byte) error

	DeleteModel(modelId uint) error

	Routes() chi.Router
}

type LocalStorage struct {
	path string

	downloadAuth *jwtauth.JWTAuth
}

func NewLocalStorage(path string) Storage {
	return &LocalStorage{path: path}
}

func (s *LocalStorage) Type() string {
	return "local-storage"
}

func (s *LocalStorage) GetDownloadLink(storageUrl string, modelId uint) (string, error) {
	claims := map[string]interface{}{"model_id": modelId, "exp": time.Now().Add(time.Minute * 5)}
	_, token, err := s.downloadAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	endpoint, err := url.JoinPath(storageUrl, "/download")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v?token=%v", endpoint, token), nil
}

func (s *LocalStorage) getModelPath(modelId uint) string {
	return filepath.Join(s.path, strconv.FormatUint(uint64(modelId), 10))
}

func (s *LocalStorage) StartUpload(modelId uint) error {
	file, err := os.Create(s.getModelPath(modelId))
	if err != nil {
		return err
	}

	file.Close()

	return nil
}

func (s *LocalStorage) UploadChunk(modelId uint, offset int64, chunk []byte) error {
	file, err := os.Open(s.getModelPath(modelId))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(file)

	n, err := buf.Write(chunk)
	if err != nil {
		return err
	}
	if n != len(chunk) {
		return fmt.Errorf("Attempted to write %d bytes, wrote %d", len(chunk), n)
	}

	return nil
}

func (s *LocalStorage) DeleteModel(modelId uint) error {
	return os.Remove(s.getModelPath(modelId))
}

func (s *LocalStorage) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/download", s.Download)
	return r
}

func (s *LocalStorage) Download(w http.ResponseWriter, r *http.Request) {
	token, err := jwtauth.VerifyToken(s.downloadAuth, r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, "Unable to parse credentials", http.StatusUnauthorized)
		return
	}

	modelIdUncasted, ok := token.Get("model_id")
	if !ok {
		http.Error(w, "Invalid claims in token", http.StatusBadRequest)
		return
	}

	modelId, ok := modelIdUncasted.(uint)
	if !ok {
		http.Error(w, "Invalid claims in token", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, fmt.Sprintf("Http response does not support chunked response."), http.StatusInternalServerError)
		return
	}

	file, err := os.Open(s.getModelPath(modelId))
	if err != nil {
		http.Error(w, "Unable to open file for model", http.StatusInternalServerError)
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
