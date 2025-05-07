package clustering

import (
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/uptrace/bun"
	"net"
	"strconv"
	"time"
)

var log = logging.GetLogger("clustering")

const ElectionTableName = "election"

type ElectionService interface {
	GetMaster() (*MasterMetadata, error)
	ResetSyncClock(master string) error
	TryWriteAsMaster(*MasterMetadata) bool
	ShiftSyncClock(time.Duration) error
	DeleteSeveralRecordsFromDb() error
}

type MasterMetadata struct {
	bun.BaseModel `bun:"election,alias:t"`

	Id        int64     `bun:"type:serial,notnull"`
	Name      string    `bun:",notnull"`
	NodeInfo  NodeInfo  `bun:"type:jsonb"`
	SyncClock time.Time `bun:"type:timestamp,notnull"`
	Namespace string    `bun:"namespace"`
}

type NodeInfo struct {
	IP       string
	SWIMPort uint16
	BusPort  uint16
	HttpPort uint16
}

func (n *NodeInfo) SWIMAddress() string {
	return makeAddress(n.IP, n.SWIMPort)
}

func (n *NodeInfo) BusAddress() string {
	return makeAddress(n.IP, n.BusPort)
}

func (n *NodeInfo) GetHttpAddress() string {
	return makeAddress(n.IP, n.HttpPort)
}

func (i *MasterMetadata) Equals(record *MasterMetadata) bool {
	if i == nil || record == nil {
		return false
	}
	return i.Id == record.Id &&
		i.Name == record.Name &&
		i.NodeInfo == record.NodeInfo
}

func (i *MasterMetadata) EqualsByNameAndNodeInfo(record *MasterMetadata) bool {
	if i == nil || record == nil {
		return false
	}
	return i.Name == record.Name && i.NodeInfo == record.NodeInfo
}

func makeAddress(addr string, port uint16) string {
	return net.JoinHostPort(addr, strconv.Itoa(int(port)))
}
