package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ColinEge/8bd-kbd-api/protocol"
)

var ErrMalformedResponse = errors.New("malformed command response")

// ReadSingle
func ReadSingle() ReadStrategy {
	return func(ctx context.Context, read func(context.Context) ([]byte, error)) ([][]byte, error) {

		resp, err := read(ctx)
		if err != nil {
			return nil, err
		}

		return [][]byte{resp}, nil
	}
}

// ReadContinuation
func ReadContinuation(contIndex int) ReadStrategy {
	return func(ctx context.Context, read func(context.Context) ([]byte, error)) ([][]byte, error) {

		var packets [][]byte

		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			resp, err := read(ctx)
			if err != nil {
				return nil, err
			}

			packets = append(packets, resp)

			if len(resp) != protocol.ReportSize || resp[0] != byte(protocol.DirIn) || resp[1] != byte(protocol.CommandReadRemapedKeys) {
				return nil, fmt.Errorf("%w: invalid mapped-keys response", ErrMalformedResponse)
			}
			if resp[contIndex] == 0 {
				break
			}
		}

		return packets, nil
	}
}

// ReadText
func ReadText(paramOffset int) ReadStrategy {
	return func(ctx context.Context, read func(context.Context) ([]byte, error)) ([][]byte, error) {

		var packets [][]byte

		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			resp, err := read(ctx)
			if err != nil {
				return nil, err
			}

			if paramOffset < 0 || len(resp) < 4+paramOffset || len(resp) > protocol.ReportSize ||
				int(resp[2+paramOffset])+4+paramOffset > len(resp) || resp[3+paramOffset] > 1 {
				return nil, fmt.Errorf("%w: invalid text response", ErrMalformedResponse)
			}
			packets = append(packets, resp)
			remaining := resp[3+paramOffset]
			if remaining == 0 {
				break
			}
		}

		return packets, nil
	}
}
