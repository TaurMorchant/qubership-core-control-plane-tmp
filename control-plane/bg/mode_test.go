package bg

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetMode_shouldBg1_whenPeerIsEmpty(t *testing.T) {
	assert.Equal(t, BlueGreen1, GetMode())
}

func TestResolveMode_shouldBg2_whenPeerIsNotEmpty(t *testing.T) {
	os.Setenv("PEER_NAMESPACE", "12345")
	defer func() {
		os.Unsetenv("PEER_NAMESPACE")
		ReloadMode()
	}()

	ReloadMode()
	assert.Equal(t, BlueGreen2, GetMode())
}
