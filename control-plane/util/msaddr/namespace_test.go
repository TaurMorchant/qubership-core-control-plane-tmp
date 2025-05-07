package msaddr

import (
	asrt "github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetGatewayNamespace(t *testing.T) {
	assert := asrt.New(t)
	os.Setenv(CloudNamespace, "cloud-namespace")
	assert.Equal("cloud-namespace", gatewayNamespace())
	os.Unsetenv(CloudNamespace)
}

func TestIsCurrentNamespace(t *testing.T) {
	assert := asrt.New(t)
	testNamespace := Namespace{Namespace: "cloud-namespace"}
	assert.Equal(false, testNamespace.IsCurrentNamespace())
	os.Setenv(CloudNamespace, "cloud-namespace")
	assert.Equal(true, testNamespace.IsCurrentNamespace())
	testNamespace.Namespace = "default"
	assert.Equal(true, testNamespace.IsCurrentNamespace())
	os.Unsetenv(CloudNamespace)
}

func TestIsLocalDevNamespace(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal(true, NewNamespace("cloud-namespace"+LocalDevNamespacePostfix).IsLocalDevNamespace())
	assert.Equal(false, NewNamespace("cloud-namespace").IsLocalDevNamespace())
}

func TestCurrentLocalDevNamespace(t *testing.T) {
	assert := asrt.New(t)
	currentNamespace := CurrentNamespace()
	assert.Equal(LocalNamespace, currentNamespace.Namespace)
	assert.False(currentNamespace.IsLocalDevNamespace())
	assert.True(currentNamespace.IsCurrentNamespace())
}

func TestEqualsEmptyNs(t *testing.T) {
	emptyNamespaces := []Namespace{CurrentNamespace(), *NewNamespace(""), {Namespace: ""}, Namespace{Namespace: DefaultNamespace}}
	nonEmptyNamespaces := []Namespace{{DefaultNamespace + "-1"}, {"another-one"}, {"cloud-core"}}

	runEqualsTC(t, emptyNamespaces, nonEmptyNamespaces)
}

func TestEqualsNonEmptyNs(t *testing.T) {
	const testNs = "test-ns"
	sameNamespaces := []Namespace{*NewNamespace(testNs), {Namespace: testNs}, {Namespace: testNs}}
	differentNamespaces := []Namespace{{DefaultNamespace + "-1"}, {"another-one"}, {"cloud-core"}}

	runEqualsTC(t, sameNamespaces, differentNamespaces)
}

func runEqualsTC(t *testing.T, sameNamespaces []Namespace, differentNamespaces []Namespace) {
	assert := asrt.New(t)
	for _, namespace := range sameNamespaces {
		for _, anotherNs := range sameNamespaces {
			assert.True(namespace.Equals(anotherNs))
			assert.True(anotherNs.Equals(namespace))
		}
		for _, anotherNs := range differentNamespaces {
			assert.False(namespace.Equals(anotherNs))
			assert.False(anotherNs.Equals(namespace))
		}
	}
}
