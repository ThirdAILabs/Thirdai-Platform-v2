package routers

import "thirdai_platform/model_bazaar/nomad"

type Variables struct {
	Driver nomad.Driver

	ModelBazaarEndpoint string
}
