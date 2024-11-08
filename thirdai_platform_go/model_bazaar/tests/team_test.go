package tests

import (
	"fmt"
	"slices"
	"testing"
	"thirdai_platform/model_bazaar/services"
)

func sortTeamList(users []services.TeamInfo) {
	slices.SortFunc(users, func(a, b services.TeamInfo) int {
		if a.Name == b.Name {
			return 0
		}
		if a.Name < b.Name {
			return -1
		}
		return 1
	})
}

func sortUserTeamList(users []services.UserTeamInfo) {
	slices.SortFunc(users, func(a, b services.UserTeamInfo) int {
		if a.TeamName == b.TeamName {
			return 0
		}
		if a.TeamName < b.TeamName {
			return -1
		}
		return 1
	})
}

func sortTeamUserList(users []services.TeamUserInfo) {
	slices.SortFunc(users, func(a, b services.TeamUserInfo) int {
		if a.Username == b.Username {
			return 0
		}
		if a.Username < b.Username {
			return -1
		}
		return 1
	})
}

func TestCreateDeleteTeams(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	user, err := env.newUser("123")
	if err != nil {
		t.Fatal(err)
	}

	_, err = user.createTeam("000")
	if err != ErrUnauthorized {
		t.Fatal("users cannot create teams")
	}

	team1, err := admin.createTeam("abc")
	if err != nil {
		t.Fatal(err)
	}

	team2, err := admin.createTeam("xyz")
	if err != nil {
		t.Fatal(err)
	}

	teams, err := admin.listTeams()
	if err != nil {
		t.Fatal(err)
	}

	if len(teams) != 2 {
		t.Fatal("expected 2 teams")
	}

	if teams[0].Id != team1 || teams[0].Name != "abc" || teams[1].Id != team2 || teams[1].Name != "xyz" {
		t.Fatal("team info wrong")
	}

	err = admin.deleteTeam(team1)
	if err != nil {
		t.Fatal(err)
	}

	err = user.deleteTeam(team2)
	if err != ErrUnauthorized {
		t.Fatal("users cannot delete teams")
	}

	teams, err = admin.listTeams()
	if err != nil {
		t.Fatal(err)
	}

	if len(teams) != 1 {
		t.Fatal("expected 2 teams")
	}

	if teams[0].Id != team2 || teams[0].Name != "xyz" {
		t.Fatal("team info wrong")
	}
}

func TestAddRemoveTeamUsers(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	users := make([]client, 0)
	for i := 0; i < 3; i++ {
		user, err := env.newUser(fmt.Sprintf("%d%d%d", i, i, i))
		if err != nil {
			t.Fatal(err)
		}
		users = append(users, user)
	}

	team1, err := admin.createTeam("abc")
	if err != nil {
		t.Fatal(err)
	}

	team2, err := admin.createTeam("xyz")
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team1, users[0].userId)
	if err != nil {
		t.Fatal(err)
	}
	err = admin.addUserToTeam(team2, users[1].userId)
	if err != nil {
		t.Fatal(err)
	}

	err = users[1].addUserToTeam(team2, users[0].userId)
	if err != ErrUnauthorized {
		t.Fatal("users cannot add to team")
	}

	info, err := users[0].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 1 || info.Teams[0].TeamId != team1 || info.Teams[0].TeamName != "abc" || info.Teams[0].TeamAdmin {
		t.Fatal("invalid team info")
	}

	info, err = users[1].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 1 || info.Teams[0].TeamId != team2 || info.Teams[0].TeamName != "xyz" || info.Teams[0].TeamAdmin {
		t.Fatal("invalid team info")
	}

	err = users[1].addUserToTeam(team2, users[0].userId)
	if err != ErrUnauthorized {
		t.Fatal("users cannot add to team")
	}

	err = admin.addUserToTeam(team2, users[0].userId)
	if err != nil {
		t.Fatal(err)
	}

	err = users[1].removeUserFromTeam(team2, users[0].userId)
	if err != ErrUnauthorized {
		t.Fatal(err)
	}

	err = admin.removeUserFromTeam(team2, users[1].userId)
	if err != nil {
		t.Fatal(err)
	}

	info, err = users[0].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	sortUserTeamList(info.Teams)
	if len(info.Teams) != 2 || info.Teams[0].TeamId != team1 || info.Teams[0].TeamName != "abc" || info.Teams[0].TeamAdmin ||
		info.Teams[1].TeamId != team2 || info.Teams[1].TeamName != "xyz" || info.Teams[1].TeamAdmin {
		t.Fatal("invalid team info")
	}

	info, err = users[1].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 0 {
		t.Fatal("invalid team info")
	}
}

func TestTeamAdmins(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	team1, err := admin.createTeam("abc")
	if err != nil {
		t.Fatal(err)
	}

	team2, err := admin.createTeam("xyz")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := env.newUser("123")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := env.newUser("456")
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addUserToTeam(team1, user2.userId)
	if err != ErrUnauthorized {
		t.Fatal("non team admin cannot add users")
	}

	err = admin.addTeamAdmin(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addUserToTeam(team2, user2.userId)
	if err != ErrUnauthorized {
		t.Fatal("team admin cannot add users to other teams")
	}

	err = user1.addUserToTeam(team1, user2.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addTeamAdmin(team1, user2.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user2.removeTeamAdmin(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.removeUserFromTeam(team1, user2.userId)
	if err != ErrUnauthorized {
		t.Fatal("non admin cannot remove team users")
	}

	err = user2.removeUserFromTeam(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	info, err := user1.userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 0 {
		t.Fatal("user1 should not be in teams")
	}

	info, err = user2.userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 1 || info.Teams[0].TeamId != team1 || !info.Teams[0].TeamAdmin {
		t.Fatal("user2 should be admin on 1 team")
	}
}

func TestListTeamsAndTeamUsers(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	team1, err := admin.createTeam("abc")
	if err != nil {
		t.Fatal(err)
	}

	team2, err := admin.createTeam("xyz")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := env.newUser("123")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := env.newUser("456")
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addTeamAdmin(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addUserToTeam(team1, user2.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team2, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	teams, err := user1.listTeams()
	if err != nil {
		t.Fatal(err)
	}
	sortTeamList(teams)
	if len(teams) != 2 || teams[0].Id != team1 || teams[0].Name != "abc" || teams[1].Id != team2 || teams[1].Name != "xyz" {
		t.Fatal("invalid teams list for user 1")
	}

	teams, err = user2.listTeams()
	if len(teams) != 1 || teams[0].Id != team1 || teams[0].Name != "abc" {
		t.Fatal("invalid teams list for user 2")
	}

	teamUsers, err := user1.listTeamUsers(team1)
	if err != nil {
		t.Fatal(err)
	}
	sortTeamUserList(teamUsers)
	if len(teamUsers) != 2 || teamUsers[0].Username != "123" || teamUsers[0].UserId != user1.userId || !teamUsers[0].TeamAdmin ||
		teamUsers[1].Username != "456" || teamUsers[1].UserId != user2.userId || teamUsers[1].TeamAdmin {
		t.Fatal("invalid team users")
	}

	_, err = user1.listTeamUsers(team2)
	if err != ErrUnauthorized {
		t.Fatal("only admins can list team users")
	}
}
