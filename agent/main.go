package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/sahanchathurangaherath/loglift/internal/logtypes"
)

// sources maps service name -> log file path.
// Hardcoded for now; will move this to env vars / a config file.
var sources = map[string]string{
	"svc-a": "/tmp/logs/svc-a.log",
	"svc-b": "/tmp/logs/svc-b.log",
	"svc-c": "/tmp/logs/svc-c.log",
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown on Ctrl+C or SIGTERM (matters once this runs in Docker/K8s)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		cancel()
	}()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("cannot connect to redis: %v", err)
	}
	log.Println("connected to redis")

	rawLines := make(chan rawLine, 100)
	records := make(chan logtypes.LogRecord, 100)

	for service, path := range sources {
		go tailFile(service, path, rawLines)
		log.Printf("tailing %s (%s)", path, service)
	}

	go func() {
		for rl := range rawLines {
			records <- parseLine(rl.text, rl.service)
		}
	}()

	runShipper(ctx, rdb, records)
	log.Println("agent stopped")
}
