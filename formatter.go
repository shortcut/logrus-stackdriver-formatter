package stackdriver

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-stack/stack"
	"github.com/sirupsen/logrus"
)

var skipTimestamp bool

type severity string

const (
	severityDebug    severity = "DEBUG"
	severityInfo     severity = "INFO"
	severityWarning  severity = "WARNING"
	severityError    severity = "ERROR"
	severityCritical severity = "CRITICAL"
	severityAlert    severity = "ALERT"
)

var levelsToSeverity = map[logrus.Level]severity{
	logrus.DebugLevel: severityDebug,
	logrus.InfoLevel:  severityInfo,
	logrus.WarnLevel:  severityWarning,
	logrus.ErrorLevel: severityError,
	logrus.FatalLevel: severityCritical,
	logrus.PanicLevel: severityAlert,
}

// Known keys
const (
	KeyTrace       = "trace"
	KeySpanID      = "spanID"
	KeyHTTPRequest = "httpRequest"
	KeyLogID       = "logID"
)

// ServiceContext provides the data about the service we are sending to Google.
type ServiceContext struct {
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
}

// Entry stores a log entry. More information here: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
type Entry struct {
	LogName        string          `json:"logName,omitempty"`
	Timestamp      string          `json:"timestamp,omitempty"`
	HTTPRequest    *HTTPRequest    `json:"httpRequest,omitempty"`
	Trace          string          `json:"trace,omitempty"`
	SpanID         string          `json:"spanId,omitempty"`
	ServiceContext *ServiceContext `json:"serviceContext,omitempty"`
	Message        string          `json:"message,omitempty"`
	Severity       severity        `json:"severity,omitempty"`
	Context        *Context        `json:"context,omitempty"`
	SourceLocation *ReportLocation `json:"sourceLocation,omitempty"`
}

// ReportLocation is the information about where an error occurred.
type ReportLocation struct {
	FilePath     string `json:"filePath,omitempty"`
	LineNumber   int    `json:"lineNumber,omitempty"`
	FunctionName string `json:"functionName,omitempty"`
}

// Context is sent with every message to stackdriver.
type Context struct {
	Data           map[string]interface{} `json:"data,omitempty"`
	ReportLocation *ReportLocation        `json:"reportLocation,omitempty"`
	HTTPRequest    *HTTPRequest           `json:"httpRequest,omitempty"`
}

// HTTPRequest defines details of a request and response to append to a log.
type HTTPRequest struct {
	RequestMethod string `json:"requestMethod,omitempty"`
	RequestURL    string `json:"requestUrl,omitempty"`
	RequestSize   string `json:"requestSize,omitempty"`
	Status        string `json:"status,omitempty"`
	ResponseSize  string `json:"responseSize,omitempty"`
	UserAgent     string `json:"userAgent,omitempty"`
	RemoteIP      string `json:"remoteIp,omitempty"`
	ServerIP      string `json:"serverIp,omitempty"`
	Referer       string `json:"referer,omitempty"`
	Latency       string `json:"latency,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

// Formatter implements Stackdriver formatting for logrus.
type Formatter struct {
	Service   string
	Version   string
	ProjectID string
	StackSkip []string
}

// Option lets you configure the Formatter.
type Option func(*Formatter)

// WithService lets you configure the service name used for error reporting.
func WithService(n string) Option {
	return func(f *Formatter) {
		f.Service = n
	}
}

// WithVersion lets you configure the service version used for error reporting.
func WithVersion(v string) Option {
	return func(f *Formatter) {
		f.Version = v
	}
}

// WithProjectID makes sure all entries have your Project information.
func WithProjectID(i string) Option {
	return func(f *Formatter) {
		f.ProjectID = i
	}
}

// WithStackSkip lets you configure which packages should be skipped for locating the error.
func WithStackSkip(v string) Option {
	return func(f *Formatter) {
		f.StackSkip = append(f.StackSkip, v)
	}
}

// NewFormatter returns a new Formatter.
func NewFormatter(options ...Option) *Formatter {
	fmtr := Formatter{
		StackSkip: []string{
			"github.com/sirupsen/logrus",
		},
	}
	for _, option := range options {
		option(&fmtr)
	}
	return &fmtr
}

func (f *Formatter) errorOrigin() (stack.Call, error) {
	skip := func(pkg string) bool {
		for _, skip := range f.StackSkip {
			if pkg == skip {
				return true
			}
		}
		return false
	}

	// We start at 3 to skip this call, our caller's call, and our caller's caller's call.
	for i := 3; ; i++ {
		c := stack.Caller(i)
		// ErrNoFunc indicates we're over traversing the stack.
		if _, err := c.MarshalText(); err != nil {
			return stack.Call{}, nil
		}
		pkg := fmt.Sprintf("%+k", c)
		// Remove vendoring from package path.
		parts := strings.SplitN(pkg, "/vendor/", 2)
		pkg = parts[len(parts)-1]
		if !skip(pkg) {
			return c, nil
		}
	}
}

// taken from https://github.com/sirupsen/logrus/blob/0fb945b034620199c178b1b7067672a9f8f69c3a/json_formatter.go#L61
func replaceErrors(source logrus.Fields) logrus.Fields {
	data := make(logrus.Fields, len(source))
	for k, v := range source {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	return data
}

// ToEntry formats a logrus entry to a stackdriver entry.
func (f *Formatter) ToEntry(e *logrus.Entry) Entry {
	severity := levelsToSeverity[e.Level]

	ee := Entry{
		Message:  e.Message,
		Severity: severity,
		Context: &Context{
			Data: replaceErrors(e.Data),
		},
		ServiceContext: &ServiceContext{
			Service: f.Service,
			Version: f.Version,
		},
	}

	if val, ok := e.Data[KeyTrace]; ok {
		if str, ok := val.(string); ok {
			ee.Trace = str
			delete(ee.Context.Data, KeyTrace)
		}
	}

	if val, ok := e.Data[KeySpanID]; ok {
		if str, ok := val.(string); ok {
			ee.SpanID = str
			delete(ee.Context.Data, KeySpanID)
		}
	}

	if val, ok := e.Data[KeyHTTPRequest]; ok {
		if req, ok := val.(*HTTPRequest); ok {
			ee.HTTPRequest = req
			ee.Context.HTTPRequest = req
			delete(ee.Context.Data, KeyHTTPRequest)
		}
	}

	if val, ok := e.Data[KeyLogID]; ok && f.ProjectID != "" {
		if str, ok := val.(string); ok {
			ee.LogName = fmt.Sprintf("projects/%s/logs/%s", f.ProjectID, url.QueryEscape(str))
			delete(ee.Context.Data, KeyLogID)
		}
	}

	if !skipTimestamp {
		ee.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	switch severity {
	case severityError, severityCritical, severityAlert:
		// https://cloud.google.com/error-reporting/docs/formatting-error-messages
		// When using WithError(), the error is sent separately, but Error
		// Reporting expects it to be a part of the message so we append it
		// instead.
		if err, ok := ee.Context.Data[logrus.ErrorKey]; ok {
			ee.Message = fmt.Sprintf("%s: %s", e.Message, err)
			delete(ee.Context.Data, logrus.ErrorKey)
		} else {
			ee.Message = e.Message
		}

		// Extract report location from call stack.
		if c, err := f.errorOrigin(); err == nil {
			lineNumber, _ := strconv.ParseInt(fmt.Sprintf("%d", c), 10, 64)
			location := &ReportLocation{
				FilePath:     fmt.Sprintf("%+s", c),
				LineNumber:   int(lineNumber),
				FunctionName: fmt.Sprintf("%n", c),
			}
			ee.Context.ReportLocation = location
			ee.SourceLocation = location
		}
	}

	return ee
}

// Format formats a logrus entry according to the Stackdriver specifications.
func (f *Formatter) Format(e *logrus.Entry) ([]byte, error) {
	ee := f.ToEntry(e)

	b, err := json.Marshal(ee)
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}
