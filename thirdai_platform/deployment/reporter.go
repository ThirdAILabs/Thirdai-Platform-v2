package deployment

import "thirdai_platform/model_bazaar/services"

type Reporter struct {
	ModelBazaarEndpoint string
	ModelId             string
	JobToken            string
}

func (r *Reporter) UpdateModelStatus(status string) {
	err = client.Post("/deploy/update-status").Auth(r.JobToken).Json(map[string]string{"status": "in_progress"}).Do(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func (r *Reporter) GetModelStatus(status string) {
	var internalStatus services.StatusResponse
	err = client.Get("/deploy/status-internal").Auth(r.JobToken).Do(&internalStatus)
	if err != nil {
		t.Fatal(err)
	}
	return internalStatus
}
