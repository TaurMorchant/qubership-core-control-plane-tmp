package ratelimit

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestService_Apply(t *testing.T) {
	srv := initService(t)

	ctx := context.Background()

	rateLimits, err := srv.GetRateLimits(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rateLimits))

	err = srv.SaveRateLimit(ctx, &dto.RateLimit{
		Name:                  "test-rate-limit",
		LimitRequestPerSecond: 10,
		Priority:              "PROJECT",
	})
	assert.Nil(t, err)

	rateLimits, err = srv.GetRateLimits(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rateLimits))
	rateLimit := rateLimits[0]
	assert.Equal(t, "test-rate-limit", rateLimit.Name)
	assert.Equal(t, 10, rateLimit.LimitRequestPerSecond)
	assert.Equal(t, "PROJECT", rateLimit.Priority)

	err = srv.SaveRateLimit(ctx, &dto.RateLimit{
		Name:                  "test-rate-limit",
		LimitRequestPerSecond: 0, // 0 means we need to delete this config
		Priority:              "PROJECT",
	})
	assert.Nil(t, err)

	rateLimits, err = srv.GetRateLimits(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rateLimits))

	err = srv.SaveRateLimit(ctx, &dto.RateLimit{
		Name:                  "test-rate-limit",
		LimitRequestPerSecond: 100,
		Priority:              "PRODUCT",
	})
	assert.Nil(t, err)

	rateLimits, err = srv.GetRateLimits(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rateLimits))
	rateLimit = rateLimits[0]
	assert.Equal(t, "test-rate-limit", rateLimit.Name)
	assert.Equal(t, 100, rateLimit.LimitRequestPerSecond)
	assert.Equal(t, "PRODUCT", rateLimit.Priority)

	err = srv.DeleteRateLimit(ctx, &dto.RateLimit{
		Name:     "test-rate-limit",
		Priority: "PRODUCT",
	})
	assert.Nil(t, err)

	rateLimits, err = srv.GetRateLimits(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rateLimits))
}

func TestService_GetResource(t *testing.T) {
	srv := initService(t)

	res := srv.GetRateLimitResource()
	assert.Equal(t, "RateLimit", res.GetKey().Kind)
	assert.NotNil(t, res.GetDefinition())
	isValid, _ := res.Validate(dto.RateLimit{
		Name:                  "",
		LimitRequestPerSecond: 10,
		Priority:              "PROJECT",
	})
	assert.False(t, isValid)
	isValid, _ = res.Validate(dto.RateLimit{
		Name:                  "test-rate-limit",
		LimitRequestPerSecond: 100,
		Priority:              "",
	})
	assert.True(t, isValid)
}

func TestService_ValidateRequest(t *testing.T) {
	srv := initService(t)

	assert.NotNil(t, srv.ValidateRequest(dto.RateLimit{
		Name:                  "",
		LimitRequestPerSecond: 10,
		Priority:              "PROJECT",
	}))
	assert.Nil(t, srv.ValidateRequest(dto.RateLimit{
		Name:                  "test-rate-limit",
		LimitRequestPerSecond: 100,
		Priority:              "",
	}))
}

func TestService_SingleOverriddenWithTrueValueForRoutingConfigRequestV3(t *testing.T) {
	resource := initService(t)
	isOverridden := resource.GetRateLimitResource().GetDefinition().IsOverriddenByCR(nil, nil, &dto.RateLimit{
		Name:                  "",
		LimitRequestPerSecond: 10,
		Priority:              "PROJECT",
		Overridden:            true,
	})
	assert.True(t, isOverridden)
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

func initService(t *testing.T) *Service {
	entitySrv, mockDao := getStorage(t)
	return NewService(mockDao, &BusMock{}, entitySrv)
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
