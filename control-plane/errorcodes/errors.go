package errorcodes

import (
	"encoding/json"
	errs "github.com/netcracker/qubership-core-lib-go-error-handling/v3/errors"
	"github.com/netcracker/qubership-core-lib-go-error-handling/v3/tmf"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/http"
)

var logger logging.Logger

func init() {
	logger = logging.GetLogger("errors")
}

var tmfConverter = tmf.DefaultConverter{}

func NewCpError(errCode cpRestErrorCode, detail string, cause error) *CpErrCodeError {
	return &CpErrCodeError{
		ErrCodeError:    errs.NewError(errCode.ErrorCode, detail, cause),
		httpCode:        errCode.httpCode,
		printStackTrace: errCode.printStackTrace,
	}
}

func NewRestErrorWithMeta(errCode errs.ErrorCode, detail string, statusCode int, meta interface{}) *CpErrCodeError {
	var metaMap map[string]interface{}
	res, _ := json.Marshal(meta)
	if err := json.Unmarshal(res, &metaMap); err != nil {
		metaMap = map[string]interface{}{"message": "Can not parse error details: " + err.Error()}
	}

	return &CpErrCodeError{
		ErrCodeError:    errs.NewError(errCode, detail, nil),
		httpCode:        statusCode,
		meta:            &metaMap,
		printStackTrace: false,
	}
}

func NewRemoteRestErrorOrNil(response *fasthttp.Response) *errs.ErrCodeError {
	if response.StatusCode() < fasthttp.StatusBadRequest {
		return nil
	}

	tmfResponse := tmf.Response{}
	err := json.Unmarshal(response.Body(), &tmfResponse)
	if err != nil {
		logger.Debugf("Response not in TMF format")
		return nil
	}

	return errs.NewError(RemoteRequestError, "Received error response in NC TMF format", tmfConverter.BuildErrorCodeError(tmfResponse))
}

func GetCpErrCodeErrorOrNil(err error) *CpErrCodeError {
	var cpErrorCodeError *CpErrCodeError
	if errors.As(err, &cpErrorCodeError) {
		return cpErrorCodeError
	}
	var errorCodeError *errs.ErrCodeError
	if errors.As(err, &errorCodeError) {
		return NewCpError(cpRestErrorCode{errorCodeError.ErrorCode, http.StatusInternalServerError, true}, errorCodeError.Detail, errorCodeError.Cause)
	}
	return nil
}
