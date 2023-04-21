package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simple-storage/internal/apiserver"
	"simple-storage/internal/chunkmanager"
	storageServerClient "simple-storage/internal/endpoint/storageserver"
	entrypoint "simple-storage/internal/entrypoint/http"
	handler "simple-storage/internal/entrypoint/http/apiserver"
	"syscall"
)

func main() {
	var (
		address = flag.String(
			"address", "0.0.0.0:9000", "TCP/IP address of storage-server")
		maxChunkSizeBytes     = flag.Int("max-chunk-size-bytes", 10240, "chunk size")
		erasureCodingFraction = flag.Int(
			"erasure-coding-fraction", 5, "erasure coding fraction")
	)

	flag.Parse()

	log := log.New(os.Stdout, "api", log.Lshortfile|log.Lmicroseconds)

	chunkManager := chunkmanager.New(log, chunkmanager.Config{
		MaxChunkSizeBytes:     *maxChunkSizeBytes,
		ErasureCodingFraction: *erasureCodingFraction,
	})

	apiServer := apiserver.New(
		log,
		apiserver.Config{},
		chunkManager,
		func(address string) apiserver.StorageServer {
			return storageServerClient.New(log, address, &http.Client{})
		},
	)

	server := entrypoint.New(
		log,
		entrypoint.Config{
			Address: *address,
		},
		handler.New(log, apiServer, chunkManager),
	)

	errServer := server.Start()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errServer:
		log.Printf("problem with TCP Server %s", err)
	case <-osSignals:
		log.Print("shutdown the server")

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("ERROR: failure to shutdown TCP Server: %s", err)
		}
	}

}
