package ui

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"time"
)

type DeploymentVersion struct {
	Name  string `json:"name"`
	Stage string `json:"stage"`
}

type VirtualHost struct {
	Id                 int32               `json:"id"`
	Name               string              `json:"name"`
	DeploymentVersions []DeploymentVersion `json:"deploymentVersions"`
}

type SimplifiedRouteConfig struct {
	NodeGroup    string        `json:"nodeGroup"`
	VirtualHosts []VirtualHost `json:"virtualHosts"`
}

type Route struct {
	Uuid         string `json:"uuid"`
	Match        string `json:"match"`
	MatchRewrite string `json:"matchRewrite"`
	ClusterName  string `json:"clusterName"`
	Allowed      bool   `json:"allowed"`
}

type PageRoutes struct {
	NodeGroup       string  `json:"nodeGroup"`
	VirtualHostName string  `json:"virtualHostName"`
	VersionName     string  `json:"versionName"`
	VersionStage    string  `json:"versionStage"`
	TotalCount      int     `json:"totalCount"`
	Routes          []Route `json:"routes"`
}

type NodeGroup struct {
	Name string `json:"name"`
}

type Cluster struct {
	Id                 int32                   `json:"id"`
	Name               string                  `json:"name"`
	EnableH2           bool                    `json:"enableH2"`
	NodeGroups         []NodeGroup             `json:"nodeGroups"`
	DeploymentVersions []DeploymentVersionsExt `json:"deploymentVersions"`
}

type Endpoint struct {
	Id      int32  `json:"id"`
	Address string `json:"address"`
	Port    int32  `json:"port"`
}

type BalancingPolicy struct {
	Id                   int32                 `json:"id"`
	Header               *Header               `json:"header,omitempty"`
	Cookie               *Cookie               `json:"cookie,omitempty"`
	ConnectionProperties *ConnectionProperties `json:"connectionProperties,omitempty"`
	QueryParameter       *QueryParameter       `json:"queryParameter,omitempty"`
	Terminal             domain.NullBool       `json:"terminal,omitempty" swaggertype:"boolean"`
}

type Header struct {
	HeaderName string `json:"headerName" yaml:"headerName"`
}

type Cookie struct {
	Name string            `json:"name" yaml:"name"`
	Ttl  *int64            `json:"ttl" yaml:"ttl"`
	Path domain.NullString `json:"path" yaml:"path" swaggertype:"string"`
}

type ConnectionProperties struct {
	SourceIp domain.NullBool `json:"sourceIp" yaml:"sourceIp" swaggertype:"boolean"`
}

type QueryParameter struct {
	Name string `json:"name" yaml:"name"`
}

type DeploymentVersionsExt struct {
	Name              string            `json:"name"`
	Stage             string            `json:"stage"`
	Endpoint          Endpoint          `json:"endpoint"`
	BalancingPolicies []BalancingPolicy `json:"balancingPolicies"`
}

type RouteDetails struct {
	Path                   string          `json:"path"`
	PathRewrite            string          `json:"pathRewrite"`
	Prefix                 string          `json:"prefix"`
	PrefixRewrite          string          `json:"prefixRewrite"`
	Regexp                 string          `json:"regexp"`
	RegexpRewrite          string          `json:"regexpRewrite"`
	ClusterName            string          `json:"clusterName"`
	HostRewrite            string          `json:"hostRewrite"`
	HostAutoRewrite        *bool           `json:"hostAutoRewrite"`
	DirectResponse         uint32          `json:"directResponse"`
	Endpoint               string          `json:"endpoint"`
	Timeout                *int64          `json:"timeout"`
	IdleTimeout            *int64          `json:"idleTimeout"`
	HeaderMatchers         []HeaderMatcher `json:"headerMatchers"`
	RequestHeadersToAdd    []HeaderToAdd   `json:"requestHeadersToAdd"`
	RequestHeadersToRemove []string        `json:"requestHeadersToRemove"`
}

type HeaderMatcher struct {
	Name           string          `json:"name"`
	ExactMatch     string          `json:"exactMatch"`
	SafeRegexMatch string          `json:"safeRegexMatch"`
	RangeMatch     RangeMatch      `json:"rangeMatch"`
	PresentMatch   domain.NullBool `json:"presentMatch" swaggertype:"boolean"`
	PrefixMatch    string          `json:"prefixMatch"`
	SuffixMatch    string          `json:"suffixMatch"`
	InvertMatch    bool            `json:"invertMatch"`
}

type RangeMatch struct {
	Start domain.NullInt `json:"start" swaggertype:"integer"`
	End   domain.NullInt `json:"end" swaggertype:"integer"`
}

type HeaderToAdd struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CertificateDetails struct {
	CertificateCommonName       string    `json:"certificateCommonName,omitempty"`
	CertificateId               string    `json:"certificateId,omitempty"`
	Valid                       bool      `json:"valid" swaggertype:"boolean"`
	ValidFrom                   time.Time `json:"validFrom"`
	ValidTill                   time.Time `json:"validTill"`
	DaysTillExpiry              int       `json:"daysTillExpiry" swaggertype:"integer"`
	SANs                        []string  `json:"sans,omitempty"`
	IssuerCertificateCommonName string    `json:"issuerCertificateCommonName,omitempty"`
	IssuerCertificateId         string    `json:"issuerCertificateId,omitempty"`
	Reason                      string    `json:"reason,omitempty"`
}

type TlsDefDetails struct {
	Name         string                `json:"name"`
	UsedIn       []*CertificateUsedIn  `json:"usedIn,omitempty"`
	Endpoints    []string              `json:"endpoints,omitempty"`
	Certificates []*CertificateDetails `json:"certificates,omitempty"`
}

type CertificateUsedIn struct {
	Clusters []*CertificateUsedInCluster `json:"clusters,omitempty"`
	Gateway  string                      `json:"gateway,omitempty"`
}

type CertificateUsedInCluster struct {
	Name      string   `json:"name,omitempty"`
	Endpoints []string `json:"endpoints,omitempty"`
}

type CertificateDetailsResponse struct {
	TlsDefDetails []*TlsDefDetails `json:"details"`
}
