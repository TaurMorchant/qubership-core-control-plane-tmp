package routes

import (
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"regexp"
	"strings"
)

var (
	VariableReplacement       = "*"
	SourceRouteEnding         = "/**"
	DotReplaceSequence        = "&dot;"
	PathVarRegexp             = regexp.MustCompile("\\{[^{}]*\\}")
	predefinedForbiddenRoutes = map[string]bool{
		"": true,
	}
	regexpForbiddenSymbolsRoutes = map[*regexp.Regexp]string{
		regexp.MustCompile("\\[|\\]"): "Brackets are forbidden according to RFC 3986.",
	}
	logger = logging.GetLogger("routes-formatter")
)

func IsRouteAllowed(route string) bool {
	for forbidden, _ := range predefinedForbiddenRoutes {
		if route == forbidden {
			return false
		}
		for regexpForbidden, msg := range regexpForbiddenSymbolsRoutes {
			if regexpForbidden.MatchString(route) {
				logger.Errorf("The request contains a forbidden route: %v. %v", route, msg)
				return false
			}
		}
	}
	return true
}

func getRoutePropertyKey(routeFrom string) string {
	result := escapeRouteFrom(routeFrom)
	return applyPostfix(result)
}

func escapeRouteFrom(route string) string {
	route = PathVarRegexp.ReplaceAllString(route, VariableReplacement)
	return strings.ReplaceAll(route, ".", DotReplaceSequence)
}

func applyPostfix(routeFrom string) string {
	if strings.HasSuffix(routeFrom, SourceRouteEnding) {
		return routeFrom
	}
	return routeFrom + SourceRouteEnding
}
