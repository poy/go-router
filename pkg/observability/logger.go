package observability

import (
	"context"
	"log"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

func init() {
	injection.Register[Logger](func(ctx context.Context) Logger {
		return logger{}
	})
}

// Logger is the interface that provides a logging abstraction.
type Logger interface {
	// TODO: Eventually, we should actually decide on a logger.
	// Debugf(format string, args ...any)
	// Errorf(format string, args ...any)
	// Fatalf(format string, args ...any)

	Warnf(format string, args ...any)
	Infof(format string, args ...any)
}

type logger struct{}

// Warnf implements Logger.
func (l logger) Warnf(format string, args ...any) {
	log.Printf("[WARN] "+format, args...)
}

// Infof implements Logger.
func (l logger) Infof(format string, args ...any) {
	log.Printf("[INFO] "+format, args...)
}
