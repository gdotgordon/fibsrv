# Commands to build/start and stop the container.

composeup:
	docker-compose up --build

composedown:
	docker-compose down --volumes --rmi all

composedown_norm:
	docker-compose down --volumes
