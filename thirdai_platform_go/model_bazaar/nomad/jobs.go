package nomad

type Driver interface {
	DriverType() string
}

type DockerEnv struct {
	Registry       string
	DockerUsername string
	DockerPassword string
	ShareDir       string
}

type DockerDriver struct {
	ImageName string
	Tag       string
	DockerEnv
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
	AllocationCores     int
	AllocationMhz       int
	AllocationMemory    int
	AllocationMemoryMax int
}

type CloudCredentials struct {
	AwsAccessKey       string
	AwsAccessSecret    string
	AwsRegionName      string
	AzureAccountName   string
	AzureAccountKey    string
	GcpCredentialsFile string
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

	CloudCredentials CloudCredentials
}

func (j TrainJob) GetJobName() string {
	return j.JobName
}

func (j TrainJob) TemplateName() string {
	return "train.hcl.tmpl"
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

	CloudCredentials CloudCredentials

	JobToken string
}

func (j DeployJob) GetJobName() string {
	return j.JobName
}

func (j DeployJob) TemplateName() string {
	return "deploy.hcl.tmpl"
}

type DatagenTrainJob struct {
	TrainJob

	DatagenConfigPath string
	GenaiKey          string
}

func (j DatagenTrainJob) GetJobName() string {
	return j.JobName
}

func (j DatagenTrainJob) TemplateName() string {
	return "datagen_train.hcl.tmpl"
}

type LlmCacheJob struct {
	ModelBazaarEndpoint string
	LicenseKey          string
	ShareDir            string

	Driver Driver
}

func (j LlmCacheJob) GetJobName() string {
	return "llm-cache"
}

func (j LlmCacheJob) TemplateName() string {
	return "llm_cache.hcl.tmpl"
}

type LlmDispatchJob struct {
	ModelBazaarEndpoint string
	ShareDir            string

	Driver Driver
}

func (j LlmDispatchJob) GetJobName() string {
	return "llm-dispatch"
}

func (j LlmDispatchJob) TemplateName() string {
	return "llm_dispatch.hcl.tmpl"
}

type OnPremLlmGenerationJob struct {
	AutoscalingEnabled bool
	InitialAllocations int
	MinAllocations     int
	MaxAllocations     int

	ModelDir  string
	ModelName string

	Docker DockerEnv

	Resources Resources
}

func (j OnPremLlmGenerationJob) GetJobName() string {
	return "on-prem-llm-generation"
}

func (j OnPremLlmGenerationJob) TemplateName() string {
	return "on_prem_llm_generation.hcl.tmpl"
}

type TelemetryJob struct {
	IsLocal     bool
	TargetCount int

	NomadMonitoringDir string

	AdminUsername string
	AdminEmail    string
	AdminPassword string
	GrafanaDbUrl  string

	ModelBazaarPrivateHost string

	Docker DockerEnv
}

func (j TelemetryJob) GetJobName() string {
	return "telemetry"
}

func (j TelemetryJob) TemplateName() string {
	return "telemetry.hcl.tmpl"
}

type FrontendJob struct {
	OpenaiApiKey           string
	IdentityProvider       string
	KeycloakServerHostname string
	NextAuthSecret         string

	UseSslInLogin bool
	Driver        DockerDriver
}

func (j FrontendJob) GetJobName() string {
	return "thirdai-platform-frontend"
}

func (j FrontendJob) TemplateName() string {
	return "frontend.hcl.tmpl"
}
