package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	osClient, err := newOSClient("http://localhost:9200")
	if err != nil {
		log.Fatalf("cannot create opensearch client: %v", err)
	}
	log.Println("connected to opensearch")

	runConsumer(ctx, rdb, osClient)
	log.Println("indexer stopped")
}
