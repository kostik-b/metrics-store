// Copyright Konstantin Bakanov 2023

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/kostik-b/metrics-store/pkg/datastore"
	mhandler "github.com/kostik-b/metrics-store/pkg/handler"
)

const (
	readWriteTimeout          = 10
	shutdownTimeout           = 10
	defaultListenPort         = 4000
	defaultMaxRequestBodySize = 1048576
)

func main() {
	// parse flags
	// if something is wrong, print usage
	var listenPortAsInt int = defaultListenPort
	flag.IntVar(&listenPortAsInt, "listen-port", defaultListenPort, "A port to listen on from 1 to 65535")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "Set to true to enable debug output")

	var maxRequestBodySize int64 = defaultMaxRequestBodySize
	flag.Int64Var(&maxRequestBodySize, "max-request-body-size", maxRequestBodySize, "Maximum size of request body")

	var allowUnknownFields bool
	flag.BoolVar(&allowUnknownFields, "allow-unkwnown-fields", false, "Set to true to allow unknown fields")

	flag.Parse()

	if listenPortAsInt < 1 || listenPortAsInt > 65535 {
		log.Printf("ERROR: port specified is out of range: %d\n", listenPortAsInt)
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Using the listen port %d\n", listenPortAsInt)
	listenPortAsString := strconv.Itoa(listenPortAsInt)

	// create server
	metricsDatastore := datastore.GetInstance()

	// create handler
	metricsHandler := mhandler.NewMetricsHandler(metricsDatastore, debug, allowUnknownFields, maxRequestBodySize)

	// create request multiplexer and register handler with it
	serveMux := http.NewServeMux()
	serveMux.Handle("/metrics", metricsHandler)

	metricsServer := &http.Server{
		Addr:         ":" + listenPortAsString,
		Handler:      serveMux,
		ReadTimeout:  readWriteTimeout * time.Second,
		WriteTimeout: readWriteTimeout * time.Second,
	}

	// from https://pkg.go.dev/net/http#Server.Shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		// listen for interrupt signal and perform a graceful shutdown
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		log.Printf("Starting shutdown, waiting up to %d secs\n", shutdownTimeout)

		timeoutContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		if err := metricsServer.Shutdown(timeoutContext); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	// run the server
	if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
		os.Exit(1)
	}

	// wait until all open connections are finished (or timeout expires)
	<-idleConnsClosed
}
