package format

import (
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"regexp"
	"strings"
)

var logger logging.Logger

func init() {
	logger = logging.GetLogger("route-formatter")
}

type RouteFormatter struct {
	variableRegexp *regexp.Regexp
}

func NewRouteFormatter(variablePattern string) *RouteFormatter {
	varRegexp, err := regexp.Compile(variablePattern)
	if err != nil {
		logger.Panicf("Failed to build new RouteFormatter due to variableRegexp pattern compilation error: %v", err)
	}
	return &RouteFormatter{variableRegexp: varRegexp}
}

var DefaultRouteFormatter = NewRouteFormatter("\\{[^{}]*\\}")

const VariableReplacement = "*"
const SourceRouteEnding = "/**"

var PredefinedForbiddenRoutes = []string{""}

func (formatter *RouteFormatter) IsRouteAllowed(route string) bool {
	for _, predefinedRoute := range PredefinedForbiddenRoutes {
		if route == predefinedRoute {
			return false
		}
	}
	return true
}

func (formatter *RouteFormatter) escapeRouteFrom(route string) string {
	return formatter.variableRegexp.ReplaceAllLiteralString(route, VariableReplacement)
}

func (formatter *RouteFormatter) applyPostfix(routeFrom string) string {
	if strings.HasSuffix(routeFrom, SourceRouteEnding) {
		return routeFrom
	} else {
		return routeFrom + SourceRouteEnding
	}
}

func (formatter *RouteFormatter) GetRoutePropertyKey(routeFrom string) string {
	return formatter.applyPostfix(formatter.escapeRouteFrom(routeFrom))
}

func (formatter *RouteFormatter) GetFromWithoutVariable(from string) string {
	firstBracketIndex := strings.Index(from, "{")
	if firstBracketIndex == -1 {
		return from
	} else {
		return from[:firstBracketIndex-1]
	}
}

func HasVariable(str string) bool {
	return strings.Contains(str, "{") && strings.Contains(str, "}")
}
