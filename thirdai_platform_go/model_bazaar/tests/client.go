package tests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/services"

	"github.com/go-chi/chi/v5"
)

type httpTestRequest struct {
	api http.Handler

	method   string
	endpoint string
	headers  map[string]string
	json     interface{}
	body     io.Reader
}

func newHttpTestRequest(api http.Handler, method, endpoint string) *httpTestRequest {
	return &httpTestRequest{
		api:      api,
		method:   method,
		endpoint: endpoint,
		headers:  nil,
		json:     nil,
		body:     nil,
	}
}

func (r *httpTestRequest) Header(key, value string) *httpTestRequest {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *httpTestRequest) Auth(token string) *httpTestRequest {
	return r.Header("Authorization", fmt.Sprintf("Bearer %v", token))
}

func (r *httpTestRequest) Json(data interface{}) *httpTestRequest {
	r.json = data
	return r
}

func (r *httpTestRequest) Body(body io.Reader) *httpTestRequest {
	r.body = body
	return r
}

// response body will be parsed into result, passing nil indicates that no result is returned.
func (r *httpTestRequest) Do(result interface{}) error {
	if r.json != nil {
		body := new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(r.json)
		if err != nil {
			return fmt.Errorf("error encoding json body for endpoint %v: %w", r.endpoint, err)
		}
		r.body = body
	}

	req := httptest.NewRequest(r.method, r.endpoint, r.body)
	if r.headers != nil {
		for k, v := range r.headers {
			req.Header.Add(k, v)
		}
	}

	w := httptest.NewRecorder()

	r.api.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return fmt.Errorf("%v request to endpoint %v returned status %d, content '%v'", r.method, r.endpoint, res.StatusCode, w.Body.String())
	}

	if result != nil {
		err := json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error parsing %v response from endpoint %v: %w", r.method, r.endpoint, err)
		}
	}

	return nil
}

var ErrUnauthorized = errors.New("unauthorized")

type client struct {
	api       chi.Router
	authToken string
	userId    string
}

func (c *client) Get(endpoint string) *httpTestRequest {
	r := newHttpTestRequest(c.api, "GET", endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
}

func (c *client) Post(endpoint string) *httpTestRequest {
	r := newHttpTestRequest(c.api, "POST", endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
}

func (c *client) Delete(endpoint string) *httpTestRequest {
	r := newHttpTestRequest(c.api, "DELETE", endpoint)
	if c.authToken != "" {
		return r.Auth(c.authToken)
	}
	return r
}

type loginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *client) signup(username, email, password string) (loginInfo, error) {
	body := map[string]string{
		"email": email, "username": username, "password": password,
	}

	err := c.Post("/user/signup").Json(body).Do(nil)
	if err != nil {
		return loginInfo{}, err
	}

	return loginInfo{Email: email, Password: password}, nil
}

func (c *client) login(login loginInfo) error {
	var res map[string]string
	err := c.Post("/user/login").Json(login).Do(&res)
	if err != nil {
		return err
	}

	c.authToken = res["access_token"]
	c.userId = res["user_id"]

	return nil
}

func (c *client) addUser(username, email, password string) (loginInfo, error) {
	body := map[string]string{
		"email": email, "username": username, "password": password,
	}

	err := c.Post("/user/create").Json(body).Do(nil)
	if err != nil {
		return loginInfo{}, err
	}

	return loginInfo{Email: email, Password: password}, nil
}

func (c *client) deleteUser(userId string) error {
	return c.Delete(fmt.Sprintf("/user/%v", userId)).Do(nil)
}

func (c *client) promoteAdmin(userId string) error {
	return c.Post(fmt.Sprintf("/user/%v/admin", userId)).Do(nil)
}

func (c *client) demoteAdmin(userId string) error {
	return c.Delete(fmt.Sprintf("/user/%v/admin", userId)).Do(nil)
}

func (c *client) listUsers() ([]services.UserInfo, error) {
	var res []services.UserInfo
	err := c.Get("/user/list").Do(&res)
	return res, err
}

func (c *client) userInfo() (services.UserInfo, error) {
	var res services.UserInfo
	err := c.Get("/user/info").Do(&res)
	return res, err
}

func (c *client) createTeam(name string) (string, error) {
	body := map[string]string{"name": name}

	var res map[string]string
	err := c.Post("/team/create").Json(body).Do(&res)
	return res["team_id"], err
}

func (c *client) deleteTeam(teamId string) error {
	return c.Delete(fmt.Sprintf("/team/%v", teamId)).Do(nil)
}

func (c *client) addUserToTeam(teamId, userId string) error {
	return c.Post(fmt.Sprintf("/team/%v/users/%v", teamId, userId)).Do(nil)
}

func (c *client) removeUserFromTeam(teamId, userId string) error {
	return c.Delete(fmt.Sprintf("/team/%v/users/%v", teamId, userId)).Do(nil)
}

func (c *client) addModelToTeam(teamId, modelId string) error {
	return c.Post(fmt.Sprintf("/team/%v/models/%v", teamId, modelId)).Do(nil)
}

func (c *client) removeModelFromTeam(teamId, modelId string) error {
	return c.Delete(fmt.Sprintf("/team/%v/models/%v", teamId, modelId)).Do(nil)
}

func (c *client) addTeamAdmin(teamId, userId string) error {
	return c.Post(fmt.Sprintf("/team/%v/admins/%v", teamId, userId)).Do(nil)
}

func (c *client) removeTeamAdmin(teamId, userId string) error {
	return c.Delete(fmt.Sprintf("/team/%v/admins/%v", teamId, userId)).Do(nil)
}

func (c *client) listTeams() ([]services.TeamInfo, error) {
	var res []services.TeamInfo
	err := c.Get("/team/list").Do(&res)
	return res, err
}

func (c *client) listTeamModels(teamId string) ([]services.ModelInfo, error) {
	var res []services.ModelInfo
	err := c.Get(fmt.Sprintf("/team/%v/models", teamId)).Do(&res)
	return res, err
}

func (c *client) listTeamUsers(teamId string) ([]services.TeamUserInfo, error) {
	var res []services.TeamUserInfo
	err := c.Get(fmt.Sprintf("/team/%v/users", teamId)).Do(&res)
	return res, err
}

func (c *client) modelInfo(modelId string) (services.ModelInfo, error) {
	var res services.ModelInfo
	err := c.Get(fmt.Sprintf("/model/%v", modelId)).Do(&res)
	return res, err
}

func (c *client) listModels() ([]services.ModelInfo, error) {
	var res []services.ModelInfo
	err := c.Get("/model/list").Do(&res)
	return res, err
}

func (c *client) listModelsWithWriteAccess() ([]services.ModelInfo, error) {
	var res []services.ModelInfo
	err := c.Get("/model/list-model-write-access").Do(&res)
	return res, err
}

func (c *client) createAPIKey(modelIDs []string, name string, expiry string) (string, error) {
	requestBody := map[string]interface{}{
		"model_ids": modelIDs,
		"name":      name,
		"exp":       expiry,
	}

	var response struct {
		ApiKey string `json:"api_key"`
	}
	err := c.Post("/model/create-api-key").Json(requestBody).Do(&response)
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}

	return response.ApiKey, nil
}

func (c *client) deleteModel(modelId string) error {
	return c.Delete(fmt.Sprintf("/model/%v", modelId)).Do(nil)
}

func (c *client) updateAccess(modelId, newAccess string) error {
	body := map[string]string{"access": newAccess}
	return c.Post(fmt.Sprintf("/model/%v/access", modelId)).Json(body).Do(nil)
}

func (c *client) updateDefaultPermission(modelId, newPermission string) error {
	body := map[string]string{"permission": newPermission}
	return c.Post(fmt.Sprintf("/model/%v/default-permission", modelId)).Json(body).Do(nil)
}

func (c *client) modelPermissions(modelId string) (services.ModelPermissions, error) {
	var res services.ModelPermissions
	err := c.Get(fmt.Sprintf("/model/%v/permissions", modelId)).Do(&res)
	return res, err
}

func (c *client) trainNdbDummyFile(name string) (string, error) {
	return c.trainNdb(name, config.TrainFile{Path: "n/a", Location: "s3"})
}

func (c *client) trainNdb(name string, file config.TrainFile) (string, error) {
	body := services.NdbTrainRequest{
		ModelName:    name,
		ModelOptions: &config.NdbOptions{},
		Data: config.NDBData{
			UnsupervisedFiles: []config.TrainFile{file},
		},
	}

	var res map[string]string
	err := c.Post("/train/ndb").Json(body).Do(&res)
	return res["model_id"], err
}

func (c *client) trainNlpToken(name string) (string, error) {
	body := services.NlpTokenTrainRequest{
		ModelName: name,
		ModelOptions: &config.NlpTokenOptions{
			TargetLabels: []string{"NAME", "EMAIL"},
			SourceColumn: "source",
			TargetColumn: "target",
			DefaultTag:   "O",
		},
		Data: config.NlpData{
			SupervisedFiles: []config.TrainFile{{Path: "a.txt", Location: "s3"}},
		},
	}

	var res map[string]string
	err := c.Post("/train/nlp-token").Json(body).Do(&res)
	return res["model_id"], err
}

func (c *client) createEnterpriseSearch(name, ndb, guardrail string) (string, error) {
	body := map[string]string{
		"model_name": name, "retrieval_id": ndb, "guardrail_id": guardrail,
	}

	var res map[string]string
	err := c.Post("/workflow/enterprise-search").Json(body).Do(&res)
	return res["model_id"], err
}

func zipDir(path string) (string, error) {
	newPath := path + ".zip"
	file, err := os.Create(newPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	zip := zip.NewWriter(file)
	defer zip.Close()

	err = zip.AddFS(os.DirFS(path))
	if err != nil {
		return "", err
	}

	return newPath, nil
}

func unzipDir(path string) error {
	newPath := strings.TrimSuffix(path, ".zip")
	zip, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("error opening zip reader: %w", err)
	}
	defer zip.Close()

	for _, file := range zip.File {
		if strings.HasSuffix(file.Name, "/") {
			continue // directory
		}
		fileData, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file in zipfile %v: %w", file.Name, err)
		}
		defer fileData.Close()

		part := filepath.Join(newPath, file.Name)
		err = os.MkdirAll(filepath.Dir(part), 0777)
		if err != nil {
			return fmt.Errorf("error making dir for file in zip: %w", err)
		}

		newFile, err := os.Create(part)
		if err != nil {
			return fmt.Errorf("error opening file for file in zip: %w", err)
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, fileData)
		if err != nil {
			return fmt.Errorf("error writing contents from zipfile %v: %w", file.Name, err)
		}
	}

	return nil
}

func (c *client) startUpload(modelName, modelType string) (string, error) {
	body := map[string]string{"model_name": modelName, "model_type": modelType}

	var res map[string]string
	err := c.Post("/model/upload").Json(body).Do(&res)
	return res["token"], err
}

func (c *client) uploadModel(modelName, modelType, path string, chunksize int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		newPath, err := zipDir(path)
		if err != nil {
			return "", err
		}
		path = newPath
	}

	modelData, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	uploadToken, err := c.startUpload(modelName, modelType)
	if err != nil {
		return "", err
	}

	chunk_idx := 0
	for i := 0; i < len(modelData); i += chunksize {
		chunk := modelData[i:min(i+chunksize, len(modelData))]
		err := c.Post(fmt.Sprintf("/model/upload/%d", chunk_idx)).Auth(uploadToken).Body(bytes.NewReader(chunk)).Do(nil)
		if err != nil {
			return "", err
		}
		chunk_idx++
	}

	var res map[string]string
	err = c.Post("/model/upload/commit").Auth(uploadToken).Do(&res)
	return res["model_id"], err
}

func (c *client) downloadModel(modelId string, dest string) error {
	endpoint := fmt.Sprintf("/model/%v/download", modelId)
	req := httptest.NewRequest("GET", endpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.authToken))
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return fmt.Errorf("get %v failed with status %d and res '%v'", endpoint, res.StatusCode, w.Body.String())
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, w.Body)
	if err != nil {
		return err
	}

	if strings.HasSuffix(dest, ".zip") {
		err := unzipDir(dest)
		if err != nil {
			return fmt.Errorf("error unzipping result: %w", err)
		}
	}

	return nil
}

func (c *client) trainStatus(modelId string) (services.StatusResponse, error) {
	var res services.StatusResponse
	err := c.Get(fmt.Sprintf("/train/%v/status", modelId)).Do(&res)
	return res, err
}

func (c *client) deployStatus(modelId string) (services.StatusResponse, error) {
	var res services.StatusResponse
	err := c.Get(fmt.Sprintf("/deploy/%v/status", modelId)).Do(&res)
	return res, err
}

func (c *client) deploy(modelId string) error {
	return c.Post(fmt.Sprintf("/deploy/%v", modelId)).Json(struct{}{}).Do(nil)
}

func (c *client) undeploy(modelId string) error {
	return c.Delete(fmt.Sprintf("/deploy/%v", modelId)).Do(nil)
}

func (c *client) trainReport(modelId string) (interface{}, error) {
	var res interface{}
	err := c.Get(fmt.Sprintf("/train/%v/report", modelId)).Do(&res)
	return res, err
}
