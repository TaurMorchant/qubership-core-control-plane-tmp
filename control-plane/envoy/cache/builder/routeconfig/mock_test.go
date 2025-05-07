package routeconfig

import (
	"github.com/golang/mock/gomock"
	mock_dao "github.com/netcracker/qubership-core-control-plane/test/mock/dao"
	mock_routeconfig "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder/routeconfig"
)

func getMockDao(ctrl *gomock.Controller) *mock_dao.MockDao {
	mockDao := mock_dao.NewMockDao(ctrl)
	mockDao.EXPECT().WithWTx(gomock.Any()).AnyTimes().Return(nil, nil)
	return mockDao
}

func getMockRouteBuilder(ctrl *gomock.Controller) *mock_routeconfig.MockRouteBuilder {
	mockRouteBuilder := mock_routeconfig.NewMockRouteBuilder(ctrl)
	return mockRouteBuilder
}

func getMockVersionAliasesProvider(ctrl *gomock.Controller) *mock_routeconfig.MockVersionAliasesProvider {
	mockVersionAliasesProvider := mock_routeconfig.NewMockVersionAliasesProvider(ctrl)
	return mockVersionAliasesProvider
}

func getMockRoutePreparer(ctrl *gomock.Controller) *mock_routeconfig.MockRoutePreparer {
	mockRoutePreparer := mock_routeconfig.NewMockRoutePreparer(ctrl)
	return mockRoutePreparer
}
