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

func TestProject(t *testing.T) {
	var out bytes.Buffer

	logger := logrus.New()
	logger.Out = &out
	logger.Formatter = NewFormatter(
		WithService("test"),
		WithVersion("0.1"),
		WithProjectID("my-project-id"),
	)

	logger.WithField(KeyLogID, "my-id").Info("my log entry")

	var got map[string]interface{}
	json.Unmarshal(out.Bytes(), &got)

	want := map[string]interface{}{
		"logName":  "my-id",
		"severity": "INFO",
		"message":  "my log entry",
		"context":  map[string]interface{}{},
		"serviceContext": map[string]interface{}{
			"service": "test",
			"version": "0.1",
		},
	}

	require.True(t, reflect.DeepEqual(got, want), "unexpected output = %# v; \n want = %# v; \n diff: %# v", pretty.Formatter(got), pretty.Formatter(want), pretty.Diff(got, want))

}
