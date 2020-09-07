// +build integration

// Run as: go test -tags=integration
package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gdotgordon/fibsrv/api"
)

var (
	fibAddr   string
	fibClient *http.Client
)

// This test is similar to the unit tests in service, but invokes everything
// going through the live server via the REST API.
func TestMain(m *testing.M) {
	var err error

	port := 8080
	fa := os.Getenv("FIB_PORT")
	if fa != "" {
		port, err = strconv.Atoi(fa)
		if err != nil {
			fmt.Println("bad port value for FIB_PORT:", fa)
			os.Exit(1)
		}
	}
	fibAddr = fmt.Sprintf("http://localhost:%d/v1/", port)

	fibClient = http.DefaultClient

	// Make sure we can reach the server.
	connected := false
	for i := 0; i < 10; i++ {
		if err := invokeClear(); err != nil {
			time.Sleep(1 * time.Second)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		fmt.Println("cannot connect to server")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestFibInteg(t *testing.T) {
	if err := invokeClear(); err != nil {
		t.Fatal("error clearing database:", err)
	}

	for i, v := range []struct {
		n      int
		result uint64
	}{
		{n: 0, result: 0},
		{n: 1, result: 1},
		{n: 2, result: 1},
		{n: 5, result: 5},
		{n: 10, result: 55},
		{n: 15, result: 610},
		{n: 20, result: 6765},
		{n: 8, result: 21},
	} {
		resp, err := http.Get(fmt.Sprintf("%s%s?n=%d", fibAddr, "fib", v.n))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Bad status code getting fib %dd: %d", v.n, resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		var res api.ResultResponse
		if err = json.Unmarshal(b, &res); err != nil {
			t.Fatal(err)
		}
		if res.Result != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.n, v.result, res)
		}
	}

}

func TestFibLessInteg(t *testing.T) {
	if err := invokeClear(); err != nil {
		t.Fatal("error clearing database:", err)
	}

	for i, v := range []struct {
		target uint64
		result int
	}{
		{target: 0, result: 0},
		{target: 1, result: 1},
		{target: 2, result: 3},
		{target: 11, result: 7},
		{target: 120, result: 12},
		{target: 58, result: 11},
	} {
		resp, err := http.Get(fmt.Sprintf("%s%s?target=%d", fibAddr, "fibless", v.target))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Bad status code getting fibless %dd: %d", v.target, resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		var res api.ResultResponse
		if err = json.Unmarshal(b, &res); err != nil {
			t.Fatal(err)
		}
		if int(res.Result) != v.result {
			t.Fatalf("%d: less(%d), expected %d, got %d", i, v.result, res.Result, res)
		}
	}
}

func invokeClear() error {
	resp, err := http.Get(fibAddr + "clear")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code clearing db: %d", resp.StatusCode)
	}
	return nil
}
