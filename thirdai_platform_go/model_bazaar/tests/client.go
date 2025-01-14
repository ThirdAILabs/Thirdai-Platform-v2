package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
	login    *loginInfo
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

func (r *httpTestRequest) Login(email, password string) *httpTestRequest {
	r.login = &loginInfo{Email: email, Password: password}
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

	if r.login != nil {
		req.SetBasicAuth(r.login.Email, r.login.Password)
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
	err := c.Get("/user/login").Login(login.Email, login.Password).Do(&res)
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

func (c *client) startUpload(modelName string) (string, error) {
	body := map[string]string{"model_name": modelName}

	var res map[string]string
	err := c.Post("/model/upload").Json(body).Do(&res)
	return res["token"], err
}

func (c *client) uploadModel(modelName string, data io.Reader, chunksize int) (string, error) {
	uploadToken, err := c.startUpload(modelName)
	if err != nil {
		return "", err
	}

	modelData, err := io.ReadAll(data)
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

func (c *client) downloadModel(modelId string) (io.Reader, error) {
	endpoint := fmt.Sprintf("/model/%v/download", modelId)
	req := httptest.NewRequest("GET", endpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.authToken))
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("get %v failed with status %d and res '%v'", endpoint, res.StatusCode, w.Body.String())
	}

	dst := new(bytes.Buffer)

	if _, err := io.Copy(dst, w.Body); err != nil {
		return nil, err
	}

	return dst, nil
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
