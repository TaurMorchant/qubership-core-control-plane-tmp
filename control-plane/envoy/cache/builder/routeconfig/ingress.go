package routeconfig

import (
	"github.com/golang/protobuf/ptypes/any"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"os"
)

func NewIngressVirtualHostBuilder(dao dao.Repository, entityService entity.ServiceInterface, routeBuilder RouteBuilder) *GatewayVirtualHostBuilder {
	origins := "*"
	if value, exists := os.LookupEnv("GATEWAYS_ALLOWED_ORIGIN"); exists {
		origins = value
	}
	allowedHeaders := "*"
	if value, exists := os.LookupEnv("GATEWAYS_ALLOWED_HEADERS"); exists {
		allowedHeaders = value
	}
	maxAge := "-1"
	if value, exists := os.LookupEnv("GATEWAYS_ACCESS_CONTROL_MAX_AGE"); exists {
		maxAge = value
	}

	builder := GatewayVirtualHostBuilder{dao: dao, routeBuilder: routeBuilder,
		allowedHeaders: allowedHeaders, maxAge: maxAge, originStringMatchers: convertOrigins(origins),
		builderExt: &ingressVirtualHostBuilderExt{dao: dao, entityService: entityService},
	}
	return &builder
}

type ingressVirtualHostBuilderExt struct {
	dao           dao.Repository
	entityService entity.ServiceInterface
}

func (i *ingressVirtualHostBuilderExt) BuildExtAuthzPerRoute(virtualHost *domain.VirtualHost) (*any.Any, error) {
	return buildCustomExtAuthzPerRoute(i.dao, i.entityService, virtualHost)
}

func (i *ingressVirtualHostBuilderExt) EnrichHeadersToRemove(headersToRemove []string) []string {
	return headersToRemove
}
