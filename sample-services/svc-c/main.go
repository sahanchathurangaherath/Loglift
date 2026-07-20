package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

func main() {
	messages := []string{
		"request handled successfully",
		"cache miss, fetching from db",
		"user authenticated",
	}
	errorMessages := []string{
		"database connection timeout",
		"failed to acquire lock",
		"upstream service returned 500",
	}

	for {
		rec := logtypes.LogRecord{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "svc-c",
			Level:     "INFO",
			Message:   messages[rand.Intn(len(messages))],
		}

		// ~10% of logs are errors, so you have something for alerting later
		if rand.Intn(10) == 0 {
			rec.Level = "ERROR"
			rec.Message = errorMessages[rand.Intn(len(errorMessages))]
		}

		out, _ := json.Marshal(rec)
		fmt.Println(string(out))

		time.Sleep(500 * time.Millisecond)
	}
}