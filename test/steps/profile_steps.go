package steps

import "github.com/cucumber/godog"

// Planned acceptance wiring; profile storage and apply have package-level tests.
// Whole-configuration deletion remains unsupported.
func registerProfileSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^profiles named "([^"]*)" and "([^"]*)" exist locally$`, w.profilesNamedXAndYExistLocally)
	sc.Step(`^I have created multiple profiles$`, w.iHaveCreatedMultipleProfiles)
	sc.Step(`^a profile file contains invalid configuration data$`, w.aProfileFileContainsInvalidConfigurationData)
	sc.Step(`^I create a profile named "([^"]*)"$`, w.iCreateAProfileNamedX)
	sc.Step(`^I request the available profiles$`, w.iRequestTheAvailableProfiles)
	sc.Step(`^I rename the profile to "([^"]*)"$`, w.iRenameTheProfileToX)
	sc.Step(`^I delete the profile$`, w.iDeleteTheProfile)
	sc.Step(`^I modify its key mappings$`, w.iModifyItsKeyMappings)
	sc.Step(`^I load the profile$`, w.iLoadTheProfile)
	sc.Step(`^I create another profile named "([^"]*)"$`, w.iCreateAnotherProfileNamedX)
	sc.Step(`^I attempt to load the profile$`, w.iAttemptToLoadTheProfile)
	sc.Step(`^I apply the "([^"]*)" profile$`, w.iApplyTheProfile)
	sc.Step(`^I read its configuration$`, w.iReadItsConfiguration)
	sc.Step(`^I delete its stored configuration$`, w.iDeleteItsStoredConfiguration)
	sc.Step(`^a profile file named "([^"]*)" should exist$`, w.aProfileFileNamedX)
	sc.Step(`^"([^"]*)" should be listed$`, w.xShouldBeListed)
	sc.Step(`^a profile named "([^"]*)" should exist$`, w.aProfileNamedXShouldExist)
	sc.Step(`^the profile file should no longer exist$`, w.theProfileFileShouldNoLongerExist)
	sc.Step(`^the updated configuration should be saved locally$`, w.theUpdatedConfigurationShouldBeSavedLocally)
	sc.Step(`^all profiles should still exist$`, w.allProfilesShouldStillExist)
	sc.Step(`^its configuration should be available in memory$`, w.itsConfigurationShouldBeAvailableInMemory)
	sc.Step(`^its configuration should be written to the keyboard$`, w.itsConfigurationShouldBeWrittenToTheKeyboard)
	sc.Step(`^the device configuration should be returned$`, w.theDeviceConfigurationShouldBeReturned)
	sc.Step(`^the device configuration should be deleted$`, w.theDeviceConfigurationShouldBeDeleted)
	sc.Step(`^the local profile should remain unchanged$`, w.theLocalProfileShouldRemainUnchanged)
	sc.Step(`^a profile named "([^"]*)" should no longer exist$`, w.aProfileNamedXShouldNoLongerExist)
	sc.Step(`^their configurations should be preserved$`, w.theirConfigurationsShouldBePreserved)
}

func (w *World) profilesNamedXAndYExistLocally(string, string) error { return godog.ErrPending }
func (w *World) iHaveCreatedMultipleProfiles() error                 { return godog.ErrPending }
func (w *World) aProfileFileContainsInvalidConfigurationData() error { return godog.ErrPending }
func (w *World) iCreateAProfileNamedX(string) error                  { return godog.ErrPending }
func (w *World) iRequestTheAvailableProfiles() error                 { return godog.ErrPending }
func (w *World) iRenameTheProfileToX(string) error                   { return godog.ErrPending }
func (w *World) iDeleteTheProfile() error                            { return godog.ErrPending }
func (w *World) iModifyItsKeyMappings() error                        { return godog.ErrPending }
func (w *World) iLoadTheProfile() error                              { return godog.ErrPending }
func (w *World) iCreateAnotherProfileNamedX(string) error            { return godog.ErrPending }
func (w *World) iAttemptToLoadTheProfile() error                     { return godog.ErrPending }
func (w *World) iApplyTheProfile(string) error                       { return godog.ErrPending }
func (w *World) iReadItsConfiguration() error                        { return godog.ErrPending }
func (w *World) iDeleteItsStoredConfiguration() error                { return godog.ErrPending }
func (w *World) aProfileFileNamedX(string) error                     { return godog.ErrPending }
func (w *World) xShouldBeListed(string) error                        { return godog.ErrPending }
func (w *World) aProfileNamedXShouldExist(string) error              { return godog.ErrPending }
func (w *World) theProfileFileShouldNoLongerExist() error            { return godog.ErrPending }
func (w *World) theUpdatedConfigurationShouldBeSavedLocally() error  { return godog.ErrPending }
func (w *World) allProfilesShouldStillExist() error                  { return godog.ErrPending }
func (w *World) itsConfigurationShouldBeAvailableInMemory() error    { return godog.ErrPending }
func (w *World) itsConfigurationShouldBeWrittenToTheKeyboard() error { return godog.ErrPending }
func (w *World) theDeviceConfigurationShouldBeReturned() error       { return godog.ErrPending }
func (w *World) theDeviceConfigurationShouldBeDeleted() error        { return godog.ErrPending }
func (w *World) theLocalProfileShouldRemainUnchanged() error         { return godog.ErrPending }
func (w *World) aProfileNamedXShouldNoLongerExist(string) error      { return godog.ErrPending }
func (w *World) theirConfigurationsShouldBePreserved() error         { return godog.ErrPending }
