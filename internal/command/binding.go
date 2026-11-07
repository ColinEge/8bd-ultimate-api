package command

import (
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
	"github.com/ColinEge/8bd-ultimate-api/protocol"
)

// BindAction
type BindAction struct {
	Key        keydefs.KeyID
	ActionCode keydefs.ActionCode
}

func (c *BindAction) Name() string {
	return "BindAction"
}

func (c *BindAction) Build() packets {
	commandBytes := []byte{0x03, 0x0c, 0x00, 0xaa, 0x09, 0x71}

	h := protocol.Header(protocol.DirOut, protocol.CommandBindAction, commandBytes...)
	h = append(h, byte(c.Key))
	for _, b := range c.ActionCode {
		h = append(h, b)
	}
	protocol.PadRight(&h)
	return [][]byte{h}
}

func (c *BindAction) ReadStrategy() ReadStrategy {
	return ReadSingle()
}
