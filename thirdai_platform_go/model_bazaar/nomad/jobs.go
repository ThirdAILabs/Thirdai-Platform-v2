package nomad

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
	GetJobName() string

	TemplateName() string
}

type TrainJob struct {
	JobName string

	ConfigPath string

	Driver Driver

	Resources Resources
}

func (j TrainJob) GetJobName() string {
	return j.JobName
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

func (j DeployJob) GetJobName() string {
	return j.JobName
}

func (j DeployJob) TemplateName() string {
	return "deploy_job.hcl.tmpl"
}
