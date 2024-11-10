package tests

import (
	"testing"
	"thirdai_platform/model_bazaar/schema"
)

func TestModelInfo(t *testing.T) {
	env := setupTestEnv(t)

	client, err := env.newUser("abc")
	if err != nil {
		t.Fatal(err)
	}

	model, err := client.trainNdb("test_model")
	if err != nil {
		t.Fatal(err)
	}

	info, err := client.modelInfo(model)
	if err != nil {
		t.Fatal(err)
	}

	if info.ModelId != model || info.ModelName != "test_model" || info.TeamId != nil || info.Type != "ndb" {
		t.Fatalf("invalid model info %v", info)
	}
}

func checkPermissions(c client, t *testing.T, modelId string, read, write, owner bool) {
	perm, err := c.modelPermissions(modelId)
	if err != nil {
		t.Fatal(err)
	}
	if perm.Read != read || perm.Write != write || perm.Owner != owner {
		t.Fatal("incorrect permissions")
	}
}

func TestPublicModelPermissions(t *testing.T) {
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

	model, err := user1.trainNdb("model")
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, false, false, false)

	err = user1.updateAccess(model, schema.Public)
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, true, false, false)

	err = user1.updateDefaultPermission(model, schema.WritePerm)
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, true, true, false)
}

func TestProtectedModelPermissions(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	team, err := admin.createTeam("green")
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

	user3, err := env.newUser("123")
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

	model, err := user1.trainNdb("model")
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, false, false, false)
	checkPermissions(user3, t, model, false, false, false)

	err = user1.updateAccess(model, schema.Protected)
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, false, false, false)
	checkPermissions(user3, t, model, false, false, false)

	err = user1.addModelToTeam(team, model)
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, true, false, false)
	checkPermissions(user3, t, model, false, false, false)

	err = user1.updateDefaultPermission(model, schema.WritePerm)
	if err != nil {
		t.Fatal(err)
	}

	checkPermissions(admin, t, model, true, true, true)
	checkPermissions(user1, t, model, true, true, true)
	checkPermissions(user2, t, model, true, true, false)
	checkPermissions(user3, t, model, false, false, false)
}

func TestListModels(t *testing.T) {
	env := setupTestEnv(t)

	admin, err := env.adminClient()
	if err != nil {
		t.Fatal(err)
	}

	team, err := admin.createTeam("green")
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

	user3, err := env.newUser("123")
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

	model1, err := user1.trainNdb("model1")
	if err != nil {
		t.Fatal(err)
	}

	model2, err := user2.trainNdb("model2")
	if err != nil {
		t.Fatal(err)
	}

	model3, err := user3.trainNdb("model3")
	if err != nil {
		t.Fatal(err)
	}

	err = user1.addModelToTeam(team, model1)
	if err != nil {
		t.Fatal(err)
	}

	err = user1.updateAccess(model1, schema.Protected)
	if err != nil {
		t.Fatal(err)
	}

	models1, err := user2.listModels()
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models1)
	if len(models1) != 2 || models1[0].ModelId != model1 || models1[1].ModelId != model2 {
		t.Fatalf("wrong models returned %v", models1)
	}

	models2, err := user3.listModels()
	if err != nil {
		t.Fatal(err)
	}
	if len(models2) != 1 || models2[0].ModelId != model3 {
		t.Fatalf("wrong models returned %v", models2)
	}

	models3, err := admin.listModels()
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models3)
	if len(models3) != 3 || models3[0].ModelId != model1 || models3[1].ModelId != model2 || models3[2].ModelId != model3 {
		t.Fatalf("wrong models returned %v", models3)
	}

	err = user3.updateAccess(model3, schema.Public)
	if err != nil {
		t.Fatal(err)
	}

	models4, err := user2.listModels()
	if err != nil {
		t.Fatal(err)
	}
	sortModelList(models4)
	if len(models4) != 3 || models4[0].ModelId != model1 || models4[1].ModelId != model2 || models4[2].ModelId != model3 {
		t.Fatalf("wrong models returned %v", models4)
	}
}

func TestDeleteModel(t *testing.T) {
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

	model, err := user1.trainNdb("model")
	if err != nil {
		t.Fatal(err)
	}

	err = user2.deleteModel(model)
	if err != ErrUnauthorized {
		t.Fatal("user cannot delete another user's models")
	}

	models, err := admin.listModels()
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ModelId != model {
		t.Fatal("expected a single model")
	}

	err = user1.deleteModel(model)
	if err != nil {
		t.Fatal(err)
	}

	models, err = admin.listModels()
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 0 {
		t.Fatal("expected no models")
	}
}
