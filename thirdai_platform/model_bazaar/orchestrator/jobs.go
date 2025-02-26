package orchestrator

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

	JobTemplatePath() string
}

type TrainJob struct {
	JobName          string
	ConfigPath       string
	Driver           Driver
	Resources        Resources
	CloudCredentials CloudCredentials
}

func (j TrainJob) GetJobName() string {
	return j.JobName
}

func (j TrainJob) JobTemplatePath() string {
	return "train"
}

type DeployJob struct {
	JobName string
	ModelId string

	ConfigPath     string
	DeploymentName string

	AutoscalingEnabled bool
	AutoscalingMin     int
	AutoscalingMax     int

	Driver Driver

	Resources Resources

	CloudCredentials CloudCredentials

	JobToken string
	IsKE     bool

	IngressHostname string
}

func (j DeployJob) GetJobName() string {
	return j.JobName
}

func (j DeployJob) JobTemplatePath() string {
	return "deploy"
}

type DatagenTrainJob struct {
	TrainJob

	DatagenConfigPath string
	GenaiKey          string
}

func (j DatagenTrainJob) GetJobName() string {
	return j.JobName
}

func (j DatagenTrainJob) JobTemplatePath() string {
	return "datagen_train"
}

type LlmCacheJob struct {
	ModelBazaarEndpoint string
	LicenseKey          string
	ShareDir            string

	Driver Driver

	IngressHostname string
}

func (j LlmCacheJob) GetJobName() string {
	return "llm-cache"
}

func (j LlmCacheJob) JobTemplatePath() string {
	return "llm_cache"
}

type LlmDispatchJob struct {
	ModelBazaarEndpoint string
	ShareDir            string

	Driver Driver

	IngressHostname string
}

func (j LlmDispatchJob) GetJobName() string {
	return "llm-dispatch"
}

func (j LlmDispatchJob) JobTemplatePath() string {
	return "llm_dispatch"
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

	IngressHostname string
}

func (j OnPremLlmGenerationJob) GetJobName() string {
	return "on-prem-llm-generation"
}

func (j OnPremLlmGenerationJob) JobTemplatePath() string {
	return "on_prem_llm_generation"
}

type TelemetryJob struct {
	IsLocal bool

	ClusterMonitoringDir string

	AdminUsername string
	AdminEmail    string
	AdminPassword string
	GrafanaDbUrl  string

	ModelBazaarPrivateHost string

	Docker DockerEnv

	IngressHostname string
}

func (j TelemetryJob) GetJobName() string {
	return "telemetry"
}

func (j TelemetryJob) JobTemplatePath() string {
	return "telemetry"
}

type FrontendJob struct {
	OpenaiApiKey           string
	IdentityProvider       string
	KeycloakServerHostname string
	NextAuthSecret         string

	MajorityCriticalServiceNodes int

	UseSslInLogin bool
	Driver        DockerDriver

	IngressHostname string
}

func (j FrontendJob) GetJobName() string {
	return "thirdai-platform-frontend"
}

func (j FrontendJob) JobTemplatePath() string {
	return "frontend"
}

type SnapshotJob struct {
	ConfigPath string
	ShareDir   string
	DbUri      string

	Driver Driver
}

func (j SnapshotJob) GetJobName() string {
	return "snapshot"
}

func (j SnapshotJob) JobTemplatePath() string {
	return "snapshot"
}
