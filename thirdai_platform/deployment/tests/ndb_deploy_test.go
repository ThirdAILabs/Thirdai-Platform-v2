package tests

import (
	"net/http"
	"testing"
	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/search/ndb"

	"github.com/google/uuid"
)

// func TestDeployConfig(t *testing.T) {
// 	utils.SaveConfig()
// }

func TestBasicEndpoints(t *testing.T) {
	db, err := ndb.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	err = db.Insert(
		"doc_name_1", "doc_id_1",
		[]string{"test line one", "another test line", "something without that"},
		[]map[string]interface{}{{"thing1": true}, {"thing2": true}, {"thing1": true}},
		nil)
	if err != nil {
		t.Fatal(err)
	}

	deploy_config := config.DeployConfig{ModelId: uuid.New(), }

	router, err := deployment.NewNdbRouter(&deploy_config)

	if err != nil {
		t.Fatal(err)
	}

	r := router.Routes()

	err = http.ListenAndServe(":25000", r)

	if err != nil {
		t.Fatal(err)
	}



}
