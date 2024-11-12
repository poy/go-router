package cli_test

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/observability"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

var testOutput bytes.Buffer

func TestMain(m *testing.M) {
	log.SetFlags(0)
	log.SetOutput(&testOutput)
	os.Exit(m.Run())
}

func TestLogger_WithoutFields(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log.Infof("hello %s", "world")
	log.Warnf("warn %s", "world")

	scanner := bufio.NewScanner(&testOutput)
	scanner.Scan()
	if actual, expected := scanner.Text(), fmt.Sprintf("%s[INFO] %shello world", green, reset); actual != expected {
		t.Errorf("got %s, want %s", actual, expected)
	}
	scanner.Scan()
	if actual, expected := scanner.Text(), fmt.Sprintf("%s[WARN] %swarn world", red, reset); actual != expected {
		t.Errorf("got %s, want %s", actual, expected)
	}
}

func TestLogger_WithFields_NewInstance(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)

	// We're creating a new instance of the logger here that shouldn't affect
	// the logger above.
	_ = log.WithField("key", "value")

	log.Infof("hello %s", "world")
	log.Warnf("warn %s", "world")

	scanner := bufio.NewScanner(&testOutput)
	scanner.Scan()
	if actual, expected := scanner.Text(), fmt.Sprintf("%s[INFO] %shello world", green, reset); actual != expected {
		t.Errorf("got %s, want %s", actual, expected)
	}
	scanner.Scan()
	if actual, expected := scanner.Text(), fmt.Sprintf("%s[WARN] %swarn world", red, reset); actual != expected {
		t.Errorf("got %s, want %s", actual, expected)
	}
}

func TestLogger_WithFields(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("key1", "value1")
	log = log.WithField("key2", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey1%s = value1
  %skey2%s = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got %s, want %s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_AdjustsKeyWidth(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("key1", "value1")
	log = log.WithField("long-key", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey1%s     = value1
  %slong-key%s = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got %s, want %s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_WrapsWithSpaces(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("key1", "long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value long-value")
	log = log.WithField("long-key", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey1%s     = long-value long-value long-value long-value long-value long-value
             long-value long-value long-value long-value long-value long-value
             long-value long-value long-value long-value long-value long-value
             long-value long-value long-value long-value long-value
  %slong-key%s = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got:\n%s\n\nwant:\n%s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_WrapsWithTabs(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("key1", "long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value	long-value")
	log = log.WithField("long-key", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey1%s     = long-value  long-value  long-value  long-value  long-value
             long-value  long-value  long-value  long-value  long-value
             long-value  long-value  long-value  long-value  long-value
             long-value  long-value  long-value  long-value  long-value
             long-value  long-value  long-value
  %slong-key%s = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got:\n%s\n\nwant:\n%s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_WrapsWithNewLines(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("key1", "long-value\nlong-value")
	log = log.WithField("long-key", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey1%s     = long-value
             long-value
  %slong-key%s = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got:\n%s\n\nwant:\n%s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_WrapsWithContiguousLines(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("long-key", "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
	log = log.WithField("key2", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %skey2%s     = value2
  %slong-key%s = 0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got:\n%s\n\nwant:\n%s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func TestLogger_WithFields_EmptyValue(t *testing.T) {
	testOutput.Reset()

	ctx := injectiontesting.WithTesting(t)
	log := injection.Resolve[observability.Logger](ctx)
	log = log.WithField("empty-key", "")
	log = log.WithField("key", "value2")

	log.Infof("hello %s", "world")

	expected := fmt.Sprintf(`%s[INFO] %shello world
  %sempty-key%s = <empty>
  %skey%s       = value2
`, green, reset, yellow, reset, yellow, reset)

	if actual := testOutput.String(); actual != expected {
		t.Errorf("got:\n%s\n\nwant:\n%s", replaceWhiteSpace(actual), replaceWhiteSpace(expected))
	}
}

func replaceWhiteSpace(s string) string {
	s = strings.ReplaceAll(s, " ", ".")
	s = strings.ReplaceAll(s, "\n", "-")
	return s
}
