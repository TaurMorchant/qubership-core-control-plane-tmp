package tm

import (
	"github.com/gorilla/websocket"
	go_stomp_websocket "github.com/netcracker/qubership-core-lib-go-stomp-websocket/v3"
	"github.com/netcracker/qubership-core-lib-go/v3/utils"
	"net/http"
	"net/url"
)

type SocketClient struct {
	stompClient *go_stomp_websocket.StompClient
	sub         *go_stomp_websocket.Subscription
}

func NewSocketClient() *SocketClient {
	return &SocketClient{}
}

func (client *SocketClient) ConnectAndSubscribe(socketAddress, stompTopic string) (*go_stomp_websocket.Subscription, error) {
	err := client.connect(socketAddress)
	if err != nil {
		logger.ErrorC(ctx, "Can't connect to tenant-manager: \n %v", err)
		return nil, err
	}
	sub, err := client.subscribeOnTopic(stompTopic)
	if err != nil {
		logger.ErrorC(ctx, "Can't subscribe on topic %v: \n %v", stompTopic, err)
		return nil, err
	}
	return sub, nil
}

func (client *SocketClient) UnsubscribedAndDisconnect(sub *go_stomp_websocket.Subscription) error {
	client.unSubscribe(sub)
	err := client.disconnect()
	if err != nil {
		logger.ErrorC(ctx, "Can't close websocket %v", err)
		return err
	}
	logger.InfoC(ctx, "Connection to socket is closed successful")
	return nil
}

func (client *SocketClient) connect(socketAddress string) error {
	parsedUrl, err := url.Parse(socketAddress)
	if err != nil {
		logger.ErrorC(ctx, "Could not parse websocket URL %s:\n %v", socketAddress, err)
		return err
	}
	dialer := websocket.Dialer{TLSClientConfig: utils.GetTlsConfig()}

	logger.InfoC(ctx, "Connecting to tenant manager %s", parsedUrl.String())
	stompClient, err := go_stomp_websocket.Connect(*parsedUrl, dialer, http.Header{}, ConnectionDial{})
	if err != nil {
		logger.ErrorC(ctx, "Can't connect to %v \n %v", parsedUrl.String(), err)
		return err
	}
	logger.InfoC(ctx, "Connected to tenant manager %s", parsedUrl.String())
	client.stompClient = stompClient
	return nil
}

func (client *SocketClient) subscribeOnTopic(stompTopic string) (*go_stomp_websocket.Subscription, error) {
	logger.InfoC(ctx, "Trying to subscribe on topic %s to get information about tenants", stompTopic)
	sub, err := client.stompClient.Subscribe(stompTopic)
	if err != nil {
		logger.ErrorC(ctx, "Can't subscribe on topic %v: \n %v", stompTopic, err)
		return nil, err
	}
	logger.InfoC(ctx, "Successfully subscribed on topic %s to get information about tenants", stompTopic)
	return sub, nil
}

func (client *SocketClient) disconnect() error {
	err := client.stompClient.Disconnect()
	if err != nil {
		logger.InfoC(ctx, "Can't disconnect socket \n %v", err)
		return err
	}
	logger.InfoC(ctx, "Disconnected from socket")
	return nil
}

func (client *SocketClient) unSubscribe(sub *go_stomp_websocket.Subscription) {
	sub.Unsubscribe()
	logger.InfoC(ctx, "Unsubscribed from topic %s ", sub.Topic)
}
