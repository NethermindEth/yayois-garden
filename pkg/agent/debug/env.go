package debug

import "os"

const (
	DebugPlainSetupKey = "DEBUG_PLAIN_SETUP"
	DebugShowSetupKey  = "DEBUG_SHOW_SETUP"
)

func isDebugPlainSetupSet() bool {
	return os.Getenv(DebugPlainSetupKey) == "true"
}

func isDebugShowSetupSet() bool {
	return os.Getenv(DebugShowSetupKey) == "true"
}
