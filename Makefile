all: react dockerize run

# Development targets

dev:
	cd frontend && npm start

straight:
	go run ./cmd/marafonet

matchmaking:
	go run ./cmd/matchmaking

etcd:
	docker run -d --name etcd-test -p 2379:2379 -e ALLOW_NONE_AUTHENTICATION=yes -e ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:2379" -e ETCD_ADVERTISE_CLIENT_URLS="http://0.0.0.0:2379" quay.io/coreos/etcd:v3.6.8

destroy:
	docker stop etcd-test && docker rm etcd-test

# Build targets

react:
	cd frontend && npm run build

build: react
	docker build -f deployment/marafonet/Dockerfile -t marafonet:latest .
	docker build -f deployment/matchmaking/Dockerfile -t matchmaking:latest .

push:
	docker tag marafonet:latest bagarozzi/marafonet:latest
	docker push bagarozzi/marafonet:latest
	docker tag matchmaking:latest bagarozzi/marafonet-matchmaking:latest
	docker push bagarozzi/marafonet-matchmaking:latest

run:
	docker run -p 5000:5000 marafonet:latest
	docker run matchmaking:latest
