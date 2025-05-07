package migration

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_contains(t *testing.T) {
	route1 := v7Route{
		VirtualHostId: 1,
		RouteKey:      "/api/v1",
		Prefix:        "/api/v1",
		Regexp:        "",
	}
	route2 := v7Route{
		VirtualHostId: 2,
		RouteKey:      "/api/v2",
		Prefix:        "",
		Regexp:        "/api/v2",
	}
	arr := []v7Route{
		{
			VirtualHostId: 1,
			RouteKey:      "/api/v1",
			Prefix:        "/api/v1",
			Regexp:        "",
		},
		{
			VirtualHostId: 1,
			RouteKey:      "/api/v2",
			Prefix:        "/api/v1",
			Regexp:        "",
		},
	}
	assert.True(t, contains(arr, route1))
	assert.False(t, contains(arr, route2))
}

func Test_VersionFromAddressRegexp(t *testing.T) {
	str1 := "order-capture-backend-service-v1"
	str2 := "order-capture-backend-service-v10"
	str3 := "order-capture-backend-service-v112"
	str4 := "order-capture-backend-service"

	matches := versionFromAddressRegexp.FindStringSubmatch(str1)
	assert.Equal(t, 2, len(matches))
	assert.Equal(t, "v1", matches[1])

	matches = versionFromAddressRegexp.FindStringSubmatch(str2)
	assert.Equal(t, 2, len(matches))
	assert.Equal(t, "v10", matches[1])

	matches = versionFromAddressRegexp.FindStringSubmatch(str3)
	assert.Equal(t, 2, len(matches))
	assert.Equal(t, "v112", matches[1])

	matches = versionFromAddressRegexp.FindStringSubmatch(str4)
	assert.Equal(t, 0, len(matches))
}

func Test_fixRouteRegexRewriteV21(t *testing.T) {
	assert.Equal(t,
		"/api/v1/ext-frontend-api/customers_vars/\\1/subscriptions_vars/\\2/test\\3",
		fixRouteRegexRewriteV21("/api/v1/ext-frontend-api/customers_vars/.*/subscriptions_vars/.*/test\\1"))

	assert.Equal(t,
		"/api/v2/ext-frontend-api/test/\\1\\2",
		fixRouteRegexRewriteV21("/api/v2/ext-frontend-api/test/.*\\1"))

	assert.Equal(t,
		"/api/v1/ext-frontend-api/customers_vars/\\1/subscriptions_vars/test\\2",
		fixRouteRegexRewriteV21("/api/v1/ext-frontend-api/customers_vars/.*/subscriptions_vars/test\\1"))

	assert.Equal(t,
		"/api/v2/ext-frontend-api/\\1/test/\\2\\3",
		fixRouteRegexRewriteV21("/api/v2/ext-frontend-api/.*/test/.*\\1"))
}

func Test_virtualHostHasDefaultDomains(t *testing.T) {
	type args struct {
		virtualHost *V23VirtualHost
	}
	tests := []struct {
		name          string
		args          args
		want          bool
		wantNamespace string
	}{
		{
			name: "Old format domains without namespace",
			args: args{
				virtualHost: &V23VirtualHost{
					Name: "old-domain-virtual-host",
					Domains: []V23VirtualHostDomain{
						{Domain: "old-domain-virtual-host"},
						{Domain: "old-domain-virtual-host.svc"},
						{Domain: "old-domain-virtual-host.svc.cluster"},
						{Domain: "old-domain-virtual-host.svc.cluster.local"},
					},
				},
			},
			want:          true,
			wantNamespace: "",
		},
		{
			name: "Old format domains with namespace",
			args: args{
				virtualHost: &V23VirtualHost{
					Name: "old-domain-virtual-host",
					Domains: []V23VirtualHostDomain{
						{Domain: "old-domain-virtual-host"},
						{Domain: "old-domain-virtual-host.default.svc"},
						{Domain: "old-domain-virtual-host.default.svc.cluster"},
						{Domain: "old-domain-virtual-host.default.svc.cluster.local"},
					},
				},
			},
			want:          true,
			wantNamespace: "default",
		},
		{
			name: "New format ",
			args: args{
				virtualHost: &V23VirtualHost{
					Name: "old-domain-virtual-host",
					Domains: []V23VirtualHostDomain{
						{Domain: "old-domain-virtual-host:8080"},
						{Domain: "old-domain-virtual-host.default:8080"},
						{Domain: "old-domain-virtual-host.default.svc.cluster.local:8080"},
					},
				},
			},
			want:          false,
			wantNamespace: "",
		},
		{
			name: "Old format but gateway ",
			args: args{
				virtualHost: &V23VirtualHost{
					Name: "old-domain-virtual-host",
					Domains: []V23VirtualHostDomain{
						{Domain: "*"},
					},
				},
			},
			want:          false,
			wantNamespace: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := virtualHostHasDefaultDomains(tt.args.virtualHost)
			if got != tt.want {
				t.Errorf("virtualHostHasDefaultDomains() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantNamespace {
				t.Errorf("virtualHostHasDefaultDomains() got1 = %v, want %v", got1, tt.wantNamespace)
			}
		})
	}
}

func Test27Migrate_splitToUniqueAndDuplicatesByRegexpKey(t *testing.T) {
	nonUniqueRegexp := "/test-route/([^/]+)/test(/.*)?"
	uniqueRegexp := "/unique-route/([^/]+)(/.*)?"
	routes := []*V16Route{
		{
			Regexp:        nonUniqueRegexp,
			VirtualHostId: 0,
		},
		{
			Regexp:        nonUniqueRegexp,
			VirtualHostId: 0,
		},
		{
			Regexp:        nonUniqueRegexp,
			VirtualHostId: 1,
		},
		{
			Regexp:        uniqueRegexp,
			VirtualHostId: 0,
		},
	}

	unique, forDeletion := splitToUniqueAndDuplicatesByRegexpKey(routes)
	assert.Len(t, unique, 3)
	assert.Len(t, forDeletion, 1)

	assert.Contains(t, unique, routes[3])
	assert.Contains(t, unique, routes[2])

	assert.NotContains(t, forDeletion, routes[3])
	assert.NotContains(t, forDeletion, routes[2])
}

func Test91Migrate_ResolveEndpointProtocol(t *testing.T) {
	proto := resolveEndpointProtocol(&V91Endpoint{
		ClusterId: 1,
		Cluster:   &v91Cluster{TLSId: 0},
	})
	assert.Equal(t, "http", proto)

	proto = resolveEndpointProtocol(&V91Endpoint{
		ClusterId: 1,
		Cluster:   &v91Cluster{TLSId: 1},
	})

	proto = resolveEndpointProtocol(&V91Endpoint{
		ClusterId: 1,
		Cluster:   &v91Cluster{TLSId: 123},
	})
	assert.Equal(t, "https", proto)
}



func TestGetMigrations(t *testing.T) {
	testMigrations, err := GetMigrations()
	assert.Nil(t, err)
	assert.NotNil(t, testMigrations)
}

func TestCountEmptyRowsUtils(t *testing.T) {
	rows := sqlmock.NewRows([]string{})
	counter := CountRows(mockRowsToSqlRows(*rows))
	assert.Equal(t, 0, counter)
}

func TestCountRowsUtils(t *testing.T) {
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("one").AddRow("two").AddRow("three")
	counter := CountRows(mockRowsToSqlRows(*rows))
	assert.Equal(t, 3, counter)
}

func mockRowsToSqlRows(mockRows sqlmock.Rows) *sql.Rows {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(&mockRows)
	rows, _ := db.Query("select")
	return rows
}
