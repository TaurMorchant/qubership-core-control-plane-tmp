package route

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile("\\{[^{}]*\\}")
var sourceRouteEnding = "/**"

func EscapeRouteFrom(route string) string {
	return strings.ReplaceAll(variablePattern.ReplaceAllString(route, "*"), ".", "&dot;")
}

func applyPostfix(routeFrom string) string {
	if strings.HasSuffix(routeFrom, sourceRouteEnding) {
		return routeFrom
	} else {
		return routeFrom + sourceRouteEnding
	}
}

func MakeRoutePropertyKey(routeFrom string) string {
	return applyPostfix(EscapeRouteFrom(routeFrom))
}

func IsRouteBelongsToNamespace(route *domain.Route, namespace *msaddr.Namespace) bool {
	for _, hm := range route.HeaderMatchers {
		if hm.Name == "namespace" && hm.ExactMatch == namespace.Namespace {
			return true
		}
	}

	return false
}

func IsDefaultNamespaceRoute(route *domain.Route) bool {
	for _, hm := range route.HeaderMatchers {
		if hm.Name == "namespace" {
			return false
		}
	}

	return true
}

func ValidateMetadataStringField(md configresources.Metadata, field string) (bool, string) {
	rawValue, ok := md[field]
	if !ok {
		return false, fmt.Sprintf("field '%s' must be set", field)
	}
	if value, ok := rawValue.(string); !ok {
		return false, fmt.Sprintf("field '%s' must have string type", field)
	} else if strings.TrimSpace(value) == "" {
		return false, fmt.Sprintf("field '%s' must not be empty", field)
	}
	return true, ""
}
