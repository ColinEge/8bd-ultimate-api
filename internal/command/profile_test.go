package command_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ColinEge/8bd-ultimate-api/internal/command"
	"github.com/ColinEge/8bd-ultimate-api/protocol"
)

func TestBuildProfilePacketsHeaders(t *testing.T) {
	for length := 1; length <= 40; length++ {
		// Generate profile name: "", "1", "12", "123", ... "1234567890123456"
		name := ""
		for i := 0; i < length; i++ {
			if i != 0 {
				name += strconv.Itoa((i) % 10)
			}
		}

		input := name
		cmd := &command.SetProfile{ProfileName: input}
		packets := cmd.Build()

		fmt.Printf("Command: %s\n", cmd.Name())
		output := protocol.UTF16BEDecode(protocol.ReadText(packets, 0))
		if input != output {
			t.Fail()
		}
	}
}
