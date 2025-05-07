package tm

import (
	"github.com/gorilla/websocket"
	"github.com/netcracker/qubership-core-control-plane/util"
	"net/http"
	"net/url"
)

type ConnectionDial struct {
}

func (dial ConnectionDial) Dial(webSocketURL url.URL, dialer websocket.Dialer, requestHeaders http.Header) (*websocket.Conn, *http.Response, error) {
	return util.SecureWebSocketDial(ctx, webSocketURL, dialer, requestHeaders, logger)
}
