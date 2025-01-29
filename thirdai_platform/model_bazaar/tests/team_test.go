package tests

import (
	"errors"
	"fmt"
	"testing"
	"thirdai_platform/model_bazaar/schema"
)

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
	if !errors.Is(err, ErrUnauthorized) {
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

	if teams[0].Id.String() != team1 || teams[0].Name != "abc" || teams[1].Id.String() != team2 || teams[1].Name != "xyz" {
		t.Fatal("team info wrong")
	}

	err = admin.deleteTeam(team1)
	if err != nil {
		t.Fatal(err)
	}

	err = user.deleteTeam(team2)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot delete teams")
	}

	teams, err = admin.listTeams()
	if err != nil {
		t.Fatal(err)
	}

	if len(teams) != 1 {
		t.Fatal("expected 2 teams")
	}

	if teams[0].Id.String() != team2 || teams[0].Name != "xyz" {
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
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot add to team")
	}

	info, err := users[0].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 1 || info.Teams[0].TeamId.String() != team1 || info.Teams[0].TeamName != "abc" || info.Teams[0].TeamAdmin {
		t.Fatal("invalid team info")
	}

	info, err = users[1].userInfo()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Teams) != 1 || info.Teams[0].TeamId.String() != team2 || info.Teams[0].TeamName != "xyz" || info.Teams[0].TeamAdmin {
		t.Fatal("invalid team info")
	}

	err = users[1].addUserToTeam(team2, users[0].userId)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot add to team")
	}

	err = admin.addUserToTeam(team2, users[0].userId)
	if err != nil {
		t.Fatal(err)
	}

	err = users[1].removeUserFromTeam(team2, users[0].userId)
	if !errors.Is(err, ErrUnauthorized) {
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
	if len(info.Teams) != 2 || info.Teams[0].TeamId.String() != team1 || info.Teams[0].TeamName != "abc" || info.Teams[0].TeamAdmin ||
		info.Teams[1].TeamId.String() != team2 || info.Teams[1].TeamName != "xyz" || info.Teams[1].TeamAdmin {
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
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("non team admin cannot add users")
	}

	err = admin.addTeamAdmin(team1, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addUserToTeam(team2, user2.userId)
	if !errors.Is(err, ErrUnauthorized) {
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
	if !errors.Is(err, ErrUnauthorized) {
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
	if len(info.Teams) != 1 || info.Teams[0].TeamId.String() != team1 || !info.Teams[0].TeamAdmin {
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
	if len(teams) != 2 || teams[0].Id.String() != team1 || teams[0].Name != "abc" || teams[1].Id.String() != team2 || teams[1].Name != "xyz" {
		t.Fatal("invalid teams list for user 1")
	}

	teams, err = user2.listTeams()
	if err != nil {
		t.Fatal(err)
	}
	if len(teams) != 1 || teams[0].Id.String() != team1 || teams[0].Name != "abc" {
		t.Fatal("invalid teams list for user 2")
	}

	teamUsers, err := user1.listTeamUsers(team1)
	if err != nil {
		t.Fatal(err)
	}
	sortTeamUserList(teamUsers)
	if len(teamUsers) != 2 || teamUsers[0].Username != "123" || teamUsers[0].UserId.String() != user1.userId || !teamUsers[0].TeamAdmin ||
		teamUsers[1].Username != "456" || teamUsers[1].UserId.String() != user2.userId || teamUsers[1].TeamAdmin {
		t.Fatal("invalid team users")
	}

	_, err = user1.listTeamUsers(team2)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("only admins can list team users")
	}
}

func TestTeamModels(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	team, err := admin.createTeam("abc")
	if err != nil {
		t.Fatal(err)
	}

	unusedTeam, err := admin.createTeam("xyz")
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

	err = admin.addUserToTeam(team, user1.userId)
	if err != nil {
		t.Fatal(err)
	}

	err = admin.addUserToTeam(team, user2.userId)
	if err != nil {
		t.Fatal(err)
	}

	model1, err := user1.trainNdbDummyFile("searchy")
	if err != nil {
		t.Fatal(err)
	}

	_, err = user2.modelInfo(model1)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("user shouldn't be able to access another user's model")
	}

	err = user1.addModelToTeam(unusedTeam, model1)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("user cannot add model to a team they are not a member of")
	}

	err = user1.addModelToTeam(team, model1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = user2.modelInfo(model1)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("model access must be updated after adding to team")
	}

	err = user1.updateAccess(model1, schema.Protected)
	if err != nil {
		t.Fatal(err)
	}

	info, err := user2.modelInfo(model1)
	if err != nil {
		t.Fatal(err)
	}
	if info.ModelId.String() != model1 || info.TeamId.String() != team {
		t.Fatalf("wrong model info %v", info)
	}

	models, err := admin.listTeamModels(team)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ModelId.String() != model1 {
		t.Fatalf("wrong team models %v", models)
	}

	model2, err := user2.trainNdbDummyFile("retrievy")
	if err != nil {
		t.Fatal(err)
	}
	err = user1.updateAccess(model2, schema.Protected)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("only model owner can update access")
	}

	err = user2.updateAccess(model2, schema.Protected)
	if err != nil {
		t.Fatal(err)
	}

	err = user2.addModelToTeam(team, model2)
	if err != nil {
		t.Fatal(err)
	}

	models, err = admin.listTeamModels(team)
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models)
	if len(models) != 2 || models[0].ModelId.String() != model2 || models[1].ModelId.String() != model1 {
		t.Fatalf("wrong team models %v", models)
	}

	err = admin.removeModelFromTeam(team, model1)
	if err != nil {
		t.Fatal(err)
	}

	models, err = admin.listTeamModels(team)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ModelId.String() != model2 {
		t.Fatalf("wrong team models %v", models)
	}
}
