package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"thirdai_platform/model_bazaar/config"
)

type httpRequest struct {
	method   string
	baseUrl  string
	endpoint string
	headers  map[string]string
	json     interface{}
	body     io.Reader
}

func newHttpRequest(method, baseUrl, endpoint string) *httpRequest {
	return &httpRequest{
		method:   method,
		baseUrl:  baseUrl,
		endpoint: endpoint,
		headers:  nil,
		json:     nil,
		body:     nil,
	}
}

func (r *httpRequest) Header(key, value string) *httpRequest {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *httpRequest) Auth(token string) *httpRequest {
	return r.Header("Authorization", fmt.Sprintf("Bearer %v", token))
}

func (r *httpRequest) Json(data interface{}) *httpRequest {
	r.json = data
	return r
}

func (r *httpRequest) Body(body io.Reader) *httpRequest {
	r.body = body
	return r
}

func (r *httpRequest) Do(result interface{}) error {
	fullEndpoint, err := url.JoinPath(r.baseUrl, r.endpoint)
	if err != nil {
		return fmt.Errorf("error formatting url for endpoint %v: %w", r.endpoint, err)
	}

	if r.json != nil {
		body := new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(r.json)
		if err != nil {
			return fmt.Errorf("error encoding json body for endpoint %v: %w", r.endpoint, err)
		}
		r.body = body
	}

	req, err := http.NewRequest(r.method, fullEndpoint, r.body)
	if err != nil {
		return fmt.Errorf("error creating %v request for endpoint %v: %w", r.method, r.endpoint, err)
	}

	if r.headers != nil {
		for k, v := range r.headers {
			req.Header.Add(k, v)
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending %v request to endpoint %v: %w", r.method, r.endpoint, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		content, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("%v request to endpoint %v returned status %d", r.method, r.endpoint, res.StatusCode)
		}
		return fmt.Errorf("%v request to endpoint %v returned status %d, content '%v'", r.method, r.endpoint, res.StatusCode, string(content))
	}

	if result != nil {
		err := json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error parsing %v response from endpoint %v: %w", r.method, r.endpoint, err)
		}
	}

	return nil
}

type baseClient struct {
	baseUrl   string
	authToken string
}

func (c *baseClient) Get(endpoint string) *httpRequest {
	r := newHttpRequest("GET", c.baseUrl, endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
}

func (c *baseClient) Post(endpoint string) *httpRequest {
	r := newHttpRequest("POST", c.baseUrl, endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
}

func (c *baseClient) Delete(endpoint string) *httpRequest {
	r := newHttpRequest("DELETE", c.baseUrl, endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
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
			SourceId: file.SourceId,
			Options:  file.Options,
			Metadata: file.Metadata,
		})
	}

	return newFiles
}
