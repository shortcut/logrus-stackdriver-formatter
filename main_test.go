package stackdriver

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	skipTimestamp = true
	os.Exit(m.Run())
}
