package nomad

type Client struct {
	addr string
}

func (c *Client) StartJob(jobTemplate string, args interface{}) error {
	return nil
}
