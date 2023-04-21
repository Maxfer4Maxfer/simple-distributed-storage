package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simple-storage/internal/endpoint/chunkmanager"
	entrypoint "simple-storage/internal/entrypoint/http"
	handler "simple-storage/internal/entrypoint/http/storageserver"
	"simple-storage/internal/storageserver"
	"syscall"
)

func main() {
	var (
		chunkManagerAddress = flag.String("chunk-manager", "0.0.0.0:9000",
			"TCP/IP address of chunk-manager")
		address = flag.String(
			"address", "0.0.0.0:9001", "TCP/IP address of storage-server")
		dataDirectory = flag.String(
			"data-directory", "data", "directory with chunks")
		timeBetweetRegistrationRetrySecond = flag.Int(
			"registration-retry-timeout", 4, "how long should wait between unsuccesfull registraton")
	)

	flag.Parse()

	log := log.New(os.Stdout, "ss", log.Lshortfile|log.Lmicroseconds)

	chunkManagerClient := chunkmanager.New(
		log, *chunkManagerAddress, &http.Client{})

	storageServer := storageserver.New(log, storageserver.Config{
		Address:                            *address,
		DataDirectory:                      *dataDirectory,
		TimeBetweetRegistrationRetrySecond: *timeBetweetRegistrationRetrySecond,
	}, chunkManagerClient)

	server := entrypoint.New(
		log,
		entrypoint.Config{
			Address: *address,
		},
		handler.New(log, storageServer),
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
