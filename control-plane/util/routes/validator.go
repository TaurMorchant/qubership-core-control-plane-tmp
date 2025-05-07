package routes

import (
	"regexp"
	"strings"
)

const (
	PathRegex = "\\s+"
)

func isValidPath(path string) bool {
	if path == "" {
		return false
	}
	matched, _ := regexp.MatchString(PathRegex, path)
	return !matched
}

func IsValidFromUrlPath(path string) bool {
	return isValidPath(getRoutePropertyKey(path))
}

func IsValidToUrlPath(path string) bool {
	return isValidPath(path)
}

func IsListedInForbiddenRoutes(route string) bool {
	return !IsRouteAllowed(strings.Trim(route, " "))
}
