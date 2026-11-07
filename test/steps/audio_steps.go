package steps

import (
	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/cucumber/godog"
)

func registerAudioSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^system audio is muted$`, w.systemAudioIsMuted)
	sc.Step(`^I have configured audio actions$`, w.iHaveConfiguredAudioActions)
	sc.Step(`^I trigger the volume up action$`, w.iTriggerTheVolumeUpAction)
	sc.Step(`^I trigger the volume down action$`, w.iTriggerTheVolumeDownAction)
	sc.Step(`^I trigger the mute action$`, w.iTriggerTheMuteAction)
	sc.Step(`^I trigger the unmute action$`, w.iTriggerTheUnmuteAction)
	sc.Step(`^I apply the configuration to the keyboard$`, w.iApplyTheConfiguration)
	sc.Step(`^the system volume should increase$`, w.theSystemVolumeShouldIncrease)
	sc.Step(`^the system volume should decrease$`, w.theSystemVolumeShouldDecrease)
	sc.Step(`^system audio should be muted$`, w.systemAudioShouldBeMuted)
	sc.Step(`^system audio should be unmuted$`, w.systemAudioShouldBeUnmuted)
	sc.Step(`^the keyboard should use the updated audio configuration$`, w.theConfigurationShouldBeWrittenToTheKeyboard)
}

func (w *World) iHaveConfiguredAudioActions() {
	w.prepareChanges([]client.BindingChange{
		{Key: client.KeyA, Action: client.Action{0x0c, 0xe9, 0}},
		{Key: client.KeyB, Action: client.Action{0x0c, 0xea, 0}},
		{Key: client.KeyC, Action: client.Action{0x0c, 0xe2, 0}},
	})
	// Use the public helpers to build input; the script above is the independent oracle.
	w.Changes = []client.BindingChange{
		{Key: client.KeyA, Action: client.ActionVolumeUp()},
		{Key: client.KeyB, Action: client.ActionVolumeDown()},
		{Key: client.KeyC, Action: client.ActionMute()},
	}
}

func (w *World) systemAudioIsMuted() error            { return godog.ErrPending }
func (w *World) iTriggerTheVolumeUpAction() error     { return godog.ErrPending }
func (w *World) iTriggerTheVolumeDownAction() error   { return godog.ErrPending }
func (w *World) iTriggerTheMuteAction() error         { return godog.ErrPending }
func (w *World) iTriggerTheUnmuteAction() error       { return godog.ErrPending }
func (w *World) theSystemVolumeShouldIncrease() error { return godog.ErrPending }
func (w *World) theSystemVolumeShouldDecrease() error { return godog.ErrPending }
func (w *World) systemAudioShouldBeMuted() error      { return godog.ErrPending }
func (w *World) systemAudioShouldBeUnmuted() error    { return godog.ErrPending }
