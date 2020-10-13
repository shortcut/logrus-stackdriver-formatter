package stackdriver

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFormatter(t *testing.T) {
	for _, tc := range formatterTests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer

			logger := logrus.New()
			logger.Out = &out
			logger.Formatter = NewFormatter(
				WithService("test"),
				WithVersion("0.1"),
			)

			tc.run(logger)

			var got map[string]interface{}
			json.Unmarshal(out.Bytes(), &got)

			require.True(t, reflect.DeepEqual(got, tc.out), "unexpected output = %# v; \n want = %# v; \n diff: %# v", pretty.Formatter(got), pretty.Formatter(tc.out), pretty.Diff(got, tc.out))
		})
	}
}

var formatterTests = []struct {
	name string
	run  func(*logrus.Logger)
	out  map[string]interface{}
}{
	{
		name: "info with field",
		run: func(logger *logrus.Logger) {
			logger.WithField("foo", "bar").Info("my log entry")
		},
		out: map[string]interface{}{
			"severity": "INFO",
			"message":  "my log entry",
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
			},
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
		},
	},
	{
		name: "error with field",
		run: func(logger *logrus.Logger) {
			logger.WithField("foo", "bar").Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
				"reportLocation": map[string]interface{}{
					"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
					"lineNumber":   64.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
					"functionName": "glob..func2",
				},
			},
			"sourceLocation": map[string]interface{}{
				"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
				"lineNumber":   64.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
				"functionName": "glob..func2",
			},
		},
	},
	{
		name: "error with field and err",
		run: func(logger *logrus.Logger) {
			logger.
				WithField("foo", "bar").
				WithError(errors.New("test error")).
				Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry: test error",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
				"reportLocation": map[string]interface{}{
					"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
					"lineNumber":   96.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
					"functionName": "glob..func3",
				},
			},
			"sourceLocation": map[string]interface{}{
				"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
				"lineNumber":   96.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
				"functionName": "glob..func3",
			},
		},
	},
	{
		name: "error with field and httpRequest",
		run: func(logger *logrus.Logger) {
			logger.
				WithFields(logrus.Fields{
					"foo": "bar",
					"httpRequest": &HTTPRequest{
						RequestMethod: "GET",
					},
				}).
				Error("my log entry")
		},
		out: map[string]interface{}{
			"severity": "ERROR",
			"message":  "my log entry",
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "bar",
				},
				"httpRequest": map[string]interface{}{
					"requestMethod": "GET",
				},
				"reportLocation": map[string]interface{}{
					"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
					"lineNumber":   132.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
					"functionName": "glob..func4",
				},
			},
			"sourceLocation": map[string]interface{}{
				"filePath":     "github.com/bendiknesbo/logrus-stackdriver-formatter/formatter_test.go",
				"lineNumber":   132.0, // NOTE: This is the line-number of where the logging happened, inside the `run`-func.
				"functionName": "glob..func4",
			},
			"httpRequest": map[string]interface{}{
				"requestMethod": "GET",
			},
		},
	},
	{
		name: "info with field and err",
		run: func(logger *logrus.Logger) {
			logger.
				WithField("foo", "bar").
				WithError(errors.New("test error")).
				Info("my log entry")
		},
		out: map[string]interface{}{
			"severity": "INFO",
			"message":  "my log entry",
			"context": map[string]interface{}{
				"data": map[string]interface{}{
					"foo":   "bar",
					"error": "test error",
				},
			},
			"serviceContext": map[string]interface{}{
				"service": "test",
				"version": "0.1",
			},
		},
	},
}
