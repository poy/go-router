package cli

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/poy/go-dependency-injection/pkg/injection"
	"github.com/poy/go-router/pkg/observability"
)

func init() {
	injection.Register[observability.Logger](func(ctx context.Context) observability.Logger {
		return logger{}
	})
}

type logger struct {
	fields      map[string]string
	builtFields string
}

func (l logger) WithField(name, value string) observability.Logger {
	if value == "" {
		value = "<empty>"
	}

	m := make(map[string]string)
	for k, v := range l.fields {
		m[k] = v
	}
	m[name] = value

	log := logger{
		fields: m,
	}
	log.builtFields = log.buildFields()

	return log
}

func (l logger) buildFields() string {
	maxKeyLength := 0
	keys := make([]string, 0, len(l.fields))
	for k := range l.fields {
		keys = append(keys, k)
		if len(k) > maxKeyLength {
			maxKeyLength = len(k)
		}
	}
	sort.Strings(keys)

	var results []string
	for _, k := range keys {
		prefix := "  " + yellow + k + reset + strings.Repeat(" ", maxKeyLength-len(k)+1) + "= "
		// We don't want to include the color codes for lengths.
		prefixLen := len(prefix) - len(yellow) - len(reset)

		value := l.normalizeValue(l.fields[k])
		body := l.wrapString(value, 80-prefixLen)
		results = append(results, prefix+body[0])
		for _, b := range body[1:] {
			results = append(results, strings.Repeat(" ", prefixLen)+b)
		}
	}

	return strings.Join(results, "\n")
}

func (logger) normalizeValue(s string) string {
	s = strings.ReplaceAll(s, "\t", "  ")
	return s
}

func (logger) wrapString(s string, l int) []string {
	var lines []string
	currentLine := ""
	lastWhitespace := -1
	for i := range s {
		if len(currentLine) >= l {
			if lastWhitespace != -1 {
				lines = append(lines, currentLine[:lastWhitespace])
				currentLine = currentLine[lastWhitespace+1:]
				lastWhitespace = -1
			}
			currentLine += string(s[i])
		} else {
			if s[i] == ' ' {
				lastWhitespace = len(currentLine)
			} else if s[i] == '\n' {
				lines = append(lines, currentLine)
				currentLine = ""
				lastWhitespace = -1
				continue
			}
			currentLine += string(s[i])
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}

	return lines
}

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

// Fatalf implements Logger.
func (l logger) Fatalf(format string, args ...any) {
	s := fmt.Sprintf(red+"[FATAL] "+reset+format, args...)
	log.Fatal(s + "\n" + l.builtFields)
}

// Warnf implements Logger.
func (l logger) Warnf(format string, args ...any) {
	s := fmt.Sprintf(red+"[WARN] "+reset+format, args...)
	log.Print(s + "\n" + l.builtFields)
}

// Infof implements Logger.
func (l logger) Infof(format string, args ...any) {
	s := fmt.Sprintf(green+"[INFO] "+reset+format, args...)
	log.Print(s + "\n" + l.builtFields)
}
