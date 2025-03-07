package deployment

import (
	"fmt"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/services"
)

type Reporter struct {
	client.BaseClient
	ModelId string
}

func (r *Reporter) UpdateDeployStatusInternal(status string) error {
	return r.Post(fmt.Sprintf("/api/v2/deploy/%v/update-status", r.ModelId)).Do(nil)
}

func (r *Reporter) GetDeployStatusInternal() (services.StatusResponse, error) {
	var res services.StatusResponse
	err := r.Get(fmt.Sprintf("/api/v2/deploy/%v/status-internal", r.ModelId)).Do(&res)
	return res, err
}

// TODO(any): Add log method and integration with victoria logs
