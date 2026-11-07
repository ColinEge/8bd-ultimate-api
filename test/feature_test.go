package test

import (
	"os"
	"testing"

	"github.com/ColinEge/8bd-kbd-api/test/steps"
	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "8bd-kbd",
		ScenarioInitializer: steps.InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"./features"},
			Tags:     "@api",
			Strict:   true,
			TestingT: t,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("Godog failed with status %d", status)
	}
}

// Planned application/host requirements must never pass through empty steps.
func TestPlannedFeatures(t *testing.T) {
	if os.Getenv("TEST_8BITDO_PLANNED") != "1" {
		t.Skip("planned requirements are not implemented; set TEST_8BITDO_PLANNED=1 to report them as pending")
	}
	suite := godog.TestSuite{
		Name:                "planned-requirements",
		ScenarioInitializer: steps.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty", Paths: []string{"./features"}, Tags: "@planned",
			Strict: true, TestingT: t,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("planned features remain pending (status %d)", status)
	}
}

func TestAPIBindingFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "public-binding-api",
		ScenarioInitializer: steps.InitializeAPIBindingScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"./features/api_rebinding.feature"},
			Strict:   true,
			TestingT: t,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("Godog failed with status %d", status)
	}
}
