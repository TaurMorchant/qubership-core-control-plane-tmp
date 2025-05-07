package format

import (
	"regexp"
	"strings"
)

var (
	backSlashRegexp    = regexp.MustCompile("\\*")
	endSlashRegexp     = regexp.MustCompile("/+$")
	endSlashLineRegexp = regexp.MustCompile("^/+$")
	oldRegexFormat     = regexp.MustCompile("\\\\/\\*\\*")
)

type RouteFromAddress struct {
	routeFormatter *RouteFormatter
	urlValidator   *UrlValidator

	RouteFromRaw       string
	RouteFromFormatted string
	RouteFromPrefix    string
	RouteFromRegex     string
}

func NewRouteFromAddress(routeFromRaw string) *RouteFromAddress {
	return NewRouteFromAddress2(routeFromRaw, DefaultRouteFormatter, DefaultUrlValidator)
}

func NewRouteFromAddress2(routeFromRaw string, routeFormatter *RouteFormatter, urlValidator *UrlValidator) *RouteFromAddress {
	this := RouteFromAddress{routeFormatter: routeFormatter, urlValidator: urlValidator, RouteFromRaw: strings.TrimSpace(routeFromRaw)}
	this.RouteFromFormatted = routeFormatter.GetRoutePropertyKey(this.RouteFromRaw)

	if HasVariable(this.RouteFromRaw) {
		logger.Debugf("Regexp format started from %s", routeFromRaw)
		this.RouteFromRegex = strings.ReplaceAll(this.RouteFromFormatted, "/", "\\/")
		this.RouteFromRegex = oldRegexFormat.ReplaceAllLiteralString(this.RouteFromRegex, "")
		this.RouteFromRegex = backSlashRegexp.ReplaceAllLiteralString(this.RouteFromRegex, ".*")
		this.RouteFromRegex = strings.ReplaceAll(this.RouteFromRegex, "\\", "")
		this.RouteFromRegex = strings.ReplaceAll(this.RouteFromRegex, ".*", "([^/]+)")

		// add regex group to match any path ending so route regex matcher can also act like prefix matcher
		for strings.HasSuffix(this.RouteFromRegex, "/") { // remove trailing slash
			this.RouteFromRegex = this.RouteFromRegex[:len(this.RouteFromRegex)-1]
		}
		this.RouteFromRegex = this.RouteFromRegex + "(/.*)?"
		logger.Debugf("Regexp format from %s finished. Result: %s", routeFromRaw, this.RouteFromRegex)
	} else {
		logger.Debugf("Prefix format start for %s", routeFromRaw)
		fromPrefix := backSlashRegexp.Split(this.RouteFromFormatted, 2)[0]
		this.RouteFromPrefix = replaceEndSlash(fromPrefix, "")
		logger.Debugf("Prefix format ends from %s finished. Result: %s", routeFromRaw, this.RouteFromPrefix)
	}

	return &this
}

func replaceEndSlash(src, repl string) string {
	if len(src) == 1 {
		return src
	}
	if endSlashLineRegexp.MatchString(src) {
		return "/"
	}
	return endSlashRegexp.ReplaceAllLiteralString(src, repl)
}

func (r *RouteFromAddress) IsListedInForbiddenRoutes() bool {
	return !r.routeFormatter.IsRouteAllowed(r.RouteFromRaw)
}

func (r *RouteFromAddress) IsValidUrlPath() bool {
	return r.urlValidator.IsValidPath(r.RouteFromFormatted)
}
