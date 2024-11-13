package tests

import (
	"thirdai_platform/model_bazaar/nomad"
)

type NomadStub struct {
	activeJobs map[string]string
}

func newNomadStub() *NomadStub {
	return &NomadStub{activeJobs: make(map[string]string)}
}

func (c *NomadStub) StartJob(job nomad.Job) error {
	c.activeJobs[job.GetJobName()] = job.TemplateName()
	return nil
}

func (c *NomadStub) StopJob(jobName string) error {
	delete(c.activeJobs, jobName)
	return nil
}

func (c *NomadStub) JobInfo(jobName string) (nomad.JobInfo, error) {
	if _, active := c.activeJobs[jobName]; active {
		return nomad.JobInfo{Name: jobName, Status: "running"}, nil
	}
	return nomad.JobInfo{Name: jobName, Status: "dead"}, nil
}

func (c *NomadStub) JobLogs(jobName string) ([]nomad.JobLog, error) {
	return []nomad.JobLog{}, nil
}

func (c *NomadStub) TotalCpuUsage() (int, error) {
	return 0, nil
}

func (c *NomadStub) Clear() {
	c.activeJobs = map[string]string{}
}
