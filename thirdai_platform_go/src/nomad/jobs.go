package nomad

import (
	"fmt"
	"thirdai_platform/src/schema"
)

func TrainJobName(model schema.Model) string {
	return fmt.Sprintf("train-%v-%v-%v", model.Id, model.Type, model.Subtype)
}

func DeployJobName(model schema.Model) string {
	return fmt.Sprintf("deploy-%v-%v-%v", model.Id, model.Type, model.Subtype)
}
