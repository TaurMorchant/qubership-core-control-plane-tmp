package ui

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_adaptHashPoliciesToUI_differentValueCookieTtl(t *testing.T) {
	hashPolicy := domain.HashPolicy{CookieName: "aa"}
	result := adaptHashPoliciesToUI([]*domain.HashPolicy{&hashPolicy})
	assert.Nil(t, result[0].Cookie.Ttl)

	hashPolicy = domain.HashPolicy{CookieName: "aa", CookieTTL: domain.NewNullInt(0)}
	result = adaptHashPoliciesToUI([]*domain.HashPolicy{&hashPolicy})
	assert.Equal(t, int64(0), *(result[0].Cookie.Ttl))

	hashPolicy = domain.HashPolicy{CookieName: "aa", CookieTTL: domain.NewNullInt(10)}
	result = adaptHashPoliciesToUI([]*domain.HashPolicy{&hashPolicy})
	assert.Equal(t, int64(10), *(result[0].Cookie.Ttl))
}
