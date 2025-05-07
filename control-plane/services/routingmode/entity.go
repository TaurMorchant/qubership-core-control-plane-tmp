package routingmode

import (
	"bytes"
	"encoding/json"
)

type RoutingMode int

const (
	SIMPLE RoutingMode = iota
	NAMESPACED
	VERSIONED
	MIXED
)

var toString = map[RoutingMode]string{
	SIMPLE:     "SIMPLE",
	NAMESPACED: "NAMESPACED",
	VERSIONED:  "VERSIONED",
	MIXED:      "MIXED",
}

var toID = map[string]RoutingMode{
	"SIMPLE":     SIMPLE,
	"NAMESPACED": NAMESPACED,
	"VERSIONED":  VERSIONED,
	"MIXED":      MIXED,
}

func (rm RoutingMode) String() string {
	return toString[rm]
}

func (rm RoutingMode) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[rm])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (rm *RoutingMode) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*rm = toID[j]
	return nil
}
