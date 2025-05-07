package msaddr

import (
	"github.com/netcracker/qubership-core-control-plane/tlsmode"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"strconv"
	"strings"
)

var logger = logging.GetLogger("microservice-address")

type MicroserviceAddress struct {
	Proto            string
	MicroserviceName string
	Host             string
	Port             int32
	Namespace        Namespace
}

func NewMicroserviceAddress(msUrl string, namespace string) *MicroserviceAddress {
	this := &MicroserviceAddress{}

	var portStr string
	this.Proto, this.Host, portStr = this.splitUrl(msUrl)
	this.Host = tlsmode.AdaptHostname(this.Host)
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Panicf("Failed to convert port %v to int: %v", portStr, err)
	}
	this.Port = int32(portInt)
	this.MicroserviceName = getMicroserviceName(this.Host)
	this.Namespace = Namespace{Namespace: namespace}
	return this
}

func (this *MicroserviceAddress) splitUrl(url string) (string, string, string) {
	proto, hostPortString := this.splitToProtoAndUrl(url)
	hostAndPortParsed := strings.Split(hostPortString, ":")

	if len(hostAndPortParsed) == 1 {
		return proto, hostAndPortParsed[0], this.getDefaultPort(proto)
	}
	return proto, hostAndPortParsed[0], hostAndPortParsed[1]
}

func (this *MicroserviceAddress) splitToProtoAndUrl(url string) (string, string) {
	if idx := strings.Index(url, "://"); idx == -1 {
		return "http", url
	} else {
		return strings.ToLower(url[:idx]), url[idx+3:]
	}
}
func (this *MicroserviceAddress) getDefaultPort(proto string) string {
	if proto == "https" {
		return "443"
	} else {
		return "80"
	}
}
func getMicroserviceName(url string) string {
	return strings.Split(url, ".")[0]
}

func (this *MicroserviceAddress) GetMicroserviceName() string {
	return this.MicroserviceName
}

func (this *MicroserviceAddress) GetNamespacedMicroserviceHost() string {
	if this.Namespace.IsCurrentNamespace() {
		return this.Host
	}

	//for localdev host name == namespace
	if this.Namespace.IsLocalDevNamespace() {
		return this.Namespace.Namespace
	}
	//for sandboxes host name = microservice-name.namespace
	return this.Host + "." + this.Namespace.Namespace
}

func (this *MicroserviceAddress) GetPort() int32 {
	return this.Port
}

func (this *MicroserviceAddress) GetProto() string {
	return this.Proto
}
