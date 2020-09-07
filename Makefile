# Commands to build/start and stop the container and run tests.

testall: service_test store_test integration_test bench_test

serverup: build_exec
	docker-compose up --build

serverdown:
	docker-compose down --volumes --rmi all

service_test:
	@echo "running unit tests in service (dockertest image load may take some time) ..."
	 go test ./service -v -count=1

bench_test:
	@echo "running benchmark tests (db image load and execution may take literally a minute or two) ..."
	go test ./... -run=Bench -bench=.

integration_test: build_exec
	@echo "running integration tests (invoking docker-compose first) ..."
	docker-compose up --build -d
	go test ./... -tags=integration -v -count=1
	docker-compose down --volumes --rmi all

store_test:
	@echo "running store unit tests (dockertest image load may take some time) ..."
	go test ./store -tags=store -v -count=1

build_exec:
	@echo "building executable ..."
	rm -f fibsrv
	env GOOS=linux GOARCH=amd64 go build .
