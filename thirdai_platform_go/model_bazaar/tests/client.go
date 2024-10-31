package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"thirdai_platform/model_bazaar/routers"

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

	req := httptest.NewRequest("POST", "/user/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		return loginInfo{}, fmt.Errorf("singup failed")
	}

	var data map[string]string
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return loginInfo{}, jsonError(err)
	}

	return loginInfo{Email: email, Password: password}, nil
}

func (c *client) login(login loginInfo) error {
	body, err := json.Marshal(login)
	if err != nil {
		return jsonError(err)
	}

	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed")
	}

	var data map[string]string
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return jsonError(err)
	}

	c.token = data["access_token"]
	c.userId = data["user_id"]

	return nil
}

func (c *client) promoteAdmin(userId string) error {
	body := []byte(fmt.Sprintf(`{"user_id": "%v"}`, userId))
	req := httptest.NewRequest("POST", "/user/promote-admin", bytes.NewReader(body))
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return fmt.Errorf("promote admin failed")
	}

	return nil
}

func (c *client) demoteAdmin(userId string) error {
	body := []byte(fmt.Sprintf(`{"user_id": "%v"}`, userId))
	req := httptest.NewRequest("POST", "/user/demote-admin", bytes.NewReader(body))
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return fmt.Errorf("demote admin failed: %v", w.Body.String())
	}

	return nil
}

func (c *client) listUsers() ([]routers.UserInfo, error) {
	req := httptest.NewRequest("GET", "/user/list", nil)
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("list users failed")
	}

	var data []routers.UserInfo
	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, jsonError(err)
	}

	return data, nil
}

func (c *client) userInfo() (routers.UserInfo, error) {
	req := httptest.NewRequest("GET", "/user/info", nil)
	c.addAuthHeader(req)
	w := httptest.NewRecorder()
	c.api.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return routers.UserInfo{}, ErrUnauthorized
		}
		return routers.UserInfo{}, fmt.Errorf("get user info failed %v", w.Body.String())
	}

	var data routers.UserInfo
	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return routers.UserInfo{}, jsonError(err)
	}

	return data, nil
}
