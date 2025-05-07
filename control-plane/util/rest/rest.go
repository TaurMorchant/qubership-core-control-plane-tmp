package rest

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/util"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/valyala/fasthttp"
)

type defaultClient struct {
}

var Client ClientRequest = &defaultClient{}

func (cl defaultClient) DoRetryRequest(logContext context.Context, method string, url string, data []byte, logger logging.Logger) (*fasthttp.Response, error) {
	return util.DoRetryRequest(logContext, method, url, data, logger)
}

func (cl defaultClient) DoRequest(logContext context.Context, method string, url string, data []byte, logger logging.Logger) (*fasthttp.Response, error) {
	return util.DoRequest(logContext, method, url, data, logger)
}

type ClientRequest interface {
	DoRetryRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error)
	DoRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error)
}
