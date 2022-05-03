package tests

import (
	"github.com/cucumber/godog"
	"testing"
)

func runTestSuite(t *testing.T, init func(s *godog.ScenarioContext)) {
	suite := godog.TestSuite{
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format:        "pretty",
			Paths:         []string{"features"},
			TestingT:      t,
			Randomize:     -1,
			StopOnFailure: true,
			Strict:        true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
