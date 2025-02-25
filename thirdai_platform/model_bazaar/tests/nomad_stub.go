package tests

import (
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/orchestrator/nomad"
)

type NomadStub struct {
	activeJobs map[string]string
}

func newNomadStub() *NomadStub {
	return &NomadStub{activeJobs: make(map[string]string)}
}

func (c *NomadStub) StartJob(job orchestrator.Job) error {
	c.activeJobs[job.GetJobName()] = nomad.NomadTemplatePath(job.JobTemplatePath())
	return nil
}

func (c *NomadStub) StopJob(jobName string) error {
	delete(c.activeJobs, jobName)
	return nil
}

func (c *NomadStub) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	if _, active := c.activeJobs[jobName]; active {
		return orchestrator.JobInfo{Name: jobName, Status: "running"}, nil
	}
	return orchestrator.JobInfo{Name: jobName, Status: "dead"}, nil
}

func (c *NomadStub) JobLogs(jobName string) ([]orchestrator.JobLog, error) {
	return []orchestrator.JobLog{}, nil
}

func (c *NomadStub) ListServices() ([]orchestrator.ServiceInfo, error) {
	return []orchestrator.ServiceInfo{}, nil
}

func (c *NomadStub) TotalCpuUsage() (int, error) {
	return 0, nil
}

func (c *NomadStub) Clear() {
	c.activeJobs = map[string]string{}
}

func (c *NomadStub) IngressHostname() string {
	return "ingress.hostname"
}
