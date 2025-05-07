package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/configresources"
	mock_dao "github.com/netcracker/qubership-core-control-plane/test/mock/dao"
	"github.com/netcracker/qubership-core-control-plane/util"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBGRegistryController_HandleDeleteMicroserviceVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mock_dao.NewMockDao(ctrl)
	bgRegistry := newBgRegistryMock()

	controller := NewBGRegistryController(bgRegistry, mockDao)

	req := `brokenJson`
	response := SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2",
  "exists": true
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	validateInvocation := bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}, Exists: util.WrapValue(true)}, validateInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: false})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: false})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])
	applyInvocation := bgRegistry.GetInvocation("Apply")
	assert.Equal(t, 1, len(applyInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}, Exists: util.WrapValue(false)}, applyInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	bgRegistry.EnqueueResponse("Apply", responseMock{err: errors.New("test err")})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/versions/registry/services", "/api/v3/versions/registry/services",
		bytes.NewBufferString(req), controller.HandleDeleteMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])
	applyInvocation = bgRegistry.GetInvocation("Apply")
	assert.Equal(t, 1, len(applyInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}, Exists: util.WrapValue(false)}, applyInvocation.Args[0])
}

func TestBGRegistryController_HandlePostMicroserviceVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mock_dao.NewMockDao(ctrl)
	bgRegistry := newBgRegistryMock()

	controller := NewBGRegistryController(bgRegistry, mockDao)

	req := `brokenJson`
	response := SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/versions/registry", "/api/v3/versions/registry",
		bytes.NewBufferString(req), controller.HandlePostMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: false})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/versions/registry", "/api/v3/versions/registry",
		bytes.NewBufferString(req), controller.HandlePostMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	validateInvocation := bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/versions/registry", "/api/v3/versions/registry",
		bytes.NewBufferString(req), controller.HandlePostMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])
	applyInvocation := bgRegistry.GetInvocation("Apply")
	assert.Equal(t, 1, len(applyInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, applyInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2",
  "exists": false
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/versions/registry", "/api/v3/versions/registry",
		bytes.NewBufferString(req), controller.HandlePostMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}, Exists: util.WrapValue(false)}, validateInvocation.Args[0])
	applyInvocation = bgRegistry.GetInvocation("Apply")
	assert.Equal(t, 1, len(applyInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}, Exists: util.WrapValue(false)}, applyInvocation.Args[0])

	req = `{
  "services": ["trace-service", "echo-service"],
  "version": "v2"
}`
	bgRegistry.EnqueueResponse("Validate", responseMock{specs: true})
	bgRegistry.EnqueueResponse("Apply", responseMock{err: errors.New("test err")})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/versions/registry", "/api/v3/versions/registry",
		bytes.NewBufferString(req), controller.HandlePostMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	validateInvocation = bgRegistry.GetInvocation("Validate")
	assert.Equal(t, 1, len(validateInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, validateInvocation.Args[0])
	applyInvocation = bgRegistry.GetInvocation("Apply")
	assert.Equal(t, 1, len(applyInvocation.Args))
	assert.Equal(t, dto.ServicesVersionPayload{Version: "v2", Services: []string{"trace-service", "echo-service"}}, applyInvocation.Args[0])
}

func TestBGRegistryController_HandleGetMicroserviceVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mock_dao.NewMockDao(ctrl)
	bgRegistry := newBgRegistryMock()

	controller := NewBGRegistryController(bgRegistry, mockDao)

	returnVal := []dto.VersionInRegistry{{
		Version: "v1",
		Stage:   domain.ActiveStage,
		Clusters: []dto.Microservice{{
			Cluster:   "cluster1",
			Namespace: msaddr.LocalNamespace,
			Endpoints: []string{"http://cluster1:8080"},
		}},
	}}
	expectedJson, err := json.Marshal(returnVal)
	assert.Nil(t, err)

	response := sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?initialVersion=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?serviceName=test&version=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	// test GetAll
	bgRegistry.EnqueueResponse("GetAll", responseMock{specs: returnVal})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	verifyResponseBody(t, expectedJson, response)
	invocation := bgRegistry.GetInvocation("GetAll")
	assert.NotNil(t, invocation)

	bgRegistry.EnqueueResponse("GetAll", responseMock{err: errors.New("test err")})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	invocation = bgRegistry.GetInvocation("GetAll")
	assert.NotNil(t, invocation)

	// test GetMicroservicesForVersion
	bgRegistry.EnqueueResponse("GetMicroservicesForVersion", responseMock{specs: returnVal})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?version=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	verifyResponseBody(t, expectedJson, response)
	invocation = bgRegistry.GetInvocation("GetMicroservicesForVersion")
	assert.NotNil(t, invocation)
	assert.Equal(t, 1, len(invocation.Args))
	assert.Equal(t, domain.DeploymentVersion{Version: "v1"}, *invocation.Args[0].(*domain.DeploymentVersion))

	bgRegistry.EnqueueResponse("GetMicroservicesForVersion", responseMock{err: errors.New("test err")})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?version=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	invocation = bgRegistry.GetInvocation("GetMicroservicesForVersion")
	assert.NotNil(t, invocation)
	assert.Equal(t, 1, len(invocation.Args))
	assert.Equal(t, domain.DeploymentVersion{Version: "v1"}, *invocation.Args[0].(*domain.DeploymentVersion))

	// test GetVersionsForMicroservice
	bgRegistry.EnqueueResponse("GetVersionsForMicroservice", responseMock{specs: returnVal})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?serviceName=test-service&namespace=test-namespace", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	verifyResponseBody(t, expectedJson, response)
	invocation = bgRegistry.GetInvocation("GetVersionsForMicroservice")
	assert.NotNil(t, invocation)
	assert.Equal(t, 2, len(invocation.Args))
	assert.Equal(t, "test-service", invocation.Args[0].(string))
	assert.Equal(t, msaddr.Namespace{Namespace: "test-namespace"}, invocation.Args[1].(msaddr.Namespace))

	bgRegistry.EnqueueResponse("GetVersionsForMicroservice", responseMock{err: errors.New("test err")})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?serviceName=test-service&namespace=test-namespace", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	invocation = bgRegistry.GetInvocation("GetVersionsForMicroservice")
	assert.NotNil(t, invocation)
	assert.Equal(t, 2, len(invocation.Args))
	assert.Equal(t, "test-service", invocation.Args[0].(string))
	assert.Equal(t, msaddr.Namespace{Namespace: "test-namespace"}, invocation.Args[1].(msaddr.Namespace))

	// test GetMicroserviceCurrentVersion
	bgRegistry.EnqueueResponse("GetMicroserviceCurrentVersion", responseMock{specs: returnVal})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?serviceName=test-service&namespace=test-namespace&initialVersion=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	verifyResponseBody(t, expectedJson, response)
	invocation = bgRegistry.GetInvocation("GetMicroserviceCurrentVersion")
	assert.NotNil(t, invocation)
	assert.Equal(t, 3, len(invocation.Args))
	assert.Equal(t, "test-service", invocation.Args[0].(string))
	assert.Equal(t, msaddr.Namespace{Namespace: "test-namespace"}, invocation.Args[1].(msaddr.Namespace))
	assert.Equal(t, "v1", invocation.Args[2].(string))

	bgRegistry.EnqueueResponse("GetMicroserviceCurrentVersion", responseMock{err: errors.New("test err")})
	response = sendHttpRequest(t, http.MethodGet, "/api/v3/versions/registry", "/api/v3/versions/registry?serviceName=test-service&namespace=test-namespace&initialVersion=v1", controller.HandleGetMicroserviceVersions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	invocation = bgRegistry.GetInvocation("GetMicroserviceCurrentVersion")
	assert.NotNil(t, invocation)
	assert.Equal(t, 3, len(invocation.Args))
	assert.Equal(t, "test-service", invocation.Args[0].(string))
	assert.Equal(t, msaddr.Namespace{Namespace: "test-namespace"}, invocation.Args[1].(msaddr.Namespace))
	assert.Equal(t, "v1", invocation.Args[2].(string))
}

type registryMock struct {
	serviceMock
}

func newBgRegistryMock() *registryMock {
	return &registryMock{*newServiceMock("GetMicroserviceCurrentVersion", "GetVersionsForMicroservice", "GetMicroservicesForVersion", "GetAll", "Validate", "IsOverriddenByCR", "Apply")}
}

func (m *registryMock) GetMicroserviceCurrentVersion(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace, initialVersion string) ([]dto.VersionInRegistry, error) {
	return m.invokeAndReturn("GetMicroserviceCurrentVersion", serviceName, namespace, initialVersion)
}

func (m *registryMock) GetVersionsForMicroservice(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace) ([]dto.VersionInRegistry, error) {
	return m.invokeAndReturn("GetVersionsForMicroservice", serviceName, namespace)
}

func (m *registryMock) GetMicroservicesForVersion(ctx context.Context, repo dao.Repository, version *domain.DeploymentVersion) ([]dto.VersionInRegistry, error) {
	return m.invokeAndReturn("GetMicroservicesForVersion", version)
}

func (m *registryMock) GetAll(ctx context.Context, repo dao.Repository) ([]dto.VersionInRegistry, error) {
	return m.invokeAndReturn("GetAll")
}

func (m *registryMock) invokeAndReturn(funcName string, args ...any) ([]dto.VersionInRegistry, error) {
	resp := m.invoke(funcName, args...)
	if resp.specs == nil {
		return nil, resp.err
	}
	return resp.specs.([]dto.VersionInRegistry), resp.err
}

func (m *registryMock) Validate(ctx context.Context, res dto.ServicesVersionPayload) (bool, string) {
	resp := m.invoke("Validate", res)
	return resp.specs.(bool), ""
}

func (m *registryMock) IsOverriddenByCR(ctx context.Context, res dto.ServicesVersionPayload) bool {
	resp := m.invoke("IsOverriddenByCR", res)
	return resp.specs.(bool)
}

func (m *registryMock) Apply(ctx context.Context, res dto.ServicesVersionPayload) (any, error) {
	resp := m.invoke("Apply", res)
	return resp.specs, resp.err
}

func (m *registryMock) GetConfigRes() configresources.ConfigRes[dto.ServicesVersionPayload] {
	return configresources.ConfigRes[dto.ServicesVersionPayload]{}
}
