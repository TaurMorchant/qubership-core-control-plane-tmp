package dto

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var logger logging.Logger

var disableIpRouteReg bool
var ipAddressRegexp = regexp.MustCompile("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):?[0-9]*$")
var oobGateways = []string{domain.PublicGateway, domain.PrivateGateway, domain.InternalGateway}

type RoutingV1RequestValidator struct {
}

type RoutingV2RequestValidator struct {
}

type RoutingV3RequestValidator struct {
}

func init() {
	logger = logging.GetLogger("validation")
	reloadIpRouteRegistrationValidation()
}

func reloadIpRouteRegistrationValidation() {
	disableIpRouteReg = strings.EqualFold(os.Getenv("DISABLE_IP_ROUTE_REGISTRATION"), "true")
}

type LBRequestValidator struct {
	dVersionRepository dao.DeploymentVersionRepository
}

func NewLBRequestValidator(DVersionRepository dao.DeploymentVersionRepository) *LBRequestValidator {
	return &LBRequestValidator{dVersionRepository: DVersionRepository}
}

func (r RoutingV1RequestValidator) Validate(request RouteEntityRequest, nodeGroup string) (bool, string) {
	isValidAddress := isValidAddress(request.MicroserviceUrl)
	isIpRouteEnabled := isIpRouteRegistrationEnabledByNodeGroup(request.MicroserviceUrl, nodeGroup)
	if !isValidAddress || !isIpRouteEnabled {
		msg := fmt.Sprintf("Invalid address at RouteEntityRequest: %v and nodeGroup: %v", request.MicroserviceUrl, nodeGroup)
		return false, msg
	}
	isValidRoute := isValidRoute(request.Routes, nodeGroup)
	if !isValidRoute {
		msg := fmt.Sprintf("Route / is forbidden for registration")
		return false, msg
	}
	return true, ""
}

func (r RoutingV2RequestValidator) Validate(requests []RouteRegistrationRequest, nodeGroup string) (bool, string) {
	for _, request := range requests {
		isValidDeploymentVersion := isValidDeploymentVersion(request.Version)
		isValidAddress := isValidAddress(request.Endpoint)
		if !(isValidDeploymentVersion && isValidAddress) {
			msg := fmt.Sprintf("Invalid deployment version or address for cluster: %v at namespace: %v", request.Cluster, request.Namespace)
			return false, msg
		}

		isIpRouteEnabled := isIpRouteRegistrationEnabledByNodeGroup(request.Endpoint, nodeGroup)
		if !isIpRouteEnabled {
			msg := fmt.Sprintf("Registration of routes with ip address is forbidden for cluster: %v at namespace: %v in node group: %v", request.Cluster, request.Namespace, nodeGroup)
			return false, msg
		}
		isValidRoute := isValidRouteItem(request.Routes, nodeGroup)
		if !isValidRoute {
			msg := fmt.Sprintf("Route / is forbidden for registration")
			return false, msg
		}
	}
	return true, ""
}

func (r RoutingV3RequestValidator) Validate(request RoutingConfigRequestV3) (bool, string) {
	if len(request.Gateways) == 0 {
		return false, "Field with gateways is empty"
	}
	if len(request.VirtualServices) == 0 {
		return false, "Field with virtualServices is empty"
	}
	if !r.isValidListenerPort(request) {
		return false, "Public, private and internal gateways do not allow custom listener ports"
	}
	virtualServicesByHost := make(map[string]string)
	for vsIndex, virtualService := range request.VirtualServices {
		if valid, msg := r.validateVirtualServiceInternal(virtualService, request.Gateways, true); !valid {
			return valid, fmt.Sprintf("Incorrect virtual service with index %d. Msg: %s", vsIndex, msg)
		}
		if virtualService.Hosts == nil {
			if valid, msg := r.validateVirtualServiceHostIsUnique(virtualServicesByHost, virtualService.Name, "*"); !valid {
				return valid, fmt.Sprintf("Incorrect virtual service with index %d. Msg: %s", vsIndex, msg)
			}
		}
		for _, host := range virtualService.Hosts {
			if valid, msg := r.validateVirtualServiceHostIsUnique(virtualServicesByHost, virtualService.Name, host); !valid {
				return valid, fmt.Sprintf("Incorrect virtual service with index %d. Msg: %s", vsIndex, msg)
			}
		}
		isValidRouteMatch := isValidRouteMatch(virtualService, request.Gateways)

		if !isValidRouteMatch {
			msg := fmt.Sprintf("Route / is forbidden for registration")
			return true, msg
		}
	}
	return true, ""
}

func (r RoutingV3RequestValidator) isValidListenerPort(request RoutingConfigRequestV3) bool {
	if request.ListenerPort != 0 && request.ListenerPort != 8080 {
		for _, gateway := range request.Gateways {
			if util.SliceContains(oobGateways, gateway) {
				return false
			}
		}
	}
	return true
}

func (r RoutingV3RequestValidator) validateVirtualServiceHostIsUnique(virtualServicesByHost map[string]string, serviceName string, host string) (bool, string) {
	if anotherServiceName, alreadyContains := virtualServicesByHost[host]; alreadyContains {
		if serviceName != anotherServiceName {
			return false, fmt.Sprintf("Services %s and %s contain the same host '%s'", serviceName, anotherServiceName, host)
		}
	} else {
		virtualServicesByHost[host] = serviceName
	}
	return true, ""
}

func (r RoutingV3RequestValidator) validateVirtualServiceInternal(virtualService VirtualService, gateways []string, tlsEndpointSupported bool) (bool, string) {
	if virtualService.Name == "" {
		return false, fmt.Sprintf("Name of virtual service is empty")
	}
	for _, gateway := range gateways {
		for _, allowedName := range oobGateways {
			if gateway == allowedName {
				if virtualService.Name != allowedName {
					return false, fmt.Sprintf("Expected virtual service name: '%s', but got: '%s'", allowedName, virtualService.Name)
				}
				break
			}
		}
	}
	if virtualService.RateLimit != "" {
		for _, gateway := range gateways {
			if util.SliceContains(oobGateways, gateway) {
				return false, fmt.Sprintf("Must not set ratelimit on virtualService level for %s gateway", gateway)
			}
		}
	}
	if !isValidDeploymentVersion(virtualService.RouteConfiguration.Version) {
		return false, fmt.Sprintf("Deployment version of virtual service is invalid")
	}
	isEgressGateway := isEgressGateway(gateways)
	for routeIndex, route := range virtualService.RouteConfiguration.Routes {
		if !tlsEndpointSupported && route.Destination.TlsEndpoint != "" {
			return false, fmt.Sprintf("TlsEndpoint not supported in Route with index %d in virtual service with name %s", routeIndex, virtualService.Name)
		}
		if route.Destination.Endpoint == "" {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has empty endpoint field", routeIndex, virtualService.Name)
		}
		if !isValidAddress(route.Destination.Endpoint) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong endpoint field", routeIndex, virtualService.Name)
		}
		if !isIpRouteRegistrationEnabled(route.Destination.Endpoint, isEgressGateway) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong endpoint field", routeIndex, virtualService.Name)
		}
		if !isValidTlsAddress(route.Destination, isEgressGateway) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong tls endpoint field", routeIndex, virtualService.Name)
		}
		if route.Destination.Cluster == "" {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has empty cluster field", routeIndex, virtualService.Name)
		}
		if !isValidHttpVersion(route.Destination.HttpVersion) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has invalid httpVersion field: %v (allowed values are 1, 2 or null)", routeIndex, virtualService.Name, route.Destination.HttpVersion)
		}
		if len(route.Rules) == 0 {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has empty set of rules", routeIndex, virtualService.Name)
		}
		for indexRule, rule := range route.Rules {
			if rule.Match.Prefix == "" {
				return false, fmt.Sprintf("Rule with index %d in route %d with virtual service with name %s has empty prefix field", indexRule, routeIndex, virtualService.Name)
			}
			if rule.StatefulSession != nil {
				if valid, errMsg := r.ValidateRouteStatefulSession(*rule.StatefulSession); !valid {
					return false, errMsg
				}
			}
		}
	}
	return true, ""
}

func (r RoutingV3RequestValidator) ValidateVirtualService(virtualService VirtualService, gateways []string) (bool, string) {
	return r.validateVirtualServiceInternal(virtualService, gateways, false)
}

func (r RoutingV3RequestValidator) ValidateVirtualServiceUpdate(virtualService VirtualService, nodeGroup string) (bool, string) {
	if len(virtualService.RouteConfiguration.Routes) == 0 {
		return true, ""
	}
	for routeIndex, route := range virtualService.RouteConfiguration.Routes {
		if route.Destination.Endpoint != "" && !isValidAddress(route.Destination.Endpoint) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong endpoint field", routeIndex, virtualService.Name)
		}
		if !isIpRouteRegistrationEnabledByNodeGroup(route.Destination.Endpoint, nodeGroup) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong endpoint field", routeIndex, virtualService.Name)
		}
		if !isValidTlsAddressByNodeGroup(route.Destination, nodeGroup) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has wrong tls endpoint field", routeIndex, virtualService.Name)
		}
		if route.Destination.HttpVersion != nil && !isValidHttpVersion(route.Destination.HttpVersion) {
			return false, fmt.Sprintf("Route with index %d in virtual service with name %s has invalid httpVersion field: %v (allowed values are 1, 2 or null)", routeIndex, virtualService.Name, route.Destination.HttpVersion)
		}
		if len(route.Rules) == 0 {
			return true, ""
		}
		for indexRule, rule := range route.Rules {
			if rule.Match.Prefix == "" {
				return false, fmt.Sprintf("Rule with index %d in route %d with virtual service with name %s has empty prefix field", indexRule, routeIndex, virtualService.Name)
			}
			if rule.StatefulSession != nil {
				if valid, errMsg := r.ValidateRouteStatefulSession(*rule.StatefulSession); !valid {
					return false, errMsg
				}
			}
		}
	}
	return true, ""
}

func (r RoutingV3RequestValidator) ValidateStatefulSession(request StatefulSession) (bool, string) {
	if len(request.Gateways) == 0 {
		return false, "StatefulSession configuration must contain non-empty \"gateways\" list"
	}
	if request.Cluster == "" {
		return false, "StatefulSession configuration must contain non-empty \"cluster\" field"
	}
	if request.Port == nil {
		if request.Hostname != "" {
			return false, "StatefulSession configuration for endpoint hostname must contain non-empty \"port\" field"
		}
	} else {
		if !isValidDeploymentVersion(request.Version) {
			return false, "StatefulSession configuration has invalid \"version\" field value"
		}
	}
	return r.validateCommonStatefulSessionRestrictions(request)
}

func (r RoutingV3RequestValidator) ValidateRouteStatefulSession(request StatefulSession) (bool, string) {
	if len(request.Gateways) > 0 {
		return false, "Route \"statefulSession\" must not contain \"gateways\" field"
	}
	if request.Namespace != "" {
		return false, "Route \"statefulSession\" must not contain \"namespace\" field"
	}
	if request.Hostname != "" {
		return false, "Route \"statefulSession\" must not contain \"hostname\" field"
	}
	if request.Cluster != "" {
		return false, "Route \"statefulSession\" must not contain \"cluster\" field"
	}
	if request.Port != nil {
		return false, "Route \"statefulSession\" must not contain \"port\" field"
	}
	if request.Version != "" {
		return false, "Route \"statefulSession\" must not contain \"version\" field"
	}
	return r.validateCommonStatefulSessionRestrictions(request)
}

func (r RoutingV3RequestValidator) validateCommonStatefulSessionRestrictions(request StatefulSession) (bool, string) {
	if valid, errMsg := r.validateStatefulSessionDisable(request); !valid {
		return false, errMsg
	}
	return r.validateStatefulSessionDeletion(request)
}

func (r RoutingV3RequestValidator) validateStatefulSessionDisable(request StatefulSession) (bool, string) {
	if request.Enabled != nil && !*request.Enabled && request.Cookie != nil {
		return false, "StatefulSession request with `\"enabled\": false` must not contain \"cookie\" field"
	}
	return true, ""
}

func (r RoutingV3RequestValidator) validateStatefulSessionDeletion(request StatefulSession) (bool, string) {
	if request.Cookie == nil && request.Enabled != nil && *request.Enabled {
		return false, "StatefulSession request with empty \"cookie\" cannot contain `\"enabled\": true`"
	}
	return true, ""
}

func (r RoutingV3RequestValidator) ValidateDomainDeleteRequestV3(requests []DomainDeleteRequestV3) (bool, string) {
	for requestIndex, request := range requests {
		if request.VirtualService == "" {
			return false, fmt.Sprintf("Domain deletion request with index %d has empty \"virtualService\" field", requestIndex)
		}
		if request.Gateway == "" {
			return false, fmt.Sprintf("Domain deletion request with index %d has empty \"gateway\" field", requestIndex)
		}
		if len(request.Domains) == 0 {
			return false, fmt.Sprintf("Domain deletion request with index %d has empty \"domains\" field", requestIndex)
		}

		// Forbid deletion of "*" domain on public, private and internal gateways
		for _, forbiddenGateway := range domain.Gateways() {
			if request.VirtualService == forbiddenGateway && request.Gateway == forbiddenGateway {
				for _, reqDomain := range request.Domains {
					if reqDomain == "*" {
						return false, fmt.Sprintf("Domain deletion request with index %d attempts to delete default domain \"*\" for forbidden gateway %s", requestIndex, forbiddenGateway)
					}
				}
			}
		}
	}
	return true, ""
}

func (l LBRequestValidator) Validate(request LoadBalanceSpec) (bool, string) {
	if request.Endpoint == "" {
		return false, fmt.Sprintf("request must contain \"endpoint\" field")
	}
	if !isValidAddress(request.Endpoint) {
		return false, fmt.Sprintf("loadBalancer has wrong \"endpoint\" field")
	}
	if request.Cluster == "" {
		return false, fmt.Sprintf("request must contain \"cluster\" field")
	}
	if request.Policies == nil {
		return false, fmt.Sprintf("request must contain \"Policies\" field")
	}
	if request.Version != "" {
		existingVersion, err := l.dVersionRepository.FindDeploymentVersion(request.Version)
		if err != nil {
			return false, fmt.Sprintf("Failed to load version %v during load balance configuration request validation: %v", request.Version, err)
		}
		if existingVersion == nil {
			return false, fmt.Sprintf("version %v does not exist", request.Version)
		}
	}
	return true, ""
}

func isValidHttpVersion(version *int32) bool {
	return version == nil || *version == 1 || *version == 2
}

func isValidDeploymentVersion(deploymentVersion string) bool {
	if deploymentVersion == "" {
		return true
	}
	versionPattern := "^v[[:digit:]]+$"
	validDVersion, _ := regexp.Match(versionPattern, []byte(deploymentVersion))
	return validDVersion
}

func isValidTlsAddressByNodeGroup(destination RouteDestination, nodeGroup string) bool {
	if tlsmode.GetMode() != tlsmode.Preferred {
		return true
	}
	if destination.TlsEndpoint == "" {
		return true
	}
	isEgressGateway := nodeGroup == cache.EgressGateway
	return isValidAddress(destination.TlsEndpoint) && isIpRouteRegistrationEnabled(destination.TlsEndpoint, isEgressGateway)
}

func isValidTlsAddress(destination RouteDestination, isEgressGateway bool) bool {
	if tlsmode.GetMode() != tlsmode.Preferred {
		return true
	}
	if destination.TlsEndpoint == "" {
		return true
	}

	return isValidAddress(destination.TlsEndpoint) && isIpRouteRegistrationEnabled(destination.TlsEndpoint, isEgressGateway)
}

func isValidAddress(address string) bool {
	hasScheme := strings.Contains(address, "://")
	if !hasScheme {
		address = "https://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	host := u.Hostname()
	hostPattern := "^[[:alnum:].-]+$"
	validHost, _ := regexp.Match(hostPattern, []byte(host))
	if u.Port() != "" {
		_, err := strconv.ParseUint(u.Port(), 10, 16)
		if err != nil {
			return false
		}
	}
	return validHost
}

func isEgressGateway(gateways []string) bool {
	for _, gateway := range gateways {
		if gateway != cache.EgressGateway {
			return false
		}
	}

	return true
}

func isIpRouteRegistrationEnabledByNodeGroup(address string, nodeGroup string) bool {
	isEgressGateway := nodeGroup == cache.EgressGateway
	return isIpRouteRegistrationEnabled(address, isEgressGateway)
}

func isIpRouteRegistrationEnabled(address string, isEgressGateway bool) bool {
	if !disableIpRouteReg || isEgressGateway {
		return true
	}

	return !ipAddressRegexp.MatchString(address)
}

func isValidRoute(routes *[]RouteEntry, nodeGroup string) bool {

	if nodeGroup == "public-gateway-service" || nodeGroup == "private-gateway-service" || nodeGroup == "internal-gateway-service" {
		if routes != nil {
			for _, route := range *routes {
				if route.From == "/" {
					return false
				}
			}
		}
	}
	return true
}
func isValidRouteItem(routes []RouteItem, nodeGroup string) bool {
	if nodeGroup == "public-gateway-service" || nodeGroup == "private-gateway-service" || nodeGroup == "internal-gateway-service" {
		if routes != nil {
			for _, route := range routes {
				if route.Prefix == "/" {
					return false
				}
			}
		}
	}
	return true
}
func isValidRouteMatch(virtualService VirtualService, gateways []string) bool {
	if gateways != nil {
		for _, gateway := range gateways {
			for _, allowedName := range []string{domain.PublicGateway, domain.PrivateGateway, domain.InternalGateway} {
				if gateway == allowedName {
					routeConfig := virtualService.RouteConfiguration
					if routeConfig.Routes != nil {
						for _, route := range routeConfig.Routes {
							if route.Rules != nil {
								for _, rule := range route.Rules {
									ruleMatch := rule.Match
									if ruleMatch.Prefix != "" {
										if ruleMatch.Prefix == "/" {
											return false
										}

									}
								}

							}
						}
					}

				}
			}
		}
	}
	return true
}
