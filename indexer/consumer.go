package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
	"github.com/redis/go-redis/v9"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

const (
	streamName     = "logs:incoming"
	deadLetterName = "logs:deadletter"
	groupName      = "indexers"
	consumerName   = "indexer-1"
	readCount      = 100
	blockTime      = 2 * time.Second
	maxRetries     = 3
)

// runConsumer reads batches from the Redis stream, indexes them into
// OpenSearch, and acks each entry only after a successful write.
func runConsumer(ctx context.Context, rdb *redis.Client, osClient *opensearchapi.Client) {
	for {
		select {
		case <-ctx.Done():
			log.Println("consumer shutting down")
			return
		default:
		}

		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    readCount,
			Block:    blockTime,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue // no new entries within the block window, loop again
			}
			log.Printf("XREADGROUP error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			processMessages(ctx, rdb, osClient, stream.Messages)
		}
	}
}

func processMessages(ctx context.Context, rdb *redis.Client, osClient *opensearchapi.Client, messages []redis.XMessage) {
	var records []logtypes.LogRecord
	var ids []string

	for _, msg := range messages {
		raw, ok := msg.Values["data"].(string)
		if !ok {
			log.Printf("skipping malformed entry %s: no data field", msg.ID)
			ids = append(ids, msg.ID) // ack it anyway, it'll never be valid
			continue
		}
		var rec logtypes.LogRecord
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			log.Printf("skipping unparsable entry %s: %v", msg.ID, err)
			ids = append(ids, msg.ID)
			continue
		}
		records = append(records, rec)
		ids = append(ids, msg.ID)
	}

	err := retryBulkWrite(ctx, osClient, records)
	if err != nil {
		log.Printf("bulk write failed after retries, sending %d records to dead letter: %v", len(records), err)
		sendToDeadLetter(ctx, rdb, messages)
	} else {
		log.Printf("indexed %d records", len(records))
	}

	// ack everything in this batch either way — successful writes are done,
	// and permanently-failed ones have been preserved in the dead letter stream
	if len(ids) > 0 {
		if err := rdb.XAck(ctx, streamName, groupName, ids...).Err(); err != nil {
			log.Printf("XACK failed: %v", err)
		}
	}
}

func retryBulkWrite(ctx context.Context, osClient *opensearchapi.Client, records []logtypes.LogRecord) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := bulkWrite(ctx, osClient, records); err != nil {
			lastErr = err
			log.Printf("bulk write attempt %d/%d failed: %v", attempt, maxRetries, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond) // simple backoff
			continue
		}
		return nil
	}
	return lastErr
}

func sendToDeadLetter(ctx context.Context, rdb *redis.Client, messages []redis.XMessage) {
	for _, msg := range messages {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: deadLetterName,
			Values: msg.Values,
		})
	}
}
