// Package main is the starting point for the HTTP server for the
// fibonacci service  It creates the api, store and service artifacts
// and then launches the server.  It supports signal handlers for a
// clean shutdown.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gdotgordon/fibsrv/api"
	"github.com/gdotgordon/fibsrv/service"
	"github.com/gdotgordon/fibsrv/store"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	// PostgresPort is the default postgres port
	PostgresPort = 5432

	// PostgresHost is the host from docker-compose
	PostgresHost = "db"
)

type cleanupTask func() error

var (
	portNum  int    // listen port
	logLevel string // zap log level
	timeout  int    // server timeout in seconds
)

func init() {
	flag.IntVar(&portNum, "port", 8080, "HTTP port number")
	flag.StringVar(&logLevel, "log", "production",
		"log level: 'production', 'development'")
	flag.IntVar(&timeout, "timeout", 30, "server timeout (seconds)")
}

func main() {
	flag.Parse()

	// We'll propagate the context with cancel thorughout the program,
	// to be used by various entities, such as http clients, server
	// methods we implement, and other loops using channels.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up logging.
	log, err := initLogging()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v", err)
		os.Exit(1)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresUser == "" || postgresPassword == "" || postgresDB == "" {
		fmt.Fprintf(os.Stderr,
			"environment vars POSTGRES_USER, POSTGRES_PASSWORD and POSTGRES_DB must be set")
		os.Exit(1)
	}
	dataStore, err := store.NewPostgres(ctx,
		store.PostgresConfig{
			Host:     PostgresHost,
			Port:     PostgresPort,
			User:     postgresUser,
			Password: postgresPassword,
			DBName:   postgresDB,
		},
		log,
	)
	if err != nil {
		fmt.Println("error opening store:", err)
		os.Exit(1)
	}

	svc, err := service.NewFib(dataStore)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating service:", err)
		os.Exit(1)
	}

	// Create the server to handle the Fibonacci service.  The API module will
	// set up the routes, as we don't need to know the details in the
	// main program.
	muxer := mux.NewRouter()

	// Initialize the API layer.
	if err := api.Init(ctx, muxer, svc, log); err != nil {
		log.Errorf("Error initializing API layer", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Handler:      muxer,
		Addr:         fmt.Sprintf(":%d", portNum),
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	// Start server
	go func() {
		log.Infow("Listening for connections", "port", portNum)
		if err := srv.ListenAndServe(); err != nil {
			log.Infow("Server completed", "err", err)
		}
	}()

	// Block until we shutdown.
	srvShut := dataStore.(*store.PostgresStore).Shutdown
	waitForShutdown(ctx, srv, log, srvShut)
}

// Set up the logger, condsidering any env vars.
func initLogging() (*zap.SugaredLogger, error) {
	var lg *zap.Logger
	var err error

	pdl := strings.ToLower(os.Getenv("IPVERIFY_LOG_LEVEL"))
	if strings.HasPrefix(pdl, "d") {
		logLevel = "development"
	}

	var cfg zap.Config
	if logLevel == "development" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.DisableStacktrace = true
	lg, err = cfg.Build()
	if err != nil {
		return nil, err
	}
	return lg.Sugar(), nil
}

// Setup for clean shutdown with signal handlers/cancel.
func waitForShutdown(ctx context.Context, srv *http.Server,
	log *zap.SugaredLogger, tasks ...cleanupTask) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	sig := <-interruptChan
	log.Debugw("Termination signal received", "signal", sig)
	for _, t := range tasks {
		if err := t(); err != nil {
			log.Infof("Shutdown error", "error", err.Error())
		}
	}

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Infof("Server shutting down")
}
