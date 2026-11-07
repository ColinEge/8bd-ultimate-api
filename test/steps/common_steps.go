package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func registerCommonSteps(ctx *godog.ScenarioContext, w *World) {
	ctx.Step(`^an 8BitDo keyboard is connected$`, w.an8BitDoKeyboardIsConnected)
	ctx.Step(`^the operation should fail$`, w.theOperationShouldFail)
	ctx.Step(`^an appropriate error should be returned$`, w.anAppropriateErrorShouldBeReturned)
	ctx.Step(`^I restart the application$`, w.iRestartTheApplication)
	ctx.Step(`^a profile named "([^"]*)" exists(?: (locally))?$`, w.aProfileNamedXExists)
}

func (w *World) an8BitDoKeyboardIsConnected() error {
	return w.Keyboard.Reset()
}

func (w *World) theOperationShouldFail() error {
	if w.Err == nil {
		return fmt.Errorf("operation succeeded; expected an error")
	}
	return nil
}

func (w *World) anAppropriateErrorShouldBeReturned() error {
	return w.theOperationShouldFail()
}

func (w *World) iRestartTheApplication() error {
	return godog.ErrPending
}

func (w *World) aProfileNamedXExists(x string, y string) error {
	return godog.ErrPending
}
