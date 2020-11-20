package stackdriver

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/shortcut/logrus-stackdriver-formatter/internal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStackSkip(t *testing.T) {
	var out bytes.Buffer

	logger := logrus.New()
	logger.Out = &out
	logger.Formatter = NewFormatter(
		WithService("test"),
		WithVersion("0.1"),
		WithStackSkip("github.com/shortcut/logrus-stackdriver-formatter/internal"),
	)

	mylog := internal.LogWrapper{
		Logger: logger,
	}

	mylog.Error("my log entry")

	var got map[string]interface{}
	json.Unmarshal(out.Bytes(), &got)

	want := map[string]interface{}{
		"severity": "ERROR",
		"message":  "my log entry",
		"serviceContext": map[string]interface{}{
			"service": "test",
			"version": "0.1",
		},
		"context": map[string]interface{}{
			"reportLocation": map[string]interface{}{
				"filePath":     "github.com/shortcut/logrus-stackdriver-formatter/stackskip_test.go",
				"lineNumber":   30.0,
				"functionName": "TestStackSkip",
			},
		},
		"sourceLocation": map[string]interface{}{
			"filePath":     "github.com/shortcut/logrus-stackdriver-formatter/stackskip_test.go",
			"lineNumber":   30.0,
			"functionName": "TestStackSkip",
		},
	}

	require.True(t, reflect.DeepEqual(got, want), "unexpected output = %# v; \n want = %# v; \n diff: %# v", pretty.Formatter(got), pretty.Formatter(want), pretty.Diff(got, want))
}
