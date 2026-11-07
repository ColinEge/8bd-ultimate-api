package steps

import "github.com/cucumber/godog"

// Macro capabilities are not exposed until the device protocol is verified.
func registerMacrosSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^a macro named "([^"]*)" exists$`, w.aMacroNamedXExists)
	sc.Step(`^I am recording a macro named "([^"]*)"$`, w.iAmRecordingAMacroNamedX)
	sc.Step(`^a macro contains multiple inputs with delays$`, w.aMacroContainsMultipleInputsWithDelays)
	sc.Step(`^a macro is assigned to a keyboard key$`, w.aMacroIsAssignedToAKeyboardKey)
	sc.Step(`^I have created and saved a macro$`, w.iHaveCreatedAndSavedAMacro)
	sc.Step(`^I press Ctrl$`, w.iPressCtrl)
	sc.Step(`^I press C$`, w.iPressC)
	sc.Step(`^I stop recording$`, w.iStopRecording)
	sc.Step(`^I create a macro named "([^"]*)"$`, w.iCreateAMacroNamedX)
	sc.Step(`^I rename the macro to "([^"]*)"$`, w.iRenameTheMacroToX)
	sc.Step(`^I delete the macro$`, w.iDeleteTheMacro)
	sc.Step(`^I save the macro$`, w.iSaveTheMacro)
	sc.Step(`^I change its input sequence$`, w.iChangeItsInputSequence)
	sc.Step(`^I assign the macro to a keyboard key$`, w.iAssignTheMacroToAKeyboardKey)
	sc.Step(`^I remove the assignment$`, w.iRemoveTheAssignment)
	sc.Step(`^the macro should exist$`, w.theMacroShouldExist)
	sc.Step(`^the macro should be named "([^"]*)"$`, w.theMacroShouldBeNamedX)
	sc.Step(`^the macro should no longer exist$`, w.theMacroShouldNoLongerExist)
	sc.Step(`^the macro should contain the recorded inputs in order$`, w.theMacroShouldContainTheRecordedInputsInOrder)
	sc.Step(`^the delays should be preserved$`, w.theDelaysShouldBePreserved)
	sc.Step(`^the updated sequence should be saved$`, w.theUpdatedSequenceShouldBeSaved)
	sc.Step(`^the key should reference that macro$`, w.theKeyShouldReferenceThatMacro)
	sc.Step(`^the key should no longer reference the macro$`, w.theKeyShouldNoLongerReferenceTheMacro)
	sc.Step(`^the macro should still exist$`, w.theMacroShouldStillExist)
}

func (w *World) aMacroNamedXExists(string) error                      { return godog.ErrPending }
func (w *World) iAmRecordingAMacroNamedX(string) error                { return godog.ErrPending }
func (w *World) aMacroContainsMultipleInputsWithDelays() error        { return godog.ErrPending }
func (w *World) aMacroIsAssignedToAKeyboardKey() error                { return godog.ErrPending }
func (w *World) iHaveCreatedAndSavedAMacro() error                    { return godog.ErrPending }
func (w *World) iPressCtrl() error                                    { return godog.ErrPending }
func (w *World) iPressC() error                                       { return godog.ErrPending }
func (w *World) iStopRecording() error                                { return godog.ErrPending }
func (w *World) iCreateAMacroNamedX(string) error                     { return godog.ErrPending }
func (w *World) iRenameTheMacroToX(string) error                      { return godog.ErrPending }
func (w *World) iDeleteTheMacro() error                               { return godog.ErrPending }
func (w *World) iSaveTheMacro() error                                 { return godog.ErrPending }
func (w *World) iChangeItsInputSequence() error                       { return godog.ErrPending }
func (w *World) iAssignTheMacroToAKeyboardKey() error                 { return godog.ErrPending }
func (w *World) iRemoveTheAssignment() error                          { return godog.ErrPending }
func (w *World) theMacroShouldExist() error                           { return godog.ErrPending }
func (w *World) theMacroShouldBeNamedX(string) error                  { return godog.ErrPending }
func (w *World) theMacroShouldNoLongerExist() error                   { return godog.ErrPending }
func (w *World) theMacroShouldContainTheRecordedInputsInOrder() error { return godog.ErrPending }
func (w *World) theDelaysShouldBePreserved() error                    { return godog.ErrPending }
func (w *World) theUpdatedSequenceShouldBeSaved() error               { return godog.ErrPending }
func (w *World) theKeyShouldReferenceThatMacro() error                { return godog.ErrPending }
func (w *World) theKeyShouldNoLongerReferenceTheMacro() error         { return godog.ErrPending }
func (w *World) theMacroShouldStillExist() error                      { return godog.ErrPending }
