package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseRequestBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dest)
	if err != nil {
		slog.Error("error parsing request body", "error", err)
		http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
		return false
	}
	return true
}

func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("error serializing response body", "error", err)
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func WriteSuccess(w http.ResponseWriter) {
	WriteJsonResponse(w, struct{}{})
}

func URLParam(r *http.Request, key string) (string, error) {
	param := chi.URLParam(r, key)
	if len(param) == 0 {
		return "", fmt.Errorf("missing {%v} url parameter", key)
	}
	return param, nil
}

func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)

	if len(param) == 0 {
		return uuid.Nil, fmt.Errorf("missing {%v} url parameter", key)
	}

	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' provided: %w", param, err)
	}

	return id, nil
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func ReadFileLinesBackward(fileHandle *os.File, out chan string, stop chan bool, error_ch chan error) {
	defer close(out)
	defer close(stop)
	defer close(error_ch)
	defer fileHandle.Close()

	LastPosition, err := fileHandle.Seek(0, io.SeekEnd)
	if err != nil {
		error_ch <- err
		return
	}

	var builder strings.Builder

	for currentPosition := LastPosition; currentPosition > 0; currentPosition-- {

		_, err := fileHandle.Seek(currentPosition-1, io.SeekStart) //To read the char[currentPosition]
		if err != nil {
			error_ch <- err
			return
		}
		char := make([]byte, 1)
		_, err = fileHandle.Read(char)
		if err != nil {
			error_ch <- err
			return
		}

		if char[0] == '\n' || currentPosition == 1 {
			if builder.Len() > 0 {
				// Reverse the collected string
				runes := []rune(builder.String())
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				select {
				case <-stop:
					return
				case out <- string(runes):
				}
				builder.Reset()
			}

			if currentPosition == 1 {
				// reached at the beginning
				error_ch <- io.EOF
			}
		} else {
			builder.WriteByte(char[0])
		}
	}
}
