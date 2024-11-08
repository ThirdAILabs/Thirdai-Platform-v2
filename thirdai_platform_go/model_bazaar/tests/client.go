package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	var data T

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return data, ErrUnauthorized
		}
		return data, fmt.Errorf("get %v failed with status %d and res %v", endpoint, res.StatusCode, w.Body.String())
	}

	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func post[T any](c *client, endpoint string, body []byte, parseRes bool) (T, error) {
	req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	var data T

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return data, ErrUnauthorized
		}
		return data, fmt.Errorf("post %v failed with status %d and res %v", endpoint, res.StatusCode, w.Body.String())
	}

	if parseRes {
		err := json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			return data, err
		}
	}

	return data, nil
}

func (c *client) addAuthHeader(r *http.Request) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token))
}

var ErrUnauthorized = errors.New("unauthorized")

func (c *client) signup(username, email, password string) (loginInfo, error) {
	body, err := json.Marshal(map[string]string{
		"email": email, "username": username, "password": password,
	})
	if err != nil {
		return loginInfo{}, jsonError(err)
	}

	_, err = post[map[string]string](c, "/user/signup", body, true)
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

	data, err := post[map[string]string](c, "/user/login", body, true)
	if err != nil {
		return err
	}

	c.token = data["access_token"]
	c.userId = data["user_id"]

	return nil
}

func (c *client) promoteAdmin(userId string) error {
	body := []byte(fmt.Sprintf(`{"user_id": "%v"}`, userId))

	_, err := post[int](c, "/user/promote-admin", body, false)
	return err
}

func (c *client) demoteAdmin(userId string) error {
	body := []byte(fmt.Sprintf(`{"user_id": "%v"}`, userId))

	_, err := post[int](c, "/user/demote-admin", body, false)
	return err
}

func (c *client) listUsers() ([]services.UserInfo, error) {
	return get[[]services.UserInfo](c, "/user/list")
}

func (c *client) userInfo() (services.UserInfo, error) {
	return get[services.UserInfo](c, "/user/info")
}

func (c *client) createTeam(name string) (string, error) {
	data, err := post[map[string]string](c, fmt.Sprintf("/team/create?name=%v", name), nil, true)
	if err != nil {
		return "", err
	}
	return data["team_id"], nil
}

func (c *client) deleteTeam(teamId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/delete?team_id=%v", teamId), nil, false)
	return err
}

func (c *client) addUserToTeam(teamId, userId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/add-user?team_id=%v&user_id=%v", teamId, userId), nil, false)
	return err
}

func (c *client) removeUserFromTeam(teamId, userId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/remove-user?team_id=%v&user_id=%v", teamId, userId), nil, false)
	return err
}

func (c *client) addModelToTeam(teamId, modelId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/add-model?team_id=%v&model_id=%v", teamId, modelId), nil, false)
	return err
}

func (c *client) removeModelFromTeam(teamId, modelId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/remove-model?team_id=%v&model_id=%v", teamId, modelId), nil, false)
	return err
}

func (c *client) addTeamAdmin(teamId, userId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/add-admin?team_id=%v&user_id=%v", teamId, userId), nil, false)
	return err
}

func (c *client) removeTeamAdmin(teamId, userId string) error {
	_, err := post[int](c, fmt.Sprintf("/team/remove-admin?team_id=%v&user_id=%v", teamId, userId), nil, false)
	return err
}

func (c *client) listTeams() ([]services.TeamInfo, error) {
	return get[[]services.TeamInfo](c, "/team/list")
}

func (c *client) listTeamModels(teamId string) ([]services.ModelInfo, error) {
	return get[[]services.ModelInfo](c, fmt.Sprintf("/team/models?team_id=%v", teamId))
}

func (c *client) listTeamUsers(teamId string) ([]services.TeamUserInfo, error) {
	return get[[]services.TeamUserInfo](c, fmt.Sprintf("/team/users?team_id=%v", teamId))
}
