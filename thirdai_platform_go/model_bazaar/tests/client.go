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

type client struct {
	api    chi.Router
	token  string
	userId string
}

type loginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func jsonError(err error) error {
	return fmt.Errorf("json encode/decode error: %w", err)
}

func get[T any](c *client, endpoint string) (T, error) {
	req := httptest.NewRequest("GET", endpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token))
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	var data T

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return data, ErrUnauthorized
		}
		return data, fmt.Errorf("get %v failed with status %d and res '%v'", endpoint, res.StatusCode, w.Body.String())
	}

	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

type NoBody struct{}

func post[T any](c *client, endpoint string, body []byte) (T, error) {
	return postWithToken[T](c, endpoint, body, c.token)
}

func postWithToken[T any](c *client, endpoint string, body []byte, token string) (T, error) {
	return postWithHeaders[T](c, endpoint, body, map[string]string{"Authorization": fmt.Sprintf("Bearer %v", token)})
}

func postWithHeaders[T any](c *client, endpoint string, body []byte, headers map[string]string) (T, error) {
	req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	var data T

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return data, ErrUnauthorized
		}
		return data, fmt.Errorf("post %v failed with status %d and res '%v'", endpoint, res.StatusCode, w.Body.String())
	}

	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func deleteReq(c *client, endpoint string) error {
	req := httptest.NewRequest("DELETE", endpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token))

	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return fmt.Errorf("post %v failed with status %d and res '%v'", endpoint, res.StatusCode, w.Body.String())
	}

	return nil
}

var ErrUnauthorized = errors.New("unauthorized")

func (c *client) signup(username, email, password string) (loginInfo, error) {
	body, err := json.Marshal(map[string]string{
		"email": email, "username": username, "password": password,
	})
	if err != nil {
		return loginInfo{}, jsonError(err)
	}

	_, err = post[map[string]string](c, "/user/signup", body)
	if err != nil {
		return loginInfo{}, err
	}

	return loginInfo{Email: email, Password: password}, nil
}

func (c *client) login(login loginInfo) error {
	body, err := json.Marshal(login)
	if err != nil {
		return jsonError(err)
	}

	data, err := post[map[string]string](c, "/user/login", body)
	if err != nil {
		return err
	}

	c.token = data["access_token"]
	c.userId = data["user_id"]

	return nil
}

func (c *client) promoteAdmin(userId string) error {
	_, err := post[NoBody](c, fmt.Sprintf("/user/%v/admin", userId), nil)
	return err
}

func (c *client) demoteAdmin(userId string) error {
	return deleteReq(c, fmt.Sprintf("/user/%v/admin", userId))
}

func (c *client) listUsers() ([]services.UserInfo, error) {
	return get[[]services.UserInfo](c, "/user/list")
}

func (c *client) userInfo() (services.UserInfo, error) {
	return get[services.UserInfo](c, "/user/info")
}

func (c *client) createTeam(name string) (string, error) {
	body := []byte(fmt.Sprintf(`{"name": "%v"}`, name))

	data, err := post[map[string]string](c, "/team/create", body)
	if err != nil {
		return "", err
	}
	return data["team_id"], nil
}

func (c *client) deleteTeam(teamId string) error {
	return deleteReq(c, fmt.Sprintf("/team/%v", teamId))
}

func (c *client) addUserToTeam(teamId, userId string) error {
	_, err := post[NoBody](c, fmt.Sprintf("/team/%v/users/%v", teamId, userId), nil)
	return err
}

func (c *client) removeUserFromTeam(teamId, userId string) error {
	return deleteReq(c, fmt.Sprintf("/team/%v/users/%v", teamId, userId))
}

func (c *client) addModelToTeam(teamId, modelId string) error {
	_, err := post[NoBody](c, fmt.Sprintf("/team/%v/models/%v", teamId, modelId), nil)
	return err
}

func (c *client) removeModelFromTeam(teamId, modelId string) error {
	return deleteReq(c, fmt.Sprintf("/team/%v/models/%v", teamId, modelId))
}

func (c *client) addTeamAdmin(teamId, userId string) error {
	_, err := post[NoBody](c, fmt.Sprintf("/team/%v/admins/%v", teamId, userId), nil)
	return err
}

func (c *client) removeTeamAdmin(teamId, userId string) error {
	return deleteReq(c, fmt.Sprintf("/team/%v/admins/%v", teamId, userId))
}

func (c *client) listTeams() ([]services.TeamInfo, error) {
	return get[[]services.TeamInfo](c, "/team/list")
}

func (c *client) listTeamModels(teamId string) ([]services.ModelInfo, error) {
	return get[[]services.ModelInfo](c, fmt.Sprintf("/team/%v/models", teamId))
}

func (c *client) listTeamUsers(teamId string) ([]services.TeamUserInfo, error) {
	return get[[]services.TeamUserInfo](c, fmt.Sprintf("/team/%v/users", teamId))
}

func (c *client) modelInfo(modelId string) (services.ModelInfo, error) {
	return get[services.ModelInfo](c, fmt.Sprintf("/model/%v", modelId))
}

func (c *client) listModels() ([]services.ModelInfo, error) {
	return get[[]services.ModelInfo](c, "/model/list")
}

func (c *client) deleteModel(modelId string) error {
	return deleteReq(c, fmt.Sprintf("/model/%v", modelId))
}

func (c *client) updateAccess(modelId, newAccess string) error {
	body := []byte(fmt.Sprintf(`{"access": "%v"}`, newAccess))
	_, err := post[NoBody](c, fmt.Sprintf("/model/%v/access", modelId), body)
	return err
}

func (c *client) updateDefaultPermission(modelId, newPermission string) error {
	body := []byte(fmt.Sprintf(`{"permission": "%v"}`, newPermission))
	_, err := post[NoBody](c, fmt.Sprintf("/model/%v/default-permission", modelId), body)
	return err
}

func (c *client) modelPermissions(modelId string) (services.ModelPermissions, error) {
	return get[services.ModelPermissions](c, fmt.Sprintf("/model/%v/permissions", modelId))
}

func (c *client) trainNdb(name string) (string, error) {
	params := services.NdbTrainOptions{
		ModelName:    name,
		ModelOptions: &config.NdbOptions{},
		Data: config.NDBData{
			UnsupervisedFiles: []config.FileInfo{{Path: "a.txt", Location: "local"}},
		},
	}

	body, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	res, err := post[map[string]string](c, "/train/ndb", body)
	if err != nil {
		return "", err
	}

	return res["model_id"], nil
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
	body := fmt.Sprintf(`{"model_name": "%v", "model_type": "%v"}`, modelName, modelType)
	data, err := post[map[string]string](c, "/model/upload", []byte(body))
	if err != nil {
		return "", err
	}
	return data["token"], nil
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
		_, err := postWithToken[NoBody](c, fmt.Sprintf("/model/upload/%d", chunk_idx), modelData[i:min(i+chunksize, len(modelData))], uploadToken)
		if err != nil {
			return "", err
		}
		chunk_idx++
	}

	data, err := postWithToken[map[string]string](c, "/model/upload/commit", nil, uploadToken)
	if err != nil {
		return "", err
	}

	return data["model_id"], nil
}

func (c *client) downloadModel(modelId string, dest string) error {
	endpoint := fmt.Sprintf("/model/%v/download", modelId)
	req := httptest.NewRequest("GET", endpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token))
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
	return get[services.StatusResponse](c, fmt.Sprintf("/train/%v/status", modelId))
}

func (c *client) deployStatus(modelId string) (services.StatusResponse, error) {
	return get[services.StatusResponse](c, fmt.Sprintf("/deploy/%v/status", modelId))
}

func (c *client) deploy(modelId string) error {
	_, err := post[NoBody](c, fmt.Sprintf("/deploy/%v", modelId), []byte("{}"))
	return err
}

func (c *client) undeploy(modelId string) error {
	return deleteReq(c, fmt.Sprintf("/deploy/%v", modelId))
}
