package grpc

import (
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"google.golang.org/genproto/googleapis/rpc/status"
	"testing"
)

const string500char = "sdsadsadsadsadsadsadsadsadjkasldhskajdhaklhweufhlkahfjldhflkadhsfdshfkudshfkuldsahflkadshfklduashfkudshfkuldshaflkdshaflkdshfkudhslkfuhdskufldskufhdsklafhdlskafhlkdshfkuldshfkludshflkdsahflkdsahflkdshfkludshfkldshfkludshflksdahflkdsnvjldnvlcxvjhdflkdsahflauefhleukhfaeklsuhflkahdkjfndasjklvhdsfkjhdsklfkuseahfeukslahfkluseahfkusfhjkahfdjksfhjkldshfdklshfkjdlshfkldsahfkjdshflkdjsanvlkjdvdhafkjhekufhuklsfhkluasfhekfhdjkfhdkljsahfkleuahfueklshflkseuahfekulhflaksfheksfhesuafueshfklsahfjkdhfaksehdjds"

func TestDebugCallbacks_OnStreamResponse(t *testing.T) {
	callbacks := &DebugCallbacks{logging.GetLogger("test callbacks")}
	request := &discovery.DiscoveryRequest{
		VersionInfo:   "",
		Node:          nil,
		ResourceNames: nil,
		TypeUrl:       "",
		ResponseNonce: "",
		ErrorDetail:   nil,
	}
	response := &discovery.DiscoveryResponse{
		VersionInfo:  "",
		Resources:    nil,
		Canary:       false,
		TypeUrl:      "",
		Nonce:        "",
		ControlPlane: nil,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Code mustn't throw panic, cause: %v", r)
		}
	}()

	callbacks.OnStreamResponse(ctx, 0, nil, nil)

	request.ErrorDetail = &status.Status{Message: string500char}
	callbacks.OnStreamResponse(ctx, 0, request, response)

	request.ErrorDetail = &status.Status{Message: "small string"}
	callbacks.OnStreamResponse(ctx, 0, request, response)

	request.ErrorDetail = &status.Status{Message: string500char + "1"}
}
