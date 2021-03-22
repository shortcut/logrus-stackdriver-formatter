package stackdriver

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTrace(t *testing.T) {
	var out bytes.Buffer

	logger := logrus.New()
	logger.Out = &out
	logger.Formatter = NewFormatter(
		WithService("test"),
		WithVersion("0.1"),
	)

	logger.WithField(KeyTrace, "my-trace").WithField(KeySpanID, "my-span").Info("my log entry")

	var got map[string]interface{}
	json.Unmarshal(out.Bytes(), &got)

	want := map[string]interface{}{
		"severity": "INFO",
		"message":  "my log entry",
		"context":  map[string]interface{}{},
		"serviceContext": map[string]interface{}{
			"service": "test",
			"version": "0.1",
		},
		"logging.googleapis.com/trace":  "my-trace",
		"logging.googleapis.com/spanId": "my-span",
	}

	require.True(t, reflect.DeepEqual(got, want), "unexpected output = %# v; \n want = %# v; \n diff: %# v", pretty.Formatter(got), pretty.Formatter(want), pretty.Diff(got, want))
}

func TestTraceWithProjectID(t *testing.T) {
	var out bytes.Buffer

	logger := logrus.New()
	logger.Out = &out
	logger.Formatter = NewFormatter(
		WithService("test"),
		WithVersion("0.1"),
		WithProjectID("my-project"),
	)

	logger.WithField(KeyTrace, "my-trace").WithField(KeySpanID, "my-span").Info("my log entry")

	var got map[string]interface{}
	json.Unmarshal(out.Bytes(), &got)

	want := map[string]interface{}{
		"severity": "INFO",
		"message":  "my log entry",
		"context":  map[string]interface{}{},
		"serviceContext": map[string]interface{}{
			"service": "test",
			"version": "0.1",
		},
		"trace_id":                      "my-trace",
		"logging.googleapis.com/trace":  "projects/my-project/traces/my-trace",
		"logging.googleapis.com/spanId": "my-span",
	}

	require.True(t, reflect.DeepEqual(got, want), "unexpected output = %# v; \n want = %# v; \n diff: %# v", pretty.Formatter(got), pretty.Formatter(want), pretty.Diff(got, want))

}
