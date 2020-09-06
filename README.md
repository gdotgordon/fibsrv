# fibsrv
Fibonacci number server

## Introduction and Overview
This code implements the Fibonacci service including the two required functions:
1. get the value of Fib(n) using memoization and a DB backing store
2. compute the number of *intermediate* terms for any given target value
3. clear all the rows of the database

## Implementation
The solution is implemented as a Rest-like server written in Go, which deploys completely inside Docker containers using `docker-compose`, including pulling a Postgres image from docker hub.  As the server exposes port 8080 to the native host, one could use a tool such as Postman or the `curl` command to exercise the API on http://localhost:8080.  The commands mentioned in the introduction are all supported, and the use of each will be documented later.

### Test Cases
The are three types of test cases: unit tests, benchmark tests and full server integration tests.  Some of the unit tests use *ory/dockertest* to populate a mock Postgres repository, as recommended in the assignment.  This can add a bit of startup time to the tests, due to network latency issues, etc.  They also seem to clean up well enough as long as you don't stop the tests artificially.  If port 8080 ends up being in use, then a tool like the Docker Dashboard can be used to remove the existing image.

### Running Everything From the Makefile
There is a Makefile with targets to bring the app up and down, as well a set of tests or all the tests.  Each `make` target is invoked as `make <target`>, e.g. `make serverup`, `make bench`.  The list of make targets is:

* `serverup` - launches both the server program and Postgres containers.  This runs the output to a window (not in detached mode, so it is best to leave this running in its own window.

* `serverdown` - takes down the containers and removes the Docker images.  If you start it again it will have to pull the images, but this is intentional, as someone reviewing the code isn't likely to run this over and over.

* `unit` - runs the unit tests.  This does pull in docker images to mock the database.  There are some unit tests which do use a mock hash map-based store, but essentially the same tests are also done using the Postgres image pulled by `dockertest`.

* `bench` - runs a set of benchmark tests.  The implications of the results of these tests will be discussed later on, but they are basically variations on how much the memo cache is depended on, and how that affects the results.

* `integration` - stands up the full server on docker, and connects to it using an HTTP client.  Again, not much new logic, but it is the full end-to-end test.

* testall - runs all three test types listed above

### Invoking endpoints
The three endpoints may be invoked as follows:
* Fib(n): HTTP GET endpoint with query parameter `n`, e.g.: http://localhost:8080/v1/fib?n=15

* FibLess(target): returns the number of intermediate memoized terms HTTP GET, query parameter `target`, e.g. http://localhost:8080/v1/fibless?target=120

* Clear database HTTP GET http://localhost:8080/v1/clear

The first two return a simple JSON object with the result number.  All return HTTP 200 on success, and an appropriate code on error.

## Code and architecture
There is a main function which basically launches the HTTP server and invoke the api layer.  The set of packages is:
* `api` - the HTTP handlers.  The handlers takes the requests and invoke the service layer.
* `service` - implements the Fib service "business logic".  For example, it runs the recursive fibonacci algorithm, stores results to the data layer, as well as computes the number of intermediate memos stored.
* `store` - there is a `Store` interface defined which satisfies the backend requirements of the service layer.  These are simple queries, such as storing a memoized value, trying to fetch a memoized value if it exists, and counting the number memos whose fibonacci value is less than a specified target.  There is a Postgres based store, as well as a hash-map based store (which was used to write and debug the service layer).

The source code is well-commented, and specific details may be found in the code.

Currently all of the tests are in the service layer, as this is the best place to execute all the logic.  Given more time, tests are needed in the storage lyaer.

### Performance
The memoization helps for sure and keeps resource usage down.  Accessing the db add a lot of overhead as well.

To measure performance, I wrote a benchmark test with three variants:
1. Compute a series of Fiboncci numbers utilizing the DB cache
2. Same as above, but clear the cache after each computation.
3. Same as number 2, but use a mock no-op database that never yields a cache hit.  The point was to try and eliminate the database overhead of clearing the cache and see how much that was contributing to the slowness.

Some observations for the above scenario:

```
BenchmarkFibonacciNoClearCache-12    	      40	  29999606 ns/op	   16084 B/op	     438 allocs/op
BenchmarkFibonacciClearCache-12      	       3	 372842304 ns/op	  139816 B/op	    3963 allocs/op
BenchmarkFibonacciNoCache-12         	   18772	     66555 ns/op	       0 B/op	       0 allocs/op
```
Not surprisingly, the speed and memory usage are significantly improved using the cache.  Using the mock cache blows the other two out of the water, showing us that using the database for memo storage (also not surprisingly) dwarfs the raw compute time.  That said, the use of the database no doubt helps alleviate the use of the stack for recursion, especially as more and more values get added to the cache.
