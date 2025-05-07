package bg

import (
	"os"
	"strings"
)

type Mode string

var mode Mode

const BlueGreen1 Mode = "BlueGreen1"
const BlueGreen2 Mode = "BlueGreen2"

func init() {
	ReloadMode()
}

func GetMode() Mode {
	return mode
}

func ReloadMode() {
	peerNamespace := strings.TrimSpace(os.Getenv("PEER_NAMESPACE"))
	if peerNamespace != "" {
		mode = BlueGreen2
	} else {
		mode = BlueGreen1
	}
}
