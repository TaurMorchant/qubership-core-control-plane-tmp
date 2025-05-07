package dr

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetMode_shouldStandby_whenModeIsEmpty(t *testing.T) {
	mode := GetMode()
	assert.Equal(t, Active, mode)
}

func TestResolveMode_shouldStandby_whenModeIsStandBy(t *testing.T) {
	os.Setenv("EXECUTION_MODE", "standby")
	defer func() {
		os.Unsetenv("EXECUTION_MODE")
		ReloadMode()
	}()

	ReloadMode()
	mode := GetMode()
	assert.Equal(t, Standby, mode)
}

func TestResolveMode_shouldStandby_whenModeIsDisabled(t *testing.T) {
	os.Setenv("EXECUTION_MODE", "disabled")
	defer func() {
		os.Unsetenv("EXECUTION_MODE")
		ReloadMode()
	}()

	ReloadMode()
	mode := GetMode()
	assert.Equal(t, Standby, mode)
}
