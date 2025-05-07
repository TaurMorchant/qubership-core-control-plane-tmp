package clusterkey

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	asrt "github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateClusterKey(t *testing.T) {
	assert := asrt.New(t)
	clusterName := "microservice-service"
	msAddress := msaddr.NewMicroserviceAddress("http://microservice-service-v1:8080", "some-namespace")
	assert.Equal("microservice-service||microservice-service.some-namespace||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))
}

func TestGenerateClusterKeyWithEmptyClusterName(t *testing.T) {
	assert := asrt.New(t)
	clusterName := ""
	msAddress := msaddr.NewMicroserviceAddress("http://microservice-service:8080", "some-namespace")
	assert.Equal("microservice-service||microservice-service.some-namespace||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))
}

func TestGenerateClusterKeyWithoutEndpoint(t *testing.T) {
	assert := asrt.New(t)
	clusterName := "microservice-service"
	msAddress := msaddr.NewMicroserviceAddress("", "some-namespace")
	assert.Equal("microservice-service||microservice-service.some-namespace||80",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))
}

func TestGenerateClusterKeyWithoutNamespace(t *testing.T) {
	assert := asrt.New(t)
	clusterName := ""
	msAddress := msaddr.NewMicroserviceAddress("http://microservice-service:8080", "default")
	assert.Equal("microservice-service||microservice-service||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))
}

func TestGenerateClusterKeyForCustomCluster(t *testing.T) {
	assert := asrt.New(t)
	clusterName := "custom-cluster"
	msAddress := msaddr.NewMicroserviceAddress("http://microservice-service-v1:8080", "default")
	assert.Equal("custom-cluster||custom-cluster||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))

	msAddress = msaddr.NewMicroserviceAddress("http://microservice-service-v1.fake-namespace:8080", "")
	assert.Equal("custom-cluster||custom-cluster||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))

	msAddress = msaddr.NewMicroserviceAddress("http://microservice-service-v1:8080", "another-namespace")
	assert.Equal("custom-cluster||custom-cluster.another-namespace||8080",
		DefaultClusterKeyGenerator.GenerateKey(clusterName, msAddress))
}

func TestExtractFamilyName(t *testing.T) {
	assert := asrt.New(t)
	familyName := DefaultClusterKeyGenerator.ExtractFamilyName("my-family||my-family.namespace||8080")
	assert.Equal("my-family", familyName)
	familyName = DefaultClusterKeyGenerator.ExtractFamilyName("my-family||my-family||8443")
	assert.Equal("my-family", familyName)
}

func TestExtractNamespace(t *testing.T) {
	assert := asrt.New(t)
	namespace := DefaultClusterKeyGenerator.ExtractNamespace("my-family||my-family.my-namespace||8080")
	assert.Equal("my-namespace", namespace.Namespace)
	namespace = DefaultClusterKeyGenerator.ExtractNamespace("my-family||my-family||8443")
	assert.Equal("", namespace.Namespace)
}

func TestBuildKeyPrefix(t *testing.T) {
	assert := asrt.New(t)
	prefix := DefaultClusterKeyGenerator.BuildKeyPrefix("my-cluster", msaddr.Namespace{Namespace: "my-namespace"})
	assert.Equal("my-cluster||my-cluster.my-namespace||", prefix)
	prefix = DefaultClusterKeyGenerator.BuildKeyPrefix("my-cluster", msaddr.Namespace{Namespace: ""})
	assert.Equal("my-cluster||my-cluster||", prefix)
	prefix = DefaultClusterKeyGenerator.BuildKeyPrefix("my-cluster", msaddr.CurrentNamespace())
	assert.Equal("my-cluster||my-cluster||", prefix)
}
