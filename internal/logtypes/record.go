package logtypes

type LogRecord struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"` // INFO | WARN | ERROR
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	TraceID   string            `json:"trace_id,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}