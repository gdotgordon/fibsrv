// +build integration

// Run as: go test -tags=integration
package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/gdotgordon/fibsrv/api"
)

const (
	fibAddr = "http://localhost:8080/v1/"
)

var (
	fibClient *http.Client
)

func TestMain(m *testing.M) {
	fibClient = http.DefaultClient
	os.Exit(m.Run())
}

func TestFib(t *testing.T) {
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
		fmt.Println("fib", v.n)
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?n=%d", fibAddr, "fib", v.n), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := fibClient.Do(req)
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

func TestFibLessDB(t *testing.T) {
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
		fmt.Println("fibless", v.target)
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?target=%d", fibAddr, "fibless", v.target), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := fibClient.Do(req)
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
	req, err := http.NewRequest(http.MethodGet, fibAddr+"clear", nil)
	if err != nil {
		return err
	}

	resp, err := fibClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code clearing db: %d", resp.StatusCode)
	}
	return nil
}
