# Commands to build/start and stop the container and run tests.

testall: unit integration bench

serverup:
	docker-compose up --build

serverdown:
	docker-compose down --volumes --rmi all

unit:
	@echo "running unit tests (dockertest image load may take some time) ..."
	 go test ./... -v -count=1

bench:
	@echo "running benchmark tests (db image load and execution may take literally a minute or two) ..."
	go test ./... -run=Bench -bench=.

integration:
	@echo "running integration tests (invoking docker-compose first) ..."
	docker-compose up --build -d
	go test ./... -tags=integration -v -count=1
	docker-compose down --volumes --rmi all
