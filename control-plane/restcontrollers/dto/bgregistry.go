package dto

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/util"
	"strings"
)

type VersionInRegistry struct {
	Version  string         `json:"version"`
	Stage    string         `json:"stage"`
	Clusters []Microservice `json:"clusters"`
}

type Microservice struct {
	Cluster   string   `json:"cluster"`
	Namespace string   `json:"namespace"`
	Endpoints []string `json:"endpoints"`
}

type ServicesVersionPayload struct {
	Namespace string `json:"namespace"`
	// Services contains family names of the microservices being deployed
	Services []string `json:"services"`
	// Version holds blue-green version of microservices being registered (or deleted).
	// Empty version field means that deploy is performed in rolling mode => ACTIVE version should be used.
	Version string `json:"version"`
	// Exists indicates that it is CREATE request payload (not DELETE request). Field is optional and by default
	// request considered to be CREATE. Only explicit `false` value of this field will cause request to be
	// interpreted as DELETE.
	Exists     *bool `json:"exists"`
	Overridden bool  `json:"overridden"`
}

func (p ServicesVersionPayload) Validate() (bool, string) {
	p.Namespace = strings.TrimSpace(p.Namespace)
	if p.Namespace != "" {
		if err := util.IsDNS1123Label(p.Namespace); err != nil {
			return false, fmt.Sprintf("invalid \"namespace\": %s", err.Error())
		}
	}
	if len(p.Services) == 0 {
		return false, "field \"services\" must contain one or more non-empty service names"
	}
	for _, service := range p.Services {
		if err := util.IsDNS1123Label(service); err != nil {
			return false, fmt.Sprintf("invalid service name \"%s\": %s", service, err.Error())
		}
	}
	if p.Version != "" {
		if !isValidDeploymentVersion(p.Version) {
			return false, fmt.Sprintf("invalid \"version\" value: \"%s\"", p.Version)
		}
	}
	return true, ""
}
