package routekey

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func GenerateKey(route domain.Route) string {
	routeKey := fromRoute(route)
	return Generate(routeKey)
}

func GenerateNoVersionKey(route domain.Route) string {
	routeKey := fromRoute(route)
	routeKey.Version = ""
	return Generate(routeKey)
}

func Generate(routeMatch RouteMatch) string {
	b, _ := json.Marshal(routeMatch)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func GenerateFunc(rmProvider func() RouteMatch) string {
	return Generate(rmProvider())
}

type RouteMatch struct {
	Prefix  string        `json:",omitempty"`
	Regexp  string        `json:",omitempty"`
	Path    string        `json:",omitempty"`
	Headers []HeaderMatch `json:",omitempty"`
	Version string        `json:",omitempty"`
}

type HeaderMatch struct {
	Name           string      `json:",omitempty"`
	ExactMatch     string      `json:",omitempty"`
	SafeRegexMatch string      `json:",omitempty"`
	RangeMatch     *RangeMatch `json:",omitempty"`
	PresentMatch   *bool       `json:",omitempty"`
	PrefixMatch    string      `json:",omitempty"`
	SuffixMatch    string      `json:",omitempty"`
	InvertMatch    bool        `json:",omitempty"`
}

type RangeMatch struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

func fromRoute(route domain.Route) RouteMatch {
	result := RouteMatch{
		Prefix:  route.Prefix,
		Regexp:  route.Regexp,
		Path:    route.Path,
		Version: route.InitialDeploymentVersion,
	}
	if route.HeaderMatchers != nil {
		hmKeys := make([]HeaderMatch, len(route.HeaderMatchers))
		for i, hm := range route.HeaderMatchers {
			hmKeys[i] = fromHeaderMatcher(*hm)
		}
		result.Headers = hmKeys
	}
	return result
}

func fromHeaderMatcher(hm domain.HeaderMatcher) HeaderMatch {
	result := HeaderMatch{
		Name:           hm.Name,
		ExactMatch:     hm.ExactMatch,
		SafeRegexMatch: hm.SafeRegexMatch,
		PrefixMatch:    hm.PrefixMatch,
		SuffixMatch:    hm.SuffixMatch,
		InvertMatch:    hm.InvertMatch,
	}
	if hm.RangeMatch.Start.Valid && hm.RangeMatch.End.Valid {
		result.RangeMatch = &RangeMatch{
			Start: hm.RangeMatch.Start.Int64,
			End:   hm.RangeMatch.End.Int64,
		}
	}
	if hm.PresentMatch.Valid {
		boolValue := hm.PresentMatch.Bool
		result.PresentMatch = &boolValue
	}
	return result
}
