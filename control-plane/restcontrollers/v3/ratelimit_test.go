package v3

import (
	"bytes"
	"encoding/json"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/ratelimit"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestController_Apply(t *testing.T) {
	srv := initService(t)
	controller := NewRateLimitController(srv)

	response := SendHttpRequestWithoutBody(t, http.MethodGet, "/api/v3/rate-limits",
		"/api/v3/rate-limits", controller.HandleGetRateLimit)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	var actual []*dto.RateLimit
	assert.Nil(t, json.Unmarshal(bodeBytes, &actual))
	assert.Equal(t, 0, len(actual))

	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/rate-limits",
		"/api/v3/rate-limits",
		bytes.NewBufferString(`{"name": "test-rate-limit", "limitRequestPerSecond": 10, "priority": "PROJECT"}`),
		controller.HandlePostRateLimit)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err = ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	response = SendHttpRequestWithoutBody(t, http.MethodGet, "/api/v3/rate-limits",
		"/api/v3/rate-limits", controller.HandleGetRateLimit)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err = ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Nil(t, json.Unmarshal(bodeBytes, &actual))
	assert.Equal(t, 1, len(actual))
	actualRateLimit := actual[0]
	assert.Equal(t, "test-rate-limit", actualRateLimit.Name)
	assert.Equal(t, 10, actualRateLimit.LimitRequestPerSecond)
	assert.Equal(t, "PROJECT", actualRateLimit.Priority)

	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/rate-limits",
		"/api/v3/rate-limits",
		bytes.NewBufferString(`{"name": "test-rate-limit", "priority": "PROJECT"}`),
		controller.HandleDeleteRateLimit)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err = ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	response = SendHttpRequestWithoutBody(t, http.MethodGet, "/api/v3/rate-limits",
		"/api/v3/rate-limits", controller.HandleGetRateLimit)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err = ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Nil(t, json.Unmarshal(bodeBytes, &actual))
	assert.Equal(t, 0, len(actual))
}

type BusMock struct{}

func (m *BusMock) Publish(topic string, data interface{}) error {
	return nil
}

func (m *BusMock) Shutdown() {}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

func initService(t *testing.T) *ratelimit.Service {
	entitySrv, mockDao := getStorage(t)
	return ratelimit.NewService(mockDao, &BusMock{}, entitySrv)
}

func getStorage(t *testing.T) (*entity.Service, *dao.InMemDao) {
	mockDao := dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
	v1 := &domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		return dao.SaveDeploymentVersion(v1)
	})
	assert.Nil(t, err)
	entityService := entity.NewService("v1")
	return entityService, mockDao
}
