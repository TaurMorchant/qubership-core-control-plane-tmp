package bus

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"time"
	"trace-service/trace-service/domain"
)

var ErrUnsupportedEvent = errors.New("bus: unsupported event type")

func init() {
	gob.Register([]*domain.TraceResponse{})
}

//
// Functions to convert domain structures to gRPC protobuf events
//

func buildProtobufEvent(any interface{}) (*Event, error) {
	switch t := any.(type) {
	case []byte:
		return buildProtobufRawDataEvent(any.([]byte))
	case string:
		return buildProtobufRawDataEvent([]byte(any.(string)))
	default:
		log.Errorf("Event bus doesn't support events of type %T!", t)
		return nil, ErrUnsupportedEvent
	}
}

func buildTestSnapshotEvent() (*Event, error) {
	return buildProtobufRawDataEvent([]byte(fmt.Sprintf(`{"snapshotTime": "%v"}`, time.Now())))
}

func buildProtobufRawDataEvent(source []byte) (*Event, error) {
	marshalledEvent, err := ptypes.MarshalAny(&RawBytesData{Data: source})
	if err != nil {
		log.Errorf("Failed to marshal byte slice to protobuf event: %v", err)
		return nil, err
	}
	log.Debugf("gRPC binary message has been built. size: %d bytes", len(source))
	return &Event{EventType: EventType_RAW_DATA, Data: marshalledEvent}, nil
}

//
// Functions to parse domain structures from gRPC protobuf events
//

func readProtobufEvent(event *Event) (interface{}, error) {
	switch event.EventType {
	case EventType_RAW_DATA:
		return readRawBytesDataFromProto(event)
	default:
		log.Errorf("Event bus gRPC client doesn't support events of type %v!", event.EventType)
		return nil, ErrUnsupportedEvent
	}
}

func readRawBytesDataFromProto(source *Event) ([]byte, error) {
	var data RawBytesData
	if err := ptypes.UnmarshalAny(source.Data, &data); err != nil {
		log.Errorf("Failed to unmarshal event %v data: %v", source.EventType, err)
		return nil, err
	}
	return data.Data, nil
}
