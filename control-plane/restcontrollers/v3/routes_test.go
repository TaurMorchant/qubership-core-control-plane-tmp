package v3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/dr"
	"github.com/netcracker/qubership-core-control-plane/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	mock_v3 "github.com/netcracker/qubership-core-control-plane/test/mock/restcontrollers/v3"
	"github.com/netcracker/qubership-core-lib-go-error-handling/v3/tmf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	RoutesV3Path               = "/api/v3/control-plane/routes"
	DomainsV3Path              = "/api/v3/control-plane/domains"
	EndpointsV3Path            = "/api/v3/control-plane/endpoints"
	virtualServiceName         = "virtualServiceName"
	testVirtualServiceEndpoint = "/api/v2/control-plane/routes/:nodeGroup/:virtualServiceName"
	testVirtualServiceUrlPath  = "/api/v2/control-plane/routes/" + testNodeGroupName + "/" + virtualServiceName
)

func TestHandleDeleteEndpoints_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, _ := getV3Controller(ctrl)
	correctBody, err := json.Marshal([]dto.EndpointDeleteRequest{
		{
			Version: "1",
		},
	})
	assert.Nil(t, err)
	wrongBody, err := json.Marshal("test")
	assert.Nil(t, err)

	// UnmarshalRequestError
	response := SendHttpRequestWithBody(t,
		http.MethodDelete,
		EndpointsV3Path,
		EndpointsV3Path,
		bytes.NewBuffer(wrongBody),
		activeDCsController.HandleDeleteEndpoints,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	mockService.EXPECT().DeleteEndpoints(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*domain.Endpoint{}, fmt.Errorf("test error"))
	response = SendHttpRequestWithBody(t,
		http.MethodDelete,
		EndpointsV3Path,
		EndpointsV3Path,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleDeleteEndpoints,
	)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandleDeleteVirtualServiceDomains_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	correctBody, err := json.Marshal([]dto.DomainDeleteRequestV3{})
	assert.Nil(t, err)
	wrongBody, err := json.Marshal("test")
	assert.Nil(t, err)

	// UnmarshalRequestError
	response := SendHttpRequestWithBody(t,
		http.MethodDelete,
		DomainsV3Path,
		DomainsV3Path,
		bytes.NewBuffer(wrongBody),
		activeDCsController.HandleDeleteVirtualServiceDomains,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError
	mockValidator.EXPECT().ValidateDomainDeleteRequestV3(gomock.Any()).Return(false, "")
	response = SendHttpRequestWithBody(t,
		http.MethodDelete,
		DomainsV3Path,
		DomainsV3Path,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleDeleteVirtualServiceDomains,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.Code, tmfResponse.Code)

	// UnknownErrorCode - empty nodeGroup
	mockValidator.EXPECT().ValidateDomainDeleteRequestV3(gomock.Any()).Return(true, "")
	mockService.EXPECT().DeleteDomains(gomock.Any(), gomock.Any()).Return([]*domain.VirtualHostDomain{}, fmt.Errorf("test error"))
	response = SendHttpRequestWithBody(t,
		http.MethodDelete,
		DomainsV3Path,
		DomainsV3Path,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleDeleteVirtualServiceDomains,
	)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandleDeleteVirtualServiceRoutes_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, _ := getV3Controller(ctrl)
	correctBody, err := json.Marshal([]dto.RouteDeleteRequestV3{})
	assert.Nil(t, err)
	wrongBody, err := json.Marshal("test")
	assert.Nil(t, err)

	// UnmarshalRequestError - empty nodeGroup
	response := SendHttpRequestWithBody(t,
		http.MethodDelete,
		RoutesV3Path,
		RoutesV3Path,
		bytes.NewBuffer(wrongBody),
		activeDCsController.HandleDeleteVirtualServiceRoutes,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode - empty nodeGroup
	mockService.EXPECT().DeleteRoutes(gomock.Any(), gomock.Any()).Return([]*domain.Route{}, fmt.Errorf("test error"))
	response = SendHttpRequestWithBody(t,
		http.MethodDelete,
		RoutesV3Path,
		RoutesV3Path,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleDeleteVirtualServiceRoutes,
	)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandleCreateVirtualService_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	correctBody, err := json.Marshal(&dto.RoutingConfigRequestV3{})
	assert.Nil(t, err)

	// ValidationRequestError - empty nodeGroup
	response := SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		strings.Replace(testVirtualServiceEndpoint, ":nodeGroup", " ", 1),
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleCreateVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - empty virtualServiceName
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		strings.Replace(testVirtualServiceEndpoint, ":virtualServiceName", " ", 1),
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleCreateVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnmarshalRequestError
	wrongBody, err := json.Marshal("test")
	assert.Nil(t, err)
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(wrongBody),
		activeDCsController.HandleCreateVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - body
	mockValidator.EXPECT().ValidateVirtualService(gomock.Any(), gomock.Any()).Return(false, "")
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleCreateVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	mockValidator.EXPECT().ValidateVirtualService(gomock.Any(), gomock.Any()).Return(true, "")
	mockService.EXPECT().CreateVirtualService(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandleCreateVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandlePutVirtualService_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	correctBody, err := json.Marshal(&dto.RoutingConfigRequestV3{})
	assert.Nil(t, err)

	// ValidationRequestError - empty nodeGroup
	response := SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		strings.Replace(testVirtualServiceEndpoint, ":nodeGroup", " ", 1),
		bytes.NewBuffer(correctBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - empty virtualServiceName
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		strings.Replace(testVirtualServiceEndpoint, ":virtualServiceName", " ", 1),
		bytes.NewBuffer(correctBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnmarshalRequestError
	wrongBody, err := json.Marshal("test")
	assert.Nil(t, err)
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(wrongBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - body
	mockValidator.EXPECT().ValidateVirtualServiceUpdate(gomock.Any(), gomock.Any()).Return(false, "")
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	mockValidator.EXPECT().ValidateVirtualServiceUpdate(gomock.Any(), gomock.Any()).Return(true, "")
	mockService.EXPECT().UpdateVirtualService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)

	// BlueGreenConflictError
	mockValidator.EXPECT().ValidateVirtualServiceUpdate(gomock.Any(), gomock.Any()).Return(true, "")
	mockService.EXPECT().UpdateVirtualService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.LegacyRouteDisallowed)
	response = SendHttpRequestWithBody(t,
		http.MethodPut,
		testVirtualServiceEndpoint,
		testVirtualServiceEndpoint,
		bytes.NewBuffer(correctBody),
		activeDCsController.HandlePutVirtualService,
	)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.BlueGreenConflictError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.BlueGreenConflictError.Code, tmfResponse.Code)
}

func TestHandleDeleteVirtualService_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, _ := getV3Controller(ctrl)

	// ValidationRequestError - empty nodeGroup
	response := SendHttpRequestWithoutBody(t, http.MethodDelete, testVirtualServiceEndpoint, strings.Replace(testVirtualServiceEndpoint, ":nodeGroup", " ", 1), activeDCsController.HandleDeleteVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - empty virtualServiceName
	response = SendHttpRequestWithoutBody(t, http.MethodDelete, testVirtualServiceEndpoint, strings.Replace(testVirtualServiceEndpoint, ":virtualServiceName", " ", 1), activeDCsController.HandleDeleteVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	mockService.EXPECT().DeleteVirtualService(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))
	response = SendHttpRequestWithoutBody(t, http.MethodDelete, testVirtualServiceEndpoint, testVirtualServiceEndpoint, activeDCsController.HandleDeleteVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandleGetVirtualService_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	activeDCsController, mockService, _ := getV3Controller(ctrl)

	// ValidationRequestError - empty nodeGroup
	response := SendHttpRequestWithoutBody(t, http.MethodGet, testVirtualServiceEndpoint, strings.Replace(testVirtualServiceEndpoint, ":nodeGroup", " ", 1), activeDCsController.HandleGetVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError - empty virtualServiceName
	response = SendHttpRequestWithoutBody(t, http.MethodGet, testVirtualServiceEndpoint, strings.Replace(testVirtualServiceEndpoint, ":virtualServiceName", " ", 1), activeDCsController.HandleGetVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	mockService.EXPECT().GetVirtualService(gomock.Any(), gomock.Any()).Return(dto.VirtualServiceResponse{}, fmt.Errorf("test error"))
	response = SendHttpRequestWithoutBody(t, http.MethodGet, testVirtualServiceEndpoint, testVirtualServiceEndpoint, activeDCsController.HandleGetVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandlePostRoutingConfig_shouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)

	// UnmarshalRequestError
	requestAsBytes, err := json.Marshal("test")
	assert.Nil(t, err)

	response := SendHttpRequestWithBody(t, http.MethodPost, RoutesV3Path, RoutesV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandlePostRoutingConfig)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse := readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnmarshalRequestError.ErrorCode.Code, tmfResponse.Code)

	// ValidationRequestError
	request := &dto.RoutingConfigRequestV3{
		Gateways:        []string{"gv1", "gv2"},
		VirtualServices: []dto.VirtualService{},
	}
	requestAsBytes, err = json.Marshal(request)
	assert.Nil(t, err)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(false, "")

	response = SendHttpRequestWithBody(t, http.MethodPost, RoutesV3Path, RoutesV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandlePostRoutingConfig)
	assert.NotNil(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.GetHttpCode(), response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.ValidationRequestError.ErrorCode.Code, tmfResponse.Code)

	// UnknownErrorCode
	request = &dto.RoutingConfigRequestV3{
		Namespace:       "test",
		Gateways:        []string{"gv1", "gv2"},
		VirtualServices: []dto.VirtualService{},
	}
	requestAsBytes, err = json.Marshal(request)
	assert.Nil(t, err)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(true, "")
	mockService.EXPECT().RegisterRoutingConfig(gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))

	response = SendHttpRequestWithBody(t, http.MethodPost, RoutesV3Path, RoutesV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandlePostRoutingConfig)
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	tmfResponse = readTmfResponse(t, response)
	assert.Equal(t, errorcodes.UnknownErrorCode.Code, tmfResponse.Code)
}

func TestHandleDeleteEndpoints_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, _ := getV3Controller(ctrl)
	request := &[]dto.EndpointDeleteRequest{}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockService.EXPECT().DeleteEndpoints(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodDelete, EndpointsV3Path, EndpointsV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandleDeleteEndpoints)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandleDeleteVirtualServiceDomains_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	request := &[]dto.DomainDeleteRequestV3{}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockService.EXPECT().DeleteDomains(gomock.Any(), gomock.Any()).Times(0)
	mockValidator.EXPECT().ValidateDomainDeleteRequestV3(gomock.Any()).Return(true, "")

	response := SendHttpRequestWithBody(t, http.MethodDelete, DomainsV3Path, DomainsV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandleDeleteVirtualServiceDomains)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandleDeleteVirtualServiceRoutes_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, _ := getV3Controller(ctrl)
	request := &[]dto.RouteDeleteRequestV3{}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockService.EXPECT().DeleteRoutes(gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodDelete, RoutesV3Path, RoutesV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandleDeleteVirtualServiceRoutes)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandleCreateVirtualService_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	request := &dto.VirtualService{}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockValidator.EXPECT().ValidateVirtualService(gomock.Any(), gomock.Any()).Return(true, "")
	mockService.EXPECT().CreateVirtualService(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodPost, testVirtualServiceEndpoint, testVirtualServiceUrlPath,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandleCreateVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandlePutVirtualService_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	request := &dto.VirtualService{}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockValidator.EXPECT().ValidateVirtualServiceUpdate(gomock.Any(), gomock.Any()).Return(true, "")
	mockService.EXPECT().UpdateVirtualService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodPost, testVirtualServiceEndpoint, testVirtualServiceUrlPath,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandlePutVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandleDeleteVirtualService_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, _ := getV3Controller(ctrl)
	request := &dto.RoutingConfigRequestV3{
		Namespace:       "testNamespace",
		Gateways:        []string{"gv1", "gv2"},
		VirtualServices: []dto.VirtualService{},
	}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockService.EXPECT().DeleteVirtualService(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodPost, testVirtualServiceEndpoint, testVirtualServiceUrlPath,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandleDeleteVirtualService)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandlePostRoutingConfig_withStandByMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	setStandByMode()
	defer clearStandByMode()

	activeDCsController, mockService, mockValidator := getV3Controller(ctrl)
	request := &dto.RoutingConfigRequestV3{
		Namespace:       "testNamespace",
		Gateways:        []string{"gv1", "gv2"},
		VirtualServices: []dto.VirtualService{},
	}
	requestAsBytes, err := json.Marshal(request)
	assert.Nil(t, err)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(true, "")
	mockService.EXPECT().RegisterRoutingConfig(gomock.Any(), gomock.Any()).Times(0)

	response := SendHttpRequestWithBody(t, http.MethodPost, RoutesV3Path, RoutesV3Path,
		bytes.NewBuffer(requestAsBytes), activeDCsController.HandlePostRoutingConfig)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func setStandByMode() {
	os.Setenv("EXECUTION_MODE", "standby")
	dr.ReloadMode()
}

func clearStandByMode() {
	os.Unsetenv("EXECUTION_MODE")
	dr.ReloadMode()
}

func getV3Controller(ctrl *gomock.Controller) (*RoutingConfigController, *mock_v3.MockRouteService, *mock_v3.MockRequestValidator) {
	service := getMockRouteService(ctrl)
	validator := getMockRequestValidator(ctrl)
	return NewRoutingConfigController(service, validator), service, validator
}

func getMockRouteService(ctrl *gomock.Controller) *mock_v3.MockRouteService {
	mock := mock_v3.NewMockRouteService(ctrl)
	return mock
}

func getMockRequestValidator(ctrl *gomock.Controller) *mock_v3.MockRequestValidator {
	mock := mock_v3.NewMockRequestValidator(ctrl)
	return mock
}

func readTmfResponse(t *testing.T, response *http.Response) tmf.Response {
	tmfResponse := tmf.Response{}
	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &tmfResponse)

	return tmfResponse
}
