package clustering

import (
	"bytes"
	"fmt"
)

type Role int

const (
	Initial Role = iota
	Slave
	Master
	Phantom
)

func (r Role) String() string {
	switch r {
	case Slave:
		return "Slave"
	case Master:
		return "Master"
	case Phantom:
		return "Phantom"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

func (s Role) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}
