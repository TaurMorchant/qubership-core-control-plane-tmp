package route

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouteUtils_ValidateMetadataStringField(t *testing.T) {
	res, err := ValidateMetadataStringField(map[string]interface{}{"string": "string"}, "string")
	assert.True(t, res)
	assert.Empty(t, err)
	res, err = ValidateMetadataStringField(map[string]interface{}{"emptyString": ""}, "emptyString")
	assert.False(t, res)
	assert.NotEmpty(t, err)
	res, err = ValidateMetadataStringField(map[string]interface{}{}, "none")
	assert.False(t, res)
	assert.NotEmpty(t, err)
	res, err = ValidateMetadataStringField(map[string]interface{}{"empty": nil}, "empty")
	assert.False(t, res)
	assert.NotEmpty(t, err)
	res, err = ValidateMetadataStringField(map[string]interface{}{"int": 123}, "int")
	assert.False(t, res)
	assert.NotEmpty(t, err)
}
