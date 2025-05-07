package dr

import (
	"os"
	"strings"
)

type Mode string

var mode Mode

const Active Mode = "Active"
const Standby Mode = "Standby"

func init() {
	ReloadMode()
}

func GetMode() Mode {
	return mode
}

func ReloadMode() {
	executionMode := os.Getenv("EXECUTION_MODE")
	if strings.EqualFold(executionMode, "standby") || strings.EqualFold(executionMode, "disabled") {
		mode = Standby
	} else {
		mode = Active
	}
}
