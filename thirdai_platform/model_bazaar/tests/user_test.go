package tests

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSignupAndLogin(t *testing.T) {
	env := setupTestEnv(t)

	for i := 0; i < 5; i++ {
		username := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("user%d@mail.com", i)
		password := fmt.Sprintf("user%d_password", i)

		client := env.newClient()
		login, err := client.signup(username, email, password)
		if err != nil {
			t.Fatal(err)
		}

		_, err = client.signup(username, email, password)
		if err == nil {
			t.Fatal("duplicate signup should fail")
		}

		err = client.login(loginInfo{Email: "user@mail.com", Password: password})
		if err == nil {
			t.Fatal("login should fail with wrong email")
		}

		err = client.login(loginInfo{Email: username, Password: "password"})
		if err == nil {
			t.Fatal("login fail with wrong password")
		}

		err = client.login(login)
		if err != nil {
			t.Fatal(err)
		}

		info, err := client.userInfo()
		if err != nil {
			t.Fatal(err)
		}

		if info.Username != username || info.Email != email || info.Id.String() != client.userId || info.Admin {
			t.Fatalf("invalid info %v", info)
		}
	}
}

func TestAddUser(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}
	user, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	client := env.newClient()

	_, err = user.addUser("xyz", "xyz@mail.com", "123")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot add users")
	}

	err = client.login(loginInfo{Email: "xyz@mail.com", Password: "123"})
	if !strings.Contains(err.Error(), "no user found for given email") {
		t.Fatalf("no login should be created: %v", err)
	}

	_, err = admin.addUser("xyz", "xyz@mail.com", "123")
	if err != nil {
		t.Fatal(err)
	}

	err = client.login(loginInfo{Email: "xyz@mail.com", Password: "123"})
	if err != nil {
		t.Fatal("new user should be created")
	}
}

func TestUserInfo(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}
	info, err := admin.userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Username != adminUsername || info.Email != adminEmail || info.Id.String() != admin.userId || !info.Admin {
		t.Fatalf("invalid admin info %v", info)
	}

	client := env.newClient()
	login, err := client.signup("abc", "abc@mail.com", "abc")
	if err != nil {
		t.Fatal(err)
	}

	info, err = client.userInfo()
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("expected unauthorized error")
	}

	err = client.login(login)
	if err != nil {
		t.Fatal(err)
	}

	info, err = client.userInfo()
	if err != nil {
		t.Fatal(err)
	}

	if info.Username != "abc" || info.Email != "abc@mail.com" || info.Id.String() != client.userId || info.Admin {
		t.Fatalf("invalid user info %v", info)
	}
}

func TestListUsers(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	user1, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	_, err = env.newUser("xyz")
	if err != nil {
		t.Fatal(err)
	}

	users, err := admin.listUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 3 {
		t.Fatal("expected 3 users for admin list")
	}
	sortUserList(users)
	if users[0].Username != "abc" || users[1].Username != adminUsername || users[2].Username != "xyz" {
		t.Fatalf("invalid admin user list %v", users)
	}

	users, err = user1.listUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Username != "abc" {
		t.Fatalf("invalid user1 user list: %v", users)
	}

	client := env.newClient()
	_, err = client.listUsers()
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("expected unauthorized error")
	}

	user2, err := env.newUser("qrs")
	if err != nil {
		t.Fatal(err)
	}

	team, err := admin.createTeam("team1")
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team, user2.userId)
	if err != nil {
		t.Fatal(err)
	}

	users, err = user1.listUsers()
	if err != nil {
		t.Fatal(err)
	}
	sortUserList(users)
	if len(users) != 2 || users[0].Username != "abc" || users[1].Username != "qrs" {
		t.Fatalf("invalid user1 user list: %v", users)
	}
}

func checkAdminStatus(c client, t *testing.T, isAdmin bool) {
	info, err := c.userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Admin != isAdmin {
		t.Fatalf("expected IsAdmin to be %v, got %v", isAdmin, info.Admin)
	}
}

func TestPromoteDemoteAdmin(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	user1, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := env.newUser("xyz")
	if err != nil {
		t.Fatal(err)
	}

	err = user1.promoteAdmin(user1.userId)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users can't promote admins")
	}

	err = user1.promoteAdmin(user2.userId)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users can't promote admins")
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, false)
	checkAdminStatus(user2, t, false)

	err = admin.promoteAdmin(user1.userId)
	if err != nil {
		t.Fatalf("admin should be able to promote admin: %v", err)
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, true)
	checkAdminStatus(user2, t, false)

	err = user1.promoteAdmin(user2.userId)
	if err != nil {
		t.Fatal("new admin should be able to promote admin")
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, true)
	checkAdminStatus(user2, t, true)

	err = admin.demoteAdmin(user1.userId)
	if err != nil {
		t.Fatalf("admin should be demoted %v", err)
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, false)
	checkAdminStatus(user2, t, true)

	err = user1.demoteAdmin(user2.userId)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("non admin cannot demote admin")
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, false)
	checkAdminStatus(user2, t, true)
}

func TestDeleteUser(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	user, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	m1, err := user.trainNdbDummyFile("m1")
	if err != nil {
		t.Fatal(err)
	}

	m2, err := user.trainNdbDummyFile("m2")
	if err != nil {
		t.Fatal(err)
	}

	m3, err := user.trainNdbDummyFile("m3")
	if err != nil {
		t.Fatal(err)
	}

	err = user.updateAccess(m1, "public")
	if err != nil {
		t.Fatal(err)
	}

	err = user.updateAccess(m2, "protected")
	if err != nil {
		t.Fatal(err)
	}

	users, err := admin.listUsers()
	if err != nil {
		t.Fatal(err)
	}
	sortUserList(users)
	if len(users) != 2 || users[0].Id.String() != user.userId || users[1].Id.String() != admin.userId {
		t.Fatal("invalid users")
	}

	models, err := admin.listModels()
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models)
	if len(models) != 3 || models[0].ModelId.String() != m1 || models[1].ModelId.String() != m2 || models[2].ModelId.String() != m3 {
		t.Fatal("invalid models")
	}
	for _, m := range models {
		if m.Username != "abc" {
			t.Fatal("invalid model owner")
		}
	}

	err = admin.deleteUser(user.userId)
	if err != nil {
		t.Fatal(err)
	}

	users, err = admin.listUsers()
	if err != nil {
		t.Fatal(err)
	}
	sortUserList(users)
	if len(users) != 1 || users[0].Id.String() != admin.userId {
		t.Fatal("invalid users")
	}

	models, err = admin.listModels()
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models)
	if len(models) != 3 || models[0].ModelId.String() != m1 || models[1].ModelId.String() != m2 || models[2].ModelId.String() != m3 {
		t.Fatal("invalid models")
	}
	for _, m := range models {
		if m.Username != "admin123" {
			t.Fatal("invalid model owner")
		}
	}
}
