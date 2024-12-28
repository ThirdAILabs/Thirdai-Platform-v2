package nomad

import (
	"errors"
	"fmt"
)

func JobExists(client NomadClient, jobName string) (bool, error) {
	_, err := client.JobInfo(jobName)
	if errors.Is(err, ErrJobNotFound) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func StopJobIfExists(client NomadClient, jobName string) error {
	exists, err := JobExists(client, jobName)
	if err != nil {
		return fmt.Errorf("error checking if job %v exists: %w", jobName, err)
	}

	if exists {
		err := client.StopJob(jobName)
		if err != nil {
			return fmt.Errorf("error stopping job: %w", err)
		}
	}

	return nil
}
