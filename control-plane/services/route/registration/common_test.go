package registration

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"os"
	"strconv"
	"testing"
)

func Test_resolveVersions(t *testing.T) {
	type args struct {
		requestVersion string
		activeVersion  string
	}
	tests := []struct {
		name            string
		args            args
		wantInitVersion string
		wantVersion     string
	}{
		{"No request version", args{requestVersion: "", activeVersion: "v2"}, "v2", "v2"},
		{"Request active version", args{requestVersion: "v2", activeVersion: "v2"}, "v2", "v2"},
		{"Request candidate", args{requestVersion: "v3", activeVersion: "v2"}, "v3", "v3"},
		{"Moved routes", args{requestVersion: "v2", activeVersion: "v3"}, "v2", "v2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInitVersion, gotVersion, _ := ResolveVersions(getDao(), "", tt.args.requestVersion, tt.args.activeVersion)
			if gotInitVersion != tt.wantInitVersion {
				t.Errorf("ResolveVersions() gotInitVersion = %v, want %v", gotInitVersion, tt.wantInitVersion)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("ResolveVersions() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestCreateListenerWithPort_shouldCreateListenerWithDefaultPort_whenPortIsNotProvided(t *testing.T) {
	nodeGroup := "nodeGroup"
	builder := NewCommonEntityBuilder(nodeGroup)

	tests := []struct {
		name            string
		port            int
		supportTls      bool
		expPort         int
		expectedName    string
		expectedWithTls bool
		enableTls       bool
	}{
		{"Default", 0, false, 8080, nodeGroup + "-listener", false, false},
		{"Port8080", 8080, false, 8080, nodeGroup + "-listener", false, false},
		{"Port1234", 1234, false, 1234, nodeGroup + "-listener" + "-1234", false, false},
		{"Port1234WithEnableTls", 1234, false, 1234, nodeGroup + "-listener" + "-1234", false, true},
		{"Port1234WithTls", 1234, true, 1234, nodeGroup + "-listener" + "-1234", true, true},
		{"DefaultWithTls", 0, true, 8080, nodeGroup + "-listener", true, true},
		{"Port8080WithTls", 8080, true, 8080, nodeGroup + "-listener", true, true},
		{"Port8080WithTlsAndDisableTls", 8080, true, 8080, nodeGroup + "-listener", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.enableTls {
				_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
				defer os.Unsetenv("INTERNAL_TLS_ENABLED")
				configloader.Init(configloader.EnvPropertySource())
				tlsmode.SetUpTlsProperties()
			} else {
				defer os.Unsetenv("INTERNAL_TLS_ENABLED")
				configloader.Init(configloader.EnvPropertySource())
				tlsmode.SetUpTlsProperties()
			}

			listener := builder.CreateListenerWithCustomPort(tt.port, tt.supportTls)
			if listener.Name != tt.expectedName {
				t.Errorf("Expected name = %v, actual name = %v", tt.expectedName, listener.Name)
			}
			if listener.BindPort != strconv.Itoa(tt.expPort) {
				t.Errorf("Expected port = %v, actual port = %v", strconv.Itoa(tt.expPort), listener.BindPort)
			}
			if tt.expectedWithTls != listener.WithTls {
				t.Errorf("Expected tls = %v, actual tls = %v", tt.expectedWithTls, listener.WithTls)
			}
		})
	}
}
