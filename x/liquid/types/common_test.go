package types_test

import (
	"os"
	"testing"

	"github.com/zeta-protocol/black/app"
)

func TestMain(m *testing.M) {
	app.SetSDKConfig()
	os.Exit(m.Run())
}
