package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"thirdai_platform/model_bazaar/config"
)

func authHeader(token string) map[string]string {
	return map[string]string{"Authorization": fmt.Sprintf("bearer %v", token)}
}

type noBody struct{}

func parseRes[T any](endpoint, method string, res *http.Response) (T, error) {
	var data T
	if res.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			return data, fmt.Errorf("%v '%v' failed with status %d", method, endpoint, res.StatusCode)
		}
		return data, fmt.Errorf("%v '%v' failed with status %d, message: %v", method, endpoint, res.StatusCode, string(msg))
	}

	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("%v '%v' failed to parse response: %v", method, endpoint, err)
	}

	return data, nil
}

func get[T any](endpoint string, authToken string) (T, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return *new(T), fmt.Errorf("error constructing new request: %w", err)
	}
	headers := authHeader(authToken)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return *new(T), fmt.Errorf("GET '%v' failed: %w", endpoint, err)
	}
	defer res.Body.Close()

	return parseRes[T](endpoint, "GET", res)
}

func post[T any](endpoint string, body []byte, authToken string) (T, error) {
	return postWithHeaders[T](endpoint, body, authHeader(authToken))
}

func postWithHeaders[T any](endpoint string, body []byte, headers map[string]string) (T, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return *new(T), fmt.Errorf("error constructing new request: %w", err)
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return *new(T), fmt.Errorf("POST '%v' failed: %w", endpoint, err)
	}
	defer res.Body.Close()

	return parseRes[T](endpoint, "POST", res)
}

func deleteReq(endpoint string, authToken string) error {
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error constructing new request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", authToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST '%v' failed: %w", endpoint, err)
	}
	defer res.Body.Close()

	return nil
}

func addFilesToMultipart(writer *multipart.Writer, files []config.FileInfo) error {
	for _, fileInfo := range files {
		if fileInfo.Location != "local" {
			continue
		}
		part, err := writer.CreateFormFile("files", filepath.Base(fileInfo.Path))
		if err != nil {
			return fmt.Errorf("error creating request part: %w", err)
		}

		file, err := os.Open(fileInfo.Path)
		if err != nil {
			return fmt.Errorf("unable to open file %v: %w", fileInfo.Path, err)
		}
		defer file.Close()

		_, err = io.Copy(part, file)
		if err != nil {
			return fmt.Errorf("error writing to mulitpart request: %w", err)
		}
	}

	return nil
}

func updateLocalFilePrefixes(files []config.FileInfo, prefix string) []config.FileInfo {
	newFiles := make([]config.FileInfo, 0, len(files))

	for _, file := range files {
		var newPath string
		if file.Location == "local" {
			newPath = filepath.Join(prefix, filepath.Base(file.Path))
		} else {
			newPath = file.Path
		}

		newFiles = append(newFiles, config.FileInfo{
			Path:     newPath,
			Location: file.Location,
			DocId:    file.DocId,
			Options:  file.Options,
			Metadata: file.Metadata,
		})
	}

	return newFiles
}
