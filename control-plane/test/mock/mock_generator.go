package mock

//go:generate mockgen -source=../../dao/api.go -destination=./dao/stub_api.go -package=mock_dao -imports memdb=github.com/hashicorp/go-memdb
//go:generate mockgen -source=../../envoy/cache/action/factory.go -destination=./envoy/cache/action/stub_factory.go -package=mock_action
//go:generate mockgen -source=../../envoy/cache/action/action.go -destination=./envoy/cache/action/stub_action.go -package=mock_action -imports cache=github.com/envoyproxy/go-control-plane/pkg/cache/v3
//go:generate mockgen -source=../../envoy/cache/event/parser.go -destination=./envoy/cache/event/stub_parser.go -package=mock_event -imports memdb=github.com/hashicorp/go-memdb
//go:generate mockgen -source=../../envoy/cache/event/processor.go -destination=./envoy/cache/event/stub_processor.go -package=mock_event -imports memdb=github.com/hashicorp/go-memdb
//go:generate mockgen -source=../../envoy/cache/builder/listener/listener.go -destination=./envoy/cache/builder/listener/stub_listener.go -package=mock_listener
//go:generate mockgen -source=../../envoy/cache/builder/builder.go -destination=./envoy/cache/builder/stub_builder.go -package=mock_builder
//go:generate mockgen -source=../../envoy/cache/builder/routeconfig/routeconfig.go -destination=./envoy/cache/builder/routeconfig/stub_routeconfig.go -package=mock_routeconfig
//go:generate mockgen -source=../../envoy/cache/builder/routeconfig/virtualhost.go -destination=./envoy/cache/builder/routeconfig/stub_virtualhost.go -package=mock_routeconfig
//go:generate mockgen -source=../../envoy/cache/builder/routeconfig/routepreparer.go -destination=./envoy/cache/builder/routeconfig/stub_routepreparer.go -package=mock_routeconfig
//go:generate mockgen -source=../../event/bus/bus.go -destination=./event/bus/stub_bus.go -package=mock_bus
//go:generate mockgen -source=../../db/dbprovider.go -destination=./db/stub_dbprovider.go -package=mock_db -imports pg=github.com/uptrace/bun
//go:generate mockgen -source=../../clustering/lifecycle.go -destination=./clustering/stub_lifecycle.go -package=mock_clustering
//go:generate mockgen -source=../../db/listener.go -destination=./db/stub_listener.go -package=mock_db
//go:generate mockgen -source=../../constancy/storage.go -destination=./constancy/stub_storage.go -package=mock_constancy -imports pg=github.com/uptrace/bun
//go:generate mockgen -source=../../constancy/storage_wrapper.go -destination=./constancy/stub_storage_wrapper.go -package=mock_constancy -imports pg=github.com/uptrace/bun
//go:generate mockgen -source=../../restcontrollers/v3/routes.go -destination=./restcontrollers/v3/stub_routes.go -package=mock_v3
//go:generate mockgen -source=../../services/provider/provider.go -destination=./services/provider/stub_provider.go -package=mock_provider
//go:generate mockgen -source=../../services/httpFilter/extAuthz/service.go -destination=./services/httpFilter/extAuthz/stub_service.go -package=mock_extAuthz
//go:generate mockgen -source=../../services/route/registration.go -destination=./services/route/stub_registration.go -package=mock_route
