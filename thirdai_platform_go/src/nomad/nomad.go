package nomad

type NomadClient interface {
	StartJob(jobTemplate string, args interface{}) error

	StopJob(jobName string) error
}

type NomadHttpClient struct {
	addr string
}

func (c *NomadHttpClient) StartJob(jobTemplate string, args interface{}) error {
	return nil
}
