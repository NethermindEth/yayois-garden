package debug

const (
	Debug = true
)

func IsDebug() bool {
	return Debug
}

func IsDebugPlainSetup() bool {
	return Debug && isDebugPlainSetupSet()
}

func IsDebugShowSetup() bool {
	return Debug && isDebugShowSetupSet()
}
