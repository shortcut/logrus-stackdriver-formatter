package stackdriver_test

import (
	"strconv"
	"strings"
	"testing"

	stackdriver "github.com/bendiknesbo/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogError(t *testing.T) {
	logger := logrus.New()
	strBuilder := strings.Builder{}
	logger.Out = &strBuilder
	logger.Formatter = stackdriver.NewFormatter(
		stackdriver.WithService("test-service"),
		stackdriver.WithVersion("v0.1.0"),
	)

	logger.Info("application up and running")

	_, err := strconv.ParseInt("text", 10, 64)
	require.EqualError(t, err, "strconv.ParseInt: parsing \"text\": invalid syntax")
	logger.WithError(err).Error("unable to parse integer")

	res := strBuilder.String()
	require.NotEmpty(t, res)
	require.Contains(t, res, "application up and running")
	require.Contains(t, res, "strconv.ParseInt: parsing \\\"text\\\": invalid syntax") // Note: Double-escaped, as it is stored as JSON.
	require.Contains(t, res, "unable to parse integer")

	// Output:
	// {"timestamp":"2020-10-12T12:26:00Z","message":"application up and running","severity":"INFO","context":{}}
	// {"timestamp":"2020-10-12T12:26:00Z","serviceContext":{"service":"test-service","version":"v0.1.0"},"message":"unable to parse integer: strconv.ParseInt: parsing \"text\": invalid syntax","severity":"ERROR","context":{"reportLocation":{"filePath":"github.com/bendiknesbo/logrus-stackdriver-formatter/example_test.go","lineNumber":26,"functionName":"TestLogError"}}}
}

func TestLogWarning(t *testing.T) {
	logger := logrus.New()
	strBuilder := strings.Builder{}
	logger.Out = &strBuilder
	logger.Formatter = stackdriver.NewFormatter(
		stackdriver.WithService("test-service"),
		stackdriver.WithVersion("v0.1.0"),
	)

	logger.Info("application up and running")

	_, err := strconv.ParseInt("text", 10, 64)
	require.EqualError(t, err, "strconv.ParseInt: parsing \"text\": invalid syntax")
	logger.WithError(err).Warning("unable to parse integer")

	res := strBuilder.String()
	require.NotEmpty(t, res)
	require.Contains(t, res, "application up and running")
	require.Contains(t, res, "strconv.ParseInt: parsing \\\"text\\\": invalid syntax") // Note: Double-escaped, as it is stored as JSON.
	require.Contains(t, res, "unable to parse integer")

	// Output:
	// {"timestamp":"2020-10-12T12:26:00Z","message":"application up and running","severity":"INFO","context":{}}
	// {"timestamp":"2020-10-12T12:26:00Z","serviceContext":{"service":"test-service","version":"v0.1.0"},"message":"unable to parse integer: strconv.ParseInt: parsing \"text\": invalid syntax","severity":"WARNING","context":{"reportLocation":{"filePath":"github.com/bendiknesbo/logrus-stackdriver-formatter/example_test.go","lineNumber":26,"functionName":"TestLogError"}}}
}
