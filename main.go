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

type cleanupTask func()

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

	//memo := make(map[int]uint64)
	//fmt.Println(fib(6, memo), memo)
	//fmt.Println("fibless:", fibLess(3, memo))
	_, err = store.NewPostgres(store.PostgresConfig{})
	if err != nil {
		fmt.Println("error opening store:", err)
		//os.Exit(1)
	}
	ms := store.NewMap()
	svc, err := service.NewFib(ms)
	if err != nil {
		fmt.Println("error creating service:", err)
		os.Exit(1)
	}
	val, err := svc.Fib(7)
	if err != nil {
		fmt.Println("error running fib:", err)
		os.Exit(1)
	}
	fmt.Println(7, val)
	num, err := svc.FibLess(21)
	if err != nil {
		fmt.Println("error running fibless:", err)
		os.Exit(1)
	}
	fmt.Println("less", 21, num)

	// Create the server to handle the IP verify service.  The API module will
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
	waitForShutdown(ctx, srv, log)
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
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	sig := <-interruptChan
	log.Debugw("Termination signal received", "signal", sig)
	for _, t := range tasks {
		t()
	}

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Infof("Shutting down")
}

func fib(n int, memo map[int]uint64) uint64 {
	v, ok := memo[n]
	if ok {
		fmt.Println("hit:", n, memo[n])
		return v
	}
	if n == 0 {
		fmt.Println("0 case")
		memo[0] = 0
		return 0
	}
	if n == 1 {
		fmt.Println("1 case")
		memo[1] = 1
		return 1
	}
	res := fib(n-1, memo) + fib(n-2, memo)
	memo[n] = res
	return res
}

func fibLess(target uint64, memo map[int]uint64) int {
	if target == 0 {
		return 0
	}
	max := uint64(0)
	n := 0

	for k, v := range memo {
		if v > max && v <= target {
			if v == target {
				return k
			}
			max = v
			n = k
		}
	}
	fmt.Println("intermediate:", n, max)
	for {
		if fib(n+1, memo) >= target {
			return n + 1
		}
		n++
	}
}
