package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"thirdai_platform/orchestrator"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
)

type TelemetryService struct {
	orchestratorClient orchestrator.Client
	variables          Variables
}

func (s *TelemetryService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/deployment-services", s.DeploymentServices)

	return r
}

type scrapeTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func (s *TelemetryService) DeploymentServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.orchestratorClient.ListServices()
	if err != nil {
		http.Error(w, fmt.Sprintf("error listing services: %v", err), http.StatusInternalServerError)
		return
	}

	isLocal := s.variables.BackendDriver.DriverType() == "local"

	targets := make([]scrapeTarget, 0)
	for _, service := range services {
		if !strings.HasPrefix(service.Name, "deploy") {
			continue
		}
		nameParts := strings.SplitN(service.Name, "-", 3)
		if len(nameParts) != 3 {
			slog.Error("invalid service name encountered: " + service.Name)
			http.Error(w, "invalid service name: "+service.Name, http.StatusInternalServerError)
			return
		}
		for _, allocation := range service.Allocations {
			var address string
			if isLocal {
				address = fmt.Sprintf("http://host.docker.internal:%d", allocation.Port)
			} else {
				address = fmt.Sprintf("%v:%d", allocation.Address, allocation.Port)
			}
			targets = append(targets, scrapeTarget{
				Targets: []string{address},
				Labels: map[string]string{
					"model_id": nameParts[2],
					"alloc_id": allocation.AllocID,
					"node_id":  allocation.NodeID,
					"address":  allocation.Address,
				},
			})
		}
	}

	utils.WriteJsonResponse(w, targets)
}
