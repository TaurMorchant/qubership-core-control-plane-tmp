package domain

import (
	"net/http"
)

type TraceResponse struct {
	ServiceName string `json:"serviceName"`
	FamilyName  string `json:"familyName"`
	Version     string `json:"version"`
	PodId       string `json:"podId"`

	RequestHost string      `json:"requestHost"`
	ServerHost  string      `json:"serverHost"`
	RemoteAddr  string      `json:"remoteAddr"`
	Path        string      `json:"path"`
	Method      string      `json:"method"`
	Headers     http.Header `json:"headers"`
}
