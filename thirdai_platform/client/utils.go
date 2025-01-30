package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type loginInfo struct {
	email, password string
}

type httpRequest struct {
	method      string
	baseUrl     string
	endpoint    string
	headers     map[string]string
	queryParams map[string]string
	json        interface{}
	body        io.Reader
	login       *loginInfo
}

func newHttpRequest(method, baseUrl, endpoint string) *httpRequest {
	return &httpRequest{
		method:      method,
		baseUrl:     baseUrl,
		endpoint:    endpoint,
		headers:     nil,
		queryParams: nil,
		json:        nil,
		body:        nil,
	}
}

func (r *httpRequest) Header(key, value string) *httpRequest {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *httpRequest) Login(email, password string) *httpRequest {
	r.login = &loginInfo{email: email, password: password}
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

func (r *httpRequest) Param(key, value string) *httpRequest {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	r.queryParams[key] = value
	return r
}

func (r *httpRequest) Process(resultHandler func(io.Reader) error) error {
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

	if r.login != nil {
		req.SetBasicAuth(r.login.email, r.login.password)
	}

	if r.queryParams != nil {
		query := req.URL.Query()
		for k, v := range r.queryParams {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	start := time.Now()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending %v request to endpoint %v: %w", r.method, r.endpoint, err)
	}
	defer res.Body.Close()

	end := time.Now()

	slog.Debug("thirdai platform client", "method", r.method, "endpoint", r.endpoint, "status", res.StatusCode, "duration", end.Sub(start).String())

	if res.StatusCode != http.StatusOK {
		content, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("%v request to endpoint %v returned status %d", r.method, r.endpoint, res.StatusCode)
		}
		return fmt.Errorf("%v request to endpoint %v returned status %d, content '%v'", r.method, r.endpoint, res.StatusCode, string(content))
	}

	if resultHandler != nil {
		err := resultHandler(res.Body)
		if err != nil {
			return fmt.Errorf("error processing %v response from endpoint %v: %w", r.method, r.endpoint, err)
		}
	}

	return nil
}

func (r *httpRequest) Do(result interface{}) error {
	return r.Process(func(body io.Reader) error {
		if result != nil {
			err := json.NewDecoder(body).Decode(result)
			if err != nil {
				return fmt.Errorf("error parsing %v response from endpoint %v: %w", r.method, r.endpoint, err)
			}
		}
		return nil
	})
}

type BaseClient struct {
	baseUrl   string
	authToken string
	apiKey    string
}

func NewBaseClient(baseUrl string, authToken string) BaseClient {
	return BaseClient{baseUrl: baseUrl, authToken: authToken}
}

func (c *BaseClient) addAuthHeaders(r *httpRequest) *httpRequest {
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	if c.apiKey != "" {
		return r.Header("X-API-Key", c.apiKey)
	}
	return r
}

func (c *BaseClient) Get(endpoint string) *httpRequest {
	r := newHttpRequest("GET", c.baseUrl, endpoint)
	return c.addAuthHeaders(r)
}

func (c *BaseClient) Post(endpoint string) *httpRequest {
	r := newHttpRequest("POST", c.baseUrl, endpoint)
	return c.addAuthHeaders(r)
}

func (c *BaseClient) Delete(endpoint string) *httpRequest {
	r := newHttpRequest("DELETE", c.baseUrl, endpoint)
	return c.addAuthHeaders(r)
}

func (c *BaseClient) UseApiKey(api_key string) error {

	c.apiKey = api_key
	c.authToken = ""
	return nil
}

func addFilesToMultipart(writer *multipart.Writer, files []FileInfo) error {
	for _, fileInfo := range files {
		if fileInfo.Location != "upload" {
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

type wrappedData[T any] struct {
	Data T `json:"data"`
}
