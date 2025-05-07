package dto

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_converter_convertHashPolicy_differentValueCookieTtl(t *testing.T) {
	hashPolicy := domain.HashPolicy{CookieName: "aa"}
	result := DefaultResponseConverter.convertHashPolicy(&hashPolicy)
	assert.Nil(t, result.Cookie.Ttl)

	hashPolicy = domain.HashPolicy{CookieName: "aa", CookieTTL: domain.NewNullInt(0)}
	result = DefaultResponseConverter.convertHashPolicy(&hashPolicy)
	assert.Equal(t, int64(0), *result.Cookie.Ttl)

	hashPolicy = domain.HashPolicy{CookieName: "aa", CookieTTL: domain.NewNullInt(10)}
	result = DefaultResponseConverter.convertHashPolicy(&hashPolicy)
	assert.Equal(t, int64(10), *result.Cookie.Ttl)
}
