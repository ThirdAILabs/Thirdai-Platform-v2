package tests

import (
	"fmt"
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

		if info.Username != username || info.Email != email || info.Id != client.userId || info.Admin {
			t.Fatalf("invalid info %v", info)
		}
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
	if info.Username != adminUsername || info.Email != adminEmail || info.Id != admin.userId || !info.Admin {
		t.Fatalf("invalid admin info %v", info)
	}

	client := env.newClient()
	login, err := client.signup("abc", "abc@mail.com", "abc")
	if err != nil {
		t.Fatal(err)
	}

	info, err = client.userInfo()
	if err != ErrUnauthorized {
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

	if info.Username != "abc" || info.Email != "abc@mail.com" || info.Id != client.userId || info.Admin {
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
		t.Fatal("invalid user1 user list")
	}

	client := env.newClient()
	_, err = client.listUsers()
	if err != ErrUnauthorized {
		t.Fatal("expected unauthorized error")
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
	if err != ErrUnauthorized {
		t.Fatal("users can't promote admins")
	}

	err = user1.promoteAdmin(user2.userId)
	if err != ErrUnauthorized {
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
	if err != ErrUnauthorized {
		t.Fatal("non admin cannot demote admin")
	}

	checkAdminStatus(admin, t, true)
	checkAdminStatus(user1, t, false)
	checkAdminStatus(user2, t, true)
}
