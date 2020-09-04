package main

import (
	"fmt"
	"os"

	"github.com/gdotgordon/fibsrv/service"
	"github.com/gdotgordon/fibsrv/store"
)

func main() {
	//memo := make(map[int]uint64)
	//fmt.Println(fib(6, memo), memo)
	//fmt.Println("fibless:", fibLess(3, memo))
	_, err := store.NewPostgres(store.PostgresConfig{})
	if err != nil {
		fmt.Println("error opening store:", err)
		//os.Exit(1)
	}
	ms := store.NewMap()
	srv, err := service.NewFib(ms)
	if err != nil {
		fmt.Println("error creating service:", err)
		os.Exit(1)
	}
	val, err := srv.Fib(7)
	if err != nil {
		fmt.Println("error running fib:", err)
		os.Exit(1)
	}
	fmt.Println(7, val)
	num, err := srv.FibLess(21)
	if err != nil {
		fmt.Println("error running fibless:", err)
		os.Exit(1)
	}
	fmt.Println("less", 21, num)
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
