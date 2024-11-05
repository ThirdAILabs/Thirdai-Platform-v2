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

type Driver interface {
	DriverType() string
}

type DockerDriver struct {
	Registry       string
	ImageName      string
	Tag            string
	DockerUsername string
	DockerPassword string
	ShareDir       string
}

func (p DockerDriver) DriverType() string {
	return "docker"
}

type LocalDriver struct {
	PlatformDir string
	PythonPath  string
}

func (p LocalDriver) DriverType() string {
	return "local"
}

type Resource struct {
	AllocationMhz       int
	AllocationMemory    int
	AllocationMemoryMax int
}

type Job interface {
	TemplateName() string
}

type TrainJob struct {
	JobName string

	ConfigPath string

	DriverType string
	Driver     Driver // Either DockerPlatform or LocalPlatform

	Resources Resource
}

func (j TrainJob) TemplateName() string {
	return "train_job.hcl.tmpl"
}
