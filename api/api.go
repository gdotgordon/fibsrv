// Package api is the endpoint implementation for the IP verify service.
// The HTTP endpoint implementations are here.  This package deals with
// unmarshaling and marshaling payloads, dispatching to the service (which
// itself contains an instance of the store), processing those errors,
// and implementing proper REST semantics.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gdotgordon/fibsrv/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Definitions for the supported URL endpoints.
const (
	fibURL     = "/v1/fib"     // get fib(n)
	fibLessURL = "/v1/fibless" // get count(memos) < n
	clearURL   = "/v1/clear"   // clear the DB table
)

// StatusResponse is the JSON returned for status notifications.
type StatusResponse struct {
	Status string `json:"status"`
}

// API is the item that dispatches to the endpoint implementations
type apiImpl struct {
	service service.FibService
	log     *zap.SugaredLogger
}

// Init sets up the endpoint processing.  There is nothing returned, other
// than potntial errors, because the endpoint handling is configured in
// the passed-in muxer.
func Init(ctx context.Context, r *mux.Router, service service.FibService, log *zap.SugaredLogger) error {
	ap := apiImpl{service: service, log: log}
	r.HandleFunc(fibURL, ap.fib).Queries("n", "{n:[0-9]+}").Methods(http.MethodGet)
	r.HandleFunc(fibLessURL, ap.fibLess).Queries("target", "{target:[0-9]+}").Methods(http.MethodGet)
	r.HandleFunc(clearURL, ap.clear).Methods(http.MethodGet)

	var wrapContext = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rc := r.WithContext(ctx)
			next.ServeHTTP(w, rc)
		})
	}

	var loggingMiddleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infow("Handling URL", "url", r.URL)
			next.ServeHTTP(w, r)
		})
	}
	r.Use(loggingMiddleware)
	r.Use(wrapContext)
	return nil
}

// Returns fib(n)
func (a apiImpl) fib(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		ioutil.ReadAll(r.Body)
		defer r.Body.Close()
	}
	ntxt := r.URL.Query().Get("n")
	n, err := strconv.Atoi(ntxt)
	if err != nil || n < 0 {
		a.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := a.service.Fib(n)
	if err != nil {
		a.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"result": %d}`, res)))
}

// Count number of memoized items less than target
func (a apiImpl) fibLess(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		ioutil.ReadAll(r.Body)
		defer r.Body.Close()
	}
	ttxt := r.URL.Query().Get("target")
	target, err := strconv.Atoi(ttxt)
	if err != nil || target < 0 {
		fmt.Println("******4")
		a.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	fmt.Println("******1", target)

	resp, err := a.service.FibLess(uint64(target))
	if err != nil {
		a.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Println("******2", resp)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"result": %d}`, resp)))
}

func (a *apiImpl) clear(w http.ResponseWriter, r *http.Request) {
	if err := a.service.Clear(); err != nil {
		a.writeErrorResponse(w, http.StatusInternalServerError, err)
	}
}

// For HTTP bad request responses, serialize a JSON status message with
// the cause.
func (a apiImpl) writeErrorResponse(w http.ResponseWriter, code int, err error) {
	a.log.Errorw("invoke error", "error", err, "code", code)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	b, _ := json.MarshalIndent(StatusResponse{Status: err.Error()}, "", "  ")
	w.Write(b)
}
