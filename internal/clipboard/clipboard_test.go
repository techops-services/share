package clipboard

import (
	"runtime"
	"testing"
)

func TestCopy(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("clipboard test only runs on macOS in CI")
	}

	err := Copy("share-cli-test-content")
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
}
