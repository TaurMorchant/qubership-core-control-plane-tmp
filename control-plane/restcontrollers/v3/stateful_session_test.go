package v3

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/statefulsession"
	"github.com/netcracker/qubership-core-lib-go-error-handling/v3/tmf"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

type statefulSessionSrvMock struct {
	*serviceMock
}

func mockStatefulSessionSrv() *statefulSessionSrvMock {
	return &statefulSessionSrvMock{newServiceMock("ApplyStatefulSession", "FindAll")}
}

func (m *statefulSessionSrvMock) ApplyStatefulSession(_ context.Context, spec *dto.StatefulSession) error {
	return m.invoke("ApplyStatefulSession", spec).err
}

func (m *statefulSessionSrvMock) FindAll(ctx context.Context) ([]*dto.StatefulSession, error) {
	resp := m.invoke("FindAll")
	if resp.specs == nil {
		return nil, resp.err
	}
	return resp.specs.([]*dto.StatefulSession), resp.err
}

func TestStatefulSessionController_HandlePostBadStatefulSessionTC1(t *testing.T) {
	testInvalidRequest(t, `{
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`)
}

func TestStatefulSessionController_HandlePostBadStatefulSessionTC2(t *testing.T) {
	testInvalidRequest(t, `{
  "gateways": [],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`)
}

func testInvalidRequest(t *testing.T, req string) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})

	response := SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandlePostStatefulSession)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestStatefulSessionController_HandlePostStatefulSession(t *testing.T) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})

	req := `{
  "gateways": [ "facade-gw" ],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`
	var expectedSpec dto.StatefulSession
	assert.Nil(t, json.Unmarshal([]byte(req), &expectedSpec))

	response := SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandlePostStatefulSession)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	invoke := srv.GetInvocation("ApplyStatefulSession")
	assert.Equal(t, 1, len(invoke.Args))
	assert.Equal(t, expectedSpec, *(invoke.Args[0].(*dto.StatefulSession)))
}

func TestStatefulSessionController_HandlePostStatefulSessionErr(t *testing.T) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})

	req := `{
  "gateways": [ "facade-gw" ],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: fmt.Errorf("some unknwon error")})
	response := SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandlePostStatefulSession)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: errorcodes.NewCpError(errorcodes.NotFoundEntityError, statefulsession.ErrNoCluster.Error(), nil)})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandlePostStatefulSession)
	result := readTmfBody(t, response)
	assert.Equal(t, errorcodes.NotFoundEntityError.Code, result.Code)

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: errorcodes.NewCpError(errorcodes.OperationOnArchivedVersionError, statefulsession.ErrVersionArchived.Error(), nil)})
	response = SendHttpRequestWithBody(t, http.MethodPost, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandlePostStatefulSession)
	result = readTmfBody(t, response)
	assert.Equal(t, errorcodes.OperationOnArchivedVersionError.Code, result.Code)
}

func readTmfBody(t *testing.T, response *http.Response) tmf.Response {
	bytesBody, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	var result tmf.Response
	err = json.Unmarshal(bytesBody, &result)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return result
}

func TestStatefulSessionController_HandleDeleteStatefulSessionErr(t *testing.T) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})

	req := `{
  "gateways": [ "facade-gw" ],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: fmt.Errorf("some unknwon error")})
	response := SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandleDeleteStatefulSession)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	req = `{
  "gateways": [],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: statefulsession.ErrNoCluster})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandleDeleteStatefulSession)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	srv.EnqueueResponse("ApplyStatefulSession", responseMock{err: statefulsession.ErrNoCluster})
	response = SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString("invalid req"), controller.HandleDeleteStatefulSession)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestStatefulSessionController_HandleDeleteStatefulSession(t *testing.T) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})

	req := `{
  "gateways": [ "facade-gw" ],
  "cluster": "test-service",
  "version": "v1",
  "port": 8080,
  "enabled": true,
  "cookie": {
    "name": "sticky-cookie",
    "ttl": 0,
    "path": "/"
  }
}`
	var expectedSpec dto.StatefulSession
	assert.Nil(t, json.Unmarshal([]byte(req), &expectedSpec))
	expectedSpec.Cookie = nil
	expectedSpec.Enabled = nil

	response := SendHttpRequestWithBody(t, http.MethodDelete, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(req), controller.HandleDeleteStatefulSession)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	invoke := srv.GetInvocation("ApplyStatefulSession")
	assert.Equal(t, expectedSpec, *(invoke.Args[0].(*dto.StatefulSession)))
}

func TestStatefulSessionController_HandleGetStatefulSession(t *testing.T) {
	srv := mockStatefulSessionSrv()
	controller := NewStatefulSessionController(srv, dto.RoutingV3RequestValidator{})
	portVal := 8080
	ttlVal := int64(10000)
	cookie := dto.Cookie{
		Name: "cookie",
		Ttl:  &ttlVal,
		Path: domain.NullString{NullString: sql.NullString{
			String: "/",
			Valid:  true,
		}},
	}
	expectedSessions := []*dto.StatefulSession{{
		Version:   "v1",
		Namespace: "ns1",
		Cluster:   "cluster1",
		Hostname:  "hostname1",
		Gateways:  []string{"gw1"},
		Port:      &portVal,
		Enabled:   nil,
		Cookie:    &cookie,
		Route:     nil,
	}, {
		Version:   "v2",
		Namespace: "ns2",
		Cluster:   "cluster2",
		Hostname:  "hostname2",
		Gateways:  []string{"gw2"},
		Port:      &portVal,
		Enabled:   nil,
		Cookie:    nil,
		Route: &dto.RouteMatcher{
			Prefix: domain.NullString{NullString: sql.NullString{String: "/api/v2/route", Valid: true}}},
	}}

	srv.EnqueueResponse("FindAll", responseMock{err: fmt.Errorf("some unknwon error")})
	response := SendHttpRequestWithBody(t, http.MethodGet, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(""), controller.HandleGetStatefulSessions)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	srv.EnqueueResponse("FindAll", responseMock{err: nil, specs: expectedSessions})
	response = SendHttpRequestWithBody(t, http.MethodGet, "/api/v3/stateful-session", "/api/v3/stateful-session",
		bytes.NewBufferString(""), controller.HandleGetStatefulSessions)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bodeBytes, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	var actualSessions []*dto.StatefulSession
	assert.Nil(t, json.Unmarshal(bodeBytes, &actualSessions))
	assert.Equal(t, 2, len(actualSessions))
	assert.True(t, statefulSessionsEqual(expectedSessions[0], actualSessions[0]))
	assert.True(t, statefulSessionsEqual(expectedSessions[1], actualSessions[1]))
}

func statefulSessionsEqual(one, another *dto.StatefulSession) bool {
	if one == another {
		return true
	}
	if one.Cluster != another.Cluster || one.Namespace != another.Namespace || one.Version != another.Version || one.Hostname != another.Hostname {
		return false
	}
	if one.Port == nil {
		if another.Port != nil {
			return false
		}
	} else if another.Port == nil || *one.Port != *another.Port {
		return false
	}
	if one.IsEnabled() != another.IsEnabled() {
		return false
	}
	if one.IsDeleteRequest() != another.IsDeleteRequest() {
		return false
	}
	oneCookie := one.Cookie
	anotherCookie := another.Cookie
	if oneCookie == nil {
		if anotherCookie != nil {
			return false
		}
	} else {
		if anotherCookie == nil {
			return false
		}
		if oneCookie.Name != anotherCookie.Name {
			return false
		}
	}
	oneRoute := one.Route
	anotherRoute := another.Route
	if oneRoute == nil {
		if anotherRoute != nil {
			return false
		}
	} else {
		if anotherRoute == nil {
			return false
		}
		if oneRoute.Prefix.String != anotherRoute.Prefix.String {
			return false
		}
	}
	return true
}
