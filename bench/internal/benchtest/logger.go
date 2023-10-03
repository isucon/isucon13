package benchtest

// suppress
type nopLogger struct{}

func (s nopLogger) Print(args ...interface{}) {
	// Do nothing
}
