package controller

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/configresources"

// swagger:parameters HandlePostRoutesWithNodeGroupV1
type RouteEntityRequestParameter struct {
	// in: body
	// required: true
	RouteEntityRequest dto.RouteEntityRequest `json:"routeEntityRequest"`
}

// swagger:parameters HandlePostRoutesWithNodeGroupV1 HandlePostRoutesWithNodeGroupV2 HandleDeleteRoutesWithNodeGroup HandlePostRoutesWithNodeGroup HandleDeleteRoutesWithNodeGroup HandleCreateVirtualService HandleGetVirtualService HandlePutVirtualService HandleDeleteVirtualService HandleGetHttpFilters
type NodeGroupParameter struct {
	// in: path
	// required: true
	NodeGroup string `json:"nodeGroup"`
}

// swagger:parameters HandleDeleteClusterWithID
type ClusterIdParameter struct {
	// in: path
	// required: true
	ClusterId string `json:"clusterId"`
}

// swagger:parameters HandleDeleteClusterWithID HandlePostPromoteVersion HandleDeleteDeploymentVersionWithIDV2 HandleDeleteDeploymentVersionWithIDV3
type VersionParameter struct {
	// in: path
	// required: true
	Version string `json:"version"`
}

// swagger:parameters HandlePostLoadBalanceV2 HandlePostLoadBalanceV3
type LoadBalanceSpecParameter struct {
	// in: body
	// required: true
	LoadBalanceSpec dto.LoadBalanceSpec `json:"loadBalanceSpec"`
}

// swagger:parameters HandleDeleteEndpointsV2 HandleDeleteEndpointsV3
type EndpointDeleteRequestParameter struct {
	// in: body
	// required: true
	EndpointDeleteRequest []dto.EndpointDeleteRequest `json:"endpointDeleteRequest"`
}

// swagger:parameters HandleDeleteRouteWithUUID
type UuidParameter struct {
	// in: path
	// required: true
	Uuid string `json:"uuid"`
}

// swagger:parameters HandleVersionsWatch HandleVersionsWatchV3 HandleActiveActiveWatch
type Connection struct {
	// in: header
	// required: true
	Connection string `json:"connection"`
}

// swagger:parameters HandlePostRoutesWithNodeGroup
type RouteRegistrationRequest struct {
	// in: body
	// required: true
	RouteRegistrationRequest dto.RouteRegistrationRequest `json:"routeRegistrationRequest"`
}

// swagger:parameters HandleDeleteRoutes HandleDeleteRoutesWithNodeGroup
type RouteDeleteRequestParameter struct {
	// in: body
	// required: true
	RouteDeleteRequest dto.RouteDeleteRequest `json:"routeDeleteRequest"`
}

// swagger:parameters HandleGetMicroserviceVersion
type MicroserviceParameter struct {
	// in: path
	// required: true
	Microservice string `json:"microservice"`
}

// swagger:parameters HandleDeleteRoutes HandleDeleteRateLimit HandlePostRateLimit
type RateLimitParameter struct {
	// in: body
	// required: true
	RateLimit dto.RateLimit `json:"rateLimit"`
}

// swagger:parameters HandlePostRoutingConfig
type RoutingConfigRequestV3Parameter struct {
	// in: body
	// required: true
	RoutingConfigRequestV3 dto.RoutingConfigRequestV3 `json:"routingConfigRequestV3"`
}

// swagger:parameters HandleCreateVirtualService HandleGetVirtualService HandlePutVirtualService HandleDeleteVirtualService
type VirtualServiceNameParameter struct {
	// in: path
	// required: true
	virtualServiceName string `json:"virtualServiceName"`
}

// swagger:parameters HandleDeleteVirtualServiceRoutes
type RouteDeleteRequestV3Parameter struct {
	// in: body
	// required: true
	RouteDeleteRequestV3 dto.RouteDeleteRequestV3 `json:"routeDeleteRequestV3"`
}

// swagger:parameters HandleDeleteVirtualServiceDomains
type DomainDeleteRequestV3Parameter struct {
	// in: body
	// required: true
	DomainDeleteRequestV3 dto.DomainDeleteRequestV3 `json:"domainDeleteRequestV3"`
}

// swagger:parameters HandlePostConfig HandlePostApplyConfig
type ConfigResourceParameter struct {
	// in: body
	// required: true
	ConfigResource configresources.ConfigResource `json:"configResource"`
}

// swagger:parameters HandleGetRoutes
type VirtualHostIdParameter struct {
	// in: query
	// required: true
	VirtualHostId string `json:"virtualHostId"`
}

// swagger:parameters HandleGetRoutes
type VersionIdParameter struct {
	// in: query
	// required: true
	VersionId string `json:"versionId"`
}

// swagger:parameters HandleGetRouteDetails
type RouteUuidParameter struct {
	// in: query
	// required: true
	RouteUuid string `json:"routeUuid"`
}

// swagger:parameters HandleActiveActiveConfigPost
type ActiveDCsV3Parameter struct {
	// in: body
	// required: true
	ActiveDCsV3 dto.ActiveDCsV3 `json:"activeDCsV3"`
}

// swagger:parameters HandleActiveActiveConfigPost HandlePostStatefulSession HandlePutStatefulSession HandleDeleteStatefulSession HandleGetStatefulSessions
type StatefulSessionParameter struct {
	// in: body
	// required: true
	StatefulSession dto.StatefulSession `json:"statefulSession"`
}

// swagger:parameters HandleAddNamespaceToComposite
type NamespaceParameter struct {
	// in: path
	// required: true
	Namespace string `json:"namespace"`
}

// swagger:parameters HandleDeleteRoutesWithNodeGroup
type fromParameter struct {
	// in: query
	// required: false
	From string `json:"from"`
}

// swagger:parameters HandleDeleteRoutesWithNodeGroup
type NamespaceQueryParameter struct {
	// in: query
	// required: false
	Namespace string `json:"namespace"`
}

// swagger:parameters HandlePostPromoteVersionV2 HandlePostPromoteVersionV3
type ArchiveSizeParameter struct {
	// in: query
	// required: false
	ArchiveSize int32 `json:"archiveSize"`
}
