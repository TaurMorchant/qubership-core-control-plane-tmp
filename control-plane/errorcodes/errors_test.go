package errorcodes

import (
	"encoding/json"
	"fmt"
	gerrors "github.com/go-errors/errors"
	errs "github.com/netcracker/qubership-core-lib-go-error-handling/v3/errors"
	"github.com/netcracker/qubership-core-lib-go-error-handling/v3/tmf"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"net/http"
	"testing"
)

var testCpError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-0000", Title: "Test title"}, http.StatusInternalServerError, true}

func TestGetCpErrCodeErrorOrNil(t *testing.T) {
	expectedRootCause := NewCpError(testCpError, "test details", nil)

	rootCause := GetCpErrCodeErrorOrNil(expectedRootCause)
	assert.Equal(t, expectedRootCause, rootCause)

	rootCause = GetCpErrCodeErrorOrNil(errs.NewError(expectedRootCause.ErrorCode, expectedRootCause.Detail, expectedRootCause.Cause))
	assert.Equal(t, expectedRootCause.ErrorCode, rootCause.ErrorCode)
	assert.Equal(t, expectedRootCause.httpCode, rootCause.httpCode)

	rootCause = GetCpErrCodeErrorOrNil(fmt.Errorf("test error"))
	assert.Nil(t, rootCause)

	rootCause = GetCpErrCodeErrorOrNil(gerrors.WrapPrefix(expectedRootCause, "test wrap", 0))
	assert.Equal(t, expectedRootCause, rootCause)
}

func TestNewRemoteRestErrorOrNil(t *testing.T) {
	status := "400"
	tmfResponse := tmf.Response{
		Id:      "1",
		Code:    ValidationRequestError.Code,
		Status:  &status,
		Message: "test message",
	}

	out, err := json.Marshal(tmfResponse)
	assert.Nil(t, err)
	response := getHastHttpResponse(500, string(out))
	result := NewRemoteRestErrorOrNil(response)
	assert.NotNil(t, result)
	assert.Equal(t, RemoteRequestError, result.ErrorCode)
	assert.Equal(t, tmfResponse.Code, result.Cause.(*errs.RemoteErrCodeError).ErrorCode.Code)

	response = getHastHttpResponse(200, "test non TMF body")
	result = NewRemoteRestErrorOrNil(response)
	assert.Nil(t, result)

	response = getHastHttpResponse(500, "test non TMF body")
	result = NewRemoteRestErrorOrNil(response)
	assert.Nil(t, result)
}

func getHastHttpResponse(statusCode int, body string) *fasthttp.Response {
	responseHeader := fasthttp.ResponseHeader{}
	responseHeader.SetStatusCode(statusCode)
	response := &fasthttp.Response{
		Header: responseHeader,
	}
	response.SetBody([]byte(body))
	return response
}
