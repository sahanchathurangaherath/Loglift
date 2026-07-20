package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go/v3"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

// newOSClient creates an OpenSearch client pointed at the given address.
func newOSClient(addr string) (*opensearchapi.Client, error) {
	cfg := opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{addr},
		},
	}
	return opensearchapi.NewClient(cfg)
}

// indexName returns today's daily index, e.g. logs-2026.07.16
func indexName() string {
	return "logs-" + time.Now().UTC().Format("2006.01.02")
}

// bulkWrite writes a batch of records to OpenSearch using the _bulk API.
// Returns an error if the HTTP call itself fails; individual document errors
// inside a 200 response are checked separately (see hasItemErrors below).
func bulkWrite(ctx context.Context, client *opensearchapi.Client, records []logtypes.LogRecord) error {
	if len(records) == 0 {
		return nil
	}

	var buf bytes.Buffer
	index := indexName()
	for _, rec := range records {
		meta := map[string]interface{}{
			"index": map[string]string{"_index": index},
		}
		metaLine, _ := json.Marshal(meta)
		docLine, err := json.Marshal(rec)
		if err != nil {
			continue // skip unmarshalable records rather than fail the whole batch
		}
		buf.Write(metaLine)
		buf.WriteByte('\n')
		buf.Write(docLine)
		buf.WriteByte('\n')
	}

	resp, err := client.Bulk(ctx, opensearchapi.BulkReq{Body: &buf})
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	if resp.Errors {
		return fmt.Errorf("bulk request completed with %d/%d item errors", countErrors(resp), len(records))
	}
	return nil
}

func countErrors(resp *opensearchapi.BulkResp) int {
	count := 0
	for _, item := range resp.Items {
		for _, action := range item {
			if action.Error != nil {
				count++
			}
		}
	}
	return count
}
