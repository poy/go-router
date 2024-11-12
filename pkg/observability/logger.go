package observability

// Logger is the interface that provides a logging abstraction.
type Logger interface {
	// TODO: Eventually, we should actually decide on a logger.
	// Debugf(format string, args ...any)
	// Errorf(format string, args ...any)

	Fatalf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)

	// WithField returns a new logger with the given field.
	WithField(name, value string) Logger
}
