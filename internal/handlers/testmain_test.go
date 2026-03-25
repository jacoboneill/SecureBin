package handlers

import (
	"os"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.SilenceLogs()
	os.Exit(m.Run())
}
