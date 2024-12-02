package jobs

import (
	"errors"
	"fmt"
	"thirdai_platform/model_bazaar/nomad"
)

func jobExists(client nomad.NomadClient, jobName string) (bool, error) {
	_, err := client.JobInfo(jobName)
	if errors.Is(err, nomad.ErrJobNotFound) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func stopJobIfExists(client nomad.NomadClient, jobName string) error {
	exists, err := jobExists(client, jobName)
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
