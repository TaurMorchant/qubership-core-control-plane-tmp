package dto

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/creator"
)

type RouteRegistrationRequest struct {
	Cluster    string      `json:"cluster"`
	Routes     []RouteItem `json:"routes"`
	Endpoint   string      `json:"endpoint"`
	Allowed    *bool       `json:"allowed"`
	Namespace  string      `json:"namespace"`
	Version    string      `json:"version"`
	Overridden bool        `json:"overridden"`
}

type RouteItem struct {
	Prefix         string          `json:"prefix"`
	PrefixRewrite  string          `json:"prefixRewrite"`
	Timeout        *int64          `json:"timeout"`
	HeaderMatchers []HeaderMatcher `json:"headerMatchers"`
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

type RouteEntityRequest struct {
	MicroserviceUrl string        `json:"microserviceUrl"`
	Routes          *[]RouteEntry `json:"routes"`
	Allowed         bool          `json:"allowed"`
}

type RouteEntry struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Timeout   int    `json:"timeout"`
}

func (req *RouteEntityRequest) ToRouteEntityRequestModel() *creator.RouteEntityRequest {
	routes := make([]*creator.RouteEntry, 0, len(*req.Routes))
	for _, route := range *req.Routes {
		routes = append(routes, creator.NewRouteEntry(route.From, route.To, route.Namespace, int64(route.Timeout), creator.DefaultIdleTimeoutSpec, []*domain.HeaderMatcher{}))
	}
	routeRequest := creator.NewRouteRequest(req.MicroserviceUrl, req.Allowed)
	return creator.NewRouteEntityRequest(&routeRequest, routes)
}

func (req RouteEntityRequest) String() string {
	return fmt.Sprintf("RouteEntityRequest{microserviceUrl='%s',allowed=%t,routes=%v}", req.MicroserviceUrl, req.Allowed, req.Routes)
}

func (e RouteEntry) String() string {
	return fmt.Sprintf("RouteEntry{from=%s,to=%s,namespace=%s,type=%s,timeout=%d}", e.From, e.To, e.Namespace, e.Type, e.Timeout)
}

func HeaderMatchersToDomain(headerMatchersDto []HeaderMatcher) []*domain.HeaderMatcher {
	headerMatchers := make([]*domain.HeaderMatcher, len(headerMatchersDto))
	for index, headerMatcher := range headerMatchersDto {
		headerMatchers[index] = &domain.HeaderMatcher{
			RangeMatch: domain.RangeMatch{
				Start: headerMatcher.RangeMatch.Start,
				End:   headerMatcher.RangeMatch.End,
			},
			SafeRegexMatch: headerMatcher.SafeRegexMatch,
			ExactMatch:     headerMatcher.ExactMatch,
			PrefixMatch:    headerMatcher.PrefixMatch,
			PresentMatch:   headerMatcher.PresentMatch,
			SuffixMatch:    headerMatcher.SuffixMatch,
			InvertMatch:    headerMatcher.InvertMatch,
			Name:           headerMatcher.Name,
		}
	}
	return headerMatchers
}

type RouteDeleteRequest struct {
	Routes    []RouteDeleteItem `json:"routes"`
	Namespace string            `json:"namespace"`
	Version   string            `json:"version"`
}

type RouteDeleteItem struct {
	Prefix string `json:"prefix"`
}

type EndpointDeleteRequest struct {
	Endpoints []EndpointDeleteItem `json:"endpoints"`
	Version   string               `json:"version"`
}

type EndpointDeleteItem struct {
	Address string `json:"address"`
	Port    int32  `json:"port"`
}

type LoadBalanceSpec struct {
	Cluster    string       `json:"cluster" yaml:"cluster"`
	Version    string       `json:"version" yaml:"version"`
	Endpoint   string       `json:"endpoint" yaml:"endpoint"`
	Namespace  string       `json:"namespace" yaml:"namespace"`
	Policies   []HashPolicy `json:"policies" yaml:"policies"`
	Overridden bool         `json:"overridden"`
}

func (s LoadBalanceSpec) String() string {
	return fmt.Sprintf("LoadBalanceSpec{cluster=%s,version=%s,endpoint='%s',namespace=%s,policies=%v}", s.Cluster, s.Version, s.Endpoint, s.Namespace, s.Policies)
}

type HashPolicy struct {
	Id                   int32                 `json:"id"`
	Header               *Header               `json:"header" yaml:"header"`
	Cookie               *Cookie               `json:"cookie" yaml:"cookie"`
	ConnectionProperties *ConnectionProperties `json:"connectionProperties" yaml:"connectionProperties"`
	QueryParameter       *QueryParameter       `json:"queryParameter" yaml:"queryParameter"`
	Terminal             domain.NullBool       `json:"terminal" yaml:"terminal" swaggertype:"boolean"`
}

func (p HashPolicy) String() string {
	return fmt.Sprintf("HashPolicy{id=%d,header=%v,cookie=%v,connProps=%v,queryParameter=%v,terminal=%v}", p.Id, p.Header, p.Cookie, p.ConnectionProperties, p.QueryParameter, p.Terminal)
}

type Header struct {
	HeaderName string `json:"headerName" yaml:"headerName"`
}

func (h Header) String() string {
	return fmt.Sprintf("Header{name=%s}", h.HeaderName)
}

type Cookie struct {
	Name string            `json:"name" yaml:"name"`
	Ttl  *int64            `json:"ttl" yaml:"ttl"`
	Path domain.NullString `json:"path" swaggertype:"string" yaml:"path"`
}

func (c Cookie) String() string {
	if c.Ttl == nil {
		return fmt.Sprintf("Cookie{name=%s,path=%v}", c.Name, c.Path)
	} else {
		return fmt.Sprintf("Cookie{name=%s,ttl=%d,path=%v}", c.Name, *c.Ttl, c.Path)
	}
}

type ConnectionProperties struct {
	SourceIp domain.NullBool `json:"sourceIp" yaml:"sourceIp" swaggertype:"boolean"`
}

func (p ConnectionProperties) String() string {
	return fmt.Sprintf("ConnProps{sourceIp=%v}", p.SourceIp)
}

type QueryParameter struct {
	Name string `json:"name" yaml:"name"`
}

func (p QueryParameter) String() string {
	return fmt.Sprintf("QueryParameter{name=%s}", p.Name)
}

type ClusterResponse struct {
	Id            int32               `json:"id"`
	Name          string              `json:"name"`
	LbPolicy      string              `json:"lbPolicy"`
	DiscoveryType string              `json:"type"`
	Version       int32               `json:"version"`
	EnableH2      bool                `json:"enableH2"`
	NodeGroups    []*domain.NodeGroup `json:"nodeGroups"`
	Endpoints     []*Endpoint         `json:"endpoints"`
}

type Endpoint struct {
	Id                       int32                     `json:"id"`
	Address                  string                    `json:"address"`
	Port                     int32                     `json:"port"`
	DeploymentVersion        *domain.DeploymentVersion `json:"deploymentVersion"`
	InitialDeploymentVersion string                    `json:"initialDeploymentVersion"`
	HashPolicies             []*HashPolicy             `json:"hashPolicy"`
}

type RouteConfigurationResponse struct {
	Id           int32          `json:"id"`
	Name         string         `json:"name"`
	Version      int32          `json:"version"`
	NodeGroupId  string         `json:"nodeGroup"`
	VirtualHosts []*VirtualHost `json:"virtualHosts"`
}

type VirtualHost struct {
	Id            int32              `json:"id"`
	Name          string             `json:"name"`
	Version       int32              `json:"version"`
	AddHeaders    []HeaderDefinition `json:"addHeaders"`
	RemoveHeaders []string           `json:"removeHeaders"`
	Routes        []*Route           `json:"routes"`
	Domains       []string           `json:"domains"`
}

//swagger:model VirtualServiceResponse
type VirtualServiceResponse struct {
	VirtualHost *VirtualHost       `json:"virtualHost"`
	Clusters    []*ClusterResponse `json:"clusters"`
}

type Route struct {
	Id                       int32                     `json:"id"`
	Uuid                     string                    `json:"uuid"`
	RouteKey                 string                    `json:"routeKey"`
	RouteMatcher             *RouteMatcher             `json:"matcher"`
	RouteAction              *RouteAction              `json:"action"`
	DirectResponseAction     *DirectResponseAction     `json:"directResponseAction"`
	Version                  int32                     `json:"version"`
	Timeout                  domain.NullInt            `json:"timeout" swaggertype:"integer"`
	IdleTimeout              domain.NullInt            `json:"idleTimeout" swaggertype:"integer"`
	DeploymentVersion        *domain.DeploymentVersion `json:"deploymentVersion"`
	InitialDeploymentVersion string                    `json:"initialDeploymentVersion"`
	Autogenerated            bool                      `json:"autoGenerated"`
	HashPolicies             []*HashPolicy             `json:"hashPolicy"`
}

type DirectResponseAction struct {
	Status uint32 `json:"status"`
}

type RouteMatcher struct {
	Prefix         domain.NullString       `json:"prefix" swaggertype:"string"`
	Regexp         domain.NullString       `json:"regExp" swaggertype:"string"`
	Path           domain.NullString       `json:"path" swaggertype:"string"`
	HeaderMatchers []*domain.HeaderMatcher `json:"headers"`
	AddHeaders     []HeaderDefinition      `json:"addHeaders"`
	RemoveHeaders  []string                `json:"removeHeaders"`
}

type RouteAction struct {
	ClusterName     domain.NullString `json:"clusterName" swaggertype:"string"`
	HostRewrite     domain.NullString `json:"hostRewrite" swaggertype:"string"`
	HostAutoRewrite domain.NullBool   `json:"hostAutoRewrite" swaggertype:"boolean"`
	PrefixRewrite   domain.NullString `json:"prefixRewrite" swaggertype:"string"`
	RegexpRewrite   domain.NullString `json:"regexpRewrite" swaggertype:"string"`
	PathRewrite     domain.NullString `json:"pathRewrite" swaggertype:"string"`
}
