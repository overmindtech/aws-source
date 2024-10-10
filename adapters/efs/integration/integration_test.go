package integration

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"testing"
)

func Setup() error {
	fmt.Println("Setting up EFS integration tests")
	return nil
}

func Teardown(logger *slog.Logger) error {
	logger.Info("Tearing down EFS integration tests")
	return nil
}

func TestIntegrationEFSSomeAdapter(t *testing.T) {
	slog.Info("Running EFS integration test TestSomeSource")
}

func TestMain(m *testing.M) {
	if !shouldRunIntegrationTests() {
		slog.Warn("skipping integration tests.. set RUN_ALL_INTEGRATION_TESTS=true or individual RUN_EFS_INTEGRATION_TESTS=true to run them")
		os.Exit(0)
	}

	err := Setup()
	if err != nil {
		slog.Error("failed to setup integration test environment", "err", err)
		os.Exit(1)
	}

	slog.Info("Completed setup")

	code := m.Run()
	if code != 0 {
		slog.Warn("integration tests failed", "code", code)
	} else {
		slog.Info("integration tests passed", "code", code)
	}

	slog.Info("Running teardown")
	if err := Teardown(slog.Default()); err != nil {
		slog.Error("failed to teardown integration test environment", "err", err)
		os.Exit(1)
	}

	os.Exit(code)
}

// this can be in a more general package
func shouldRunIntegrationTests() bool {
	runAll, found := os.LookupEnv("RUN_ALL_INTEGRATION_TESTS")
	if found {
		shouldRunAll, err := strconv.ParseBool(runAll)
		if err != nil {
			return false
		}

		if shouldRunAll {
			return true
		}
	}

	runEFS, found := os.LookupEnv("RUN_EFS_INTEGRATION_TESTS")
	if found {
		shouldRunEFS, err := strconv.ParseBool(runEFS)
		if err != nil {
			return false
		}

		if shouldRunEFS {
			return true
		}
	}

	return false
}
