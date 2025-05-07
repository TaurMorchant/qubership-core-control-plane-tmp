package format

import (
	"strings"
)

type RouteToAddress struct {
	urlValidator *UrlValidator

	prefixRewrite string
	regexpRewrite string
}

type RouteInfoProvider interface {
	GetTo() string
	GetFrom() string
}

func NewRouteToAddress(routeInfoProvider RouteInfoProvider) *RouteToAddress {
	prefixRewrite := ""
	regexpRewrite := ""
	if routeInfoProvider.GetTo() != "" {
		prefixRewrite, regexpRewrite = rewrite(routeInfoProvider.GetTo())
	} else {
		_, regexpRewrite = rewrite(routeInfoProvider.GetFrom())
	}
	return &RouteToAddress{urlValidator: DefaultUrlValidator, prefixRewrite: prefixRewrite, regexpRewrite: regexpRewrite}
}

func rewrite(to string) (string, string) {
	formatter := DefaultRouteFormatter
	routeToFormatted := formatter.GetRoutePropertyKey(to)

	regexToFormatted := ""
	prefixToFormatted := ""
	if HasVariable(to) {
		regexToFormatted = strings.ReplaceAll(routeToFormatted, "/", "\\/")
		regexToFormatted = oldRegexFormat.ReplaceAllLiteralString(regexToFormatted, "")
		regexToFormatted = backSlashRegexp.ReplaceAllLiteralString(regexToFormatted, ".*")
	} else {
		toPrefix := backSlashRegexp.Split(routeToFormatted, 2)[0]
		prefixToFormatted = replaceEndSlash(toPrefix, "")
	}

	return prefixToFormatted, regexToFormatted
}

func (r *RouteToAddress) GetPrefixRewrite() string {
	return r.prefixRewrite
}

func (r *RouteToAddress) GetRegexpRewrite() string {
	return r.regexpRewrite
}
