package bus

import (
	"github.com/hashicorp/go-uuid"
	"google.golang.org/grpc/metadata"
	"sync"
)

const MetadataNodeIdKey = "Node-Id"

type nodeIdHolder struct {
	nodeId string
	sync.Mutex
}

var currentNodeId = nodeIdHolder{}

func GetMetadataWithNodeId() (metadata.MD, error) {
	return metadata.New(map[string]string{MetadataNodeIdKey: getOrConstructNodeId()}), nil
}

func getOrConstructNodeId() string {
	currentNodeId.Lock()
	defer currentNodeId.Unlock()
	if len(currentNodeId.nodeId) == 0 {
		nodeUUID, err := uuid.GenerateUUID()
		if err != nil {
			log.Errorf("cannot construct node id: %v", err)
			panic(err)
		}
		currentNodeId.nodeId = nodeUUID
	}
	return currentNodeId.nodeId
}

func ExtractNodeId(meta metadata.MD) string {
	nodeIds := meta.Get(MetadataNodeIdKey)
	if nodeIds != nil && len(nodeIds) > 0 {
		return nodeIds[0]
	}
	return ""
}
