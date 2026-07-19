package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

const (
	batchSize     = 50
	flushInterval = 500 * time.Millisecond
	streamName    = "logs:incoming"
)

// runShipper reads parsed records off the records channel, batches them,
// and XADDs each record to the Redis stream on a size-or-time trigger.
func runShipper(ctx context.Context, rdb *redis.Client, records <-chan logtypes.LogRecord) {
	var batch []logtypes.LogRecord
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, rec := range batch {
			payload, err := json.Marshal(rec)
			if err != nil {
				log.Printf("marshal error, dropping record: %v", err)
				continue
			}
			_, err = rdb.XAdd(ctx, &redis.XAddArgs{
				Stream: streamName,
				Values: map[string]interface{}{"data": payload},
			}).Result()
			if err != nil {
				log.Printf("XADD failed: %v", err)
			}
		}
		log.Printf("flushed %d records to %s", len(batch), streamName)
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case rec, ok := <-records:
			if !ok {
				flush()
				return
			}
			batch = append(batch, rec)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}