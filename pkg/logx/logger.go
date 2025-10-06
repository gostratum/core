package logx

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Fields represents contextual key/value data attached to a log entry.
type Fields map[string]any

// Logger is a minimal stdout logger with optional JSON output.
type Logger struct {
	base *log.Logger
	json bool
	mu   sync.Mutex
}

// New constructs a Logger writing to stdout. When jsonLogs is true, log output
// is emitted as JSON lines.
func New(jsonLogs bool) *Logger {
	flags := 0
	if !jsonLogs {
		flags = log.LstdFlags
	}
	return &Logger{
		base: log.New(os.Stdout, "", flags),
		json: jsonLogs,
	}
}

// Info logs a message at the info level with optional fields.
func (l *Logger) Info(msg string, fields Fields) {
	l.log("info", msg, fields)
}

// Warn logs a message at the warn level with optional fields.
func (l *Logger) Warn(msg string, fields Fields) {
	l.log("warn", msg, fields)
}

// Error logs an error message alongside an error value and optional fields.
func (l *Logger) Error(msg string, err error, fields Fields) {
	if err != nil {
		if fields == nil {
			fields = Fields{"error": err.Error()}
		} else {
			// Avoid mutating caller-provided map.
			cloned := make(Fields, len(fields)+1)
			for k, v := range fields {
				cloned[k] = v
			}
			cloned["error"] = err.Error()
			fields = cloned
		}
	}
	l.log("error", msg, fields)
}

func (l *Logger) log(level, msg string, fields Fields) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.json {
		payload := make(map[string]any, len(fields)+3)
		payload["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
		payload["level"] = level
		payload["msg"] = msg
		for k, v := range fields {
			payload[k] = v
		}
		data, err := json.Marshal(payload)
		if err != nil {
			l.base.Printf("{\"level\":\"error\",\"msg\":\"logger marshal failure\",\"marshal_error\":%q}", err)
			return
		}
		l.base.Print(string(data))
		return
	}

	builder := strings.Builder{}
	builder.WriteString(strings.ToUpper(level))
	builder.WriteByte(' ')
	builder.WriteString(msg)
	for k, v := range fields {
		builder.WriteByte(' ')
		builder.WriteString(fmt.Sprintf("%s=%v", k, v))
	}
	l.base.Println(builder.String())
}
