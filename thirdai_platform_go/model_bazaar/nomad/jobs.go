package nomad

import (
	"fmt"
	"thirdai_platform/model_bazaar/schema"
)

func TrainJobName(model schema.Model) string {
	return fmt.Sprintf("train-%v-%v-%v", model.Id, model.Type, model.Subtype)
}

func DeployJobName(model schema.Model) string {
	return fmt.Sprintf("deploy-%v-%v-%v", model.Id, model.Type, model.Subtype)
}

type DockerPlatform struct {
	Registry       string
	ImageName      string
	Tag            string
	DockerUsername string
	DockerPassword string
	ShareDir       string
}

type LocalPlatform struct {
	PlatformDir string
	PythonPath  string
}

type Resource struct {
	AllocationMhz       int
	AllocationMemory    int
	AllocationMemoryMax int
}

type TrainJob struct {
	JobName string

	TrainScript string
	ConfigPath  string
	PythonPath  string

	PlatformType string
	Platform     interface{} // Either DockerPlatform or LocalPlatform

	Resources Resource
}
