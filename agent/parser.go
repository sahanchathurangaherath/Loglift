package main

import (
	"encoding/json"
	"time"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

// parseLine turns a raw log line into a LogRecord. If the line isn't valid JSON,
// it's wrapped as an INFO record instead of being dropped.
func parseLine(raw string, service string) logtypes.LogRecord {
	var rec logtypes.LogRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return logtypes.LogRecord{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     "INFO",
			Service:   service,
			Message:   raw,
		}
	}

	// backfill fields the source service might not have set
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if rec.Service == "" {
		rec.Service = service
	}
	return rec
}