package routekey

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/domain"
	asrt "github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestGenerateKey(t *testing.T) {
	assert := asrt.New(t)
	route1 := routeMatchToRandomRoute(RouteMatch{
		Prefix: "/api/v1/test",
		Headers: []HeaderMatch{
			{Name: "namespace", ExactMatch: "cloud-core"},
		},
	})
	route2 := routeMatchToRandomRoute(RouteMatch{
		Prefix: "/api/v1/test",
		Headers: []HeaderMatch{
			{Name: "namespace", ExactMatch: "cloud-core"},
		},
	})
	route3 := routeMatchToRandomRoute(RouteMatch{
		Prefix: "/api/v1/test",
	})
	key1 := GenerateKey(route1)
	key2 := GenerateKey(route2)
	key3 := GenerateKey(route3)

	fmt.Printf("Key: %s\n", key1)
	assert.True(key1 == key2)
	assert.False(key1 == key3)
}

func routeMatchToRandomRoute(rm RouteMatch) domain.Route {
	rand.Seed(time.Now().UnixNano())
	return domain.Route{
		Id:             rand.Int31(),
		Uuid:           strconv.FormatInt(rand.Int63(), 10),
		VirtualHostId:  rand.Int31(),
		Prefix:         rm.Prefix,
		Regexp:         rm.Regexp,
		Path:           rm.Path,
		ClusterName:    strconv.FormatInt(rand.Int63(), 10),
		HostRewrite:    strconv.FormatInt(rand.Int63(), 10),
		PrefixRewrite:  strconv.FormatInt(rand.Int63(), 10),
		RegexpRewrite:  strconv.FormatInt(rand.Int63(), 10),
		PathRewrite:    strconv.FormatInt(rand.Int63(), 10),
		Version:        rand.Int31(),
		HeaderMatchers: headerMatchToRandomHeaderMatchers(rm.Headers),
	}
}

func headerMatchToRandomHeaderMatchers(hms []HeaderMatch) []*domain.HeaderMatcher {
	headerMatchers := make([]*domain.HeaderMatcher, len(hms))
	for i, hm := range hms {
		headerMatchers[i] = &domain.HeaderMatcher{
			Id:             rand.Int31(),
			Name:           hm.Name,
			Version:        rand.Int31(),
			ExactMatch:     hm.ExactMatch,
			SafeRegexMatch: hm.SafeRegexMatch,
			RangeMatch:     domain.RangeMatch{},
			PresentMatch:   domain.NullBool{},
			PrefixMatch:    hm.PrefixMatch,
			SuffixMatch:    hm.SuffixMatch,
			InvertMatch:    hm.InvertMatch,
			RouteId:        rand.Int31(),
			Route:          nil,
		}
	}
	return headerMatchers
}
