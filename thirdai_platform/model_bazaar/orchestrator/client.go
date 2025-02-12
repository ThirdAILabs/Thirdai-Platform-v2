package orchestrator

type JobInfo struct {
	Name   string
	Status string
}

type JobLog struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type ServiceAllocation struct {
	Address string
	AllocID string
	NodeID  string
	Port    int
}

type ServiceInfo struct {
	Name        string
	Allocations []ServiceAllocation
}

type Client interface {
	StartJob(job Job) error

	StopJob(jobName string) error

	JobInfo(jobName string) (JobInfo, error)

	JobLogs(jobName string) ([]JobLog, error)

	ListServices() ([]ServiceInfo, error)

	TotalCpuUsage() (int, error)

	IngressHostname() string
}
