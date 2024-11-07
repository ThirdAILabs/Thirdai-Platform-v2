package nomad

import (
	"fmt"
	"thirdai_platform/model_bazaar/schema"
)

func TrainJobName(model schema.Model) string {
	return fmt.Sprintf("train-%v-%v", model.Id, model.Type)
}

func DeployJobName(model schema.Model) string {
	return fmt.Sprintf("deploy-%v-%v", model.Id, model.Type)
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

type Resources struct {
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

	Driver Driver

	Resources Resources
}

func (j TrainJob) TemplateName() string {
	return "train_job.hcl.tmpl"
}

type DeployJob struct {
	JobName string
	ModelId string

	ConfigPath     string
	DeploymentName string

	AutoscalingEnabled bool
	AutoscalingMax     int

	Driver Driver

	Resources Resources
}

func (j DeployJob) TemplateName() string {
	return "deploy_job.hcl.tmpl"
}
