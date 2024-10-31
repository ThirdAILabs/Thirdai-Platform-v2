package tests

import "testing"

func TestSignupAndLogin(t *testing.T) {
	env := setupTestEnv(t)

	const (
		username = "user1"
		email    = "user1@mail.com"
		password = "user1_password"
	)

	client := env.newClient()

	login, err := client.signup(username, email, password)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.signup(username, email, password)
	if err == nil {
		t.Fatal("duplicate signup should fail")
	}
}
