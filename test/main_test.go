package test

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Start shared container if SSH tests are enabled
	if os.Getenv("DOIT_TEST_CONTAINERS") != "" {
		if err := startSharedContainer(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start test container: %v\n", err)
			os.Exit(1)
		}
		defer stopSharedContainer()
	}

	os.Exit(m.Run())
}
