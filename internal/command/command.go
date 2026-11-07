package command

import (
	"context"

	"github.com/ColinEge/8bd-kbd-api/protocol"
)

// Commands always start with 0x52
const cmdPrefix = protocol.DirOut

type ReadStrategy func(ctx context.Context, read func(context.Context) ([]byte, error)) ([][]byte, error)

type Command interface {
	Build() [][]byte
	Name() string

	ReadStrategy() ReadStrategy
}

type packets = [][]byte

// Responses always start with 0x54
const resPrefix = protocol.DirIn

type Response interface {
	SetPayload(p packets)
}
