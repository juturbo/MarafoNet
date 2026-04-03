all: react dockerize run

marafonet: dockerize run

dev:
	cd frontend && npm start

react:
	cd frontend && npm run build

straight:
	go run ./cmd/marafonet

run:
	docker run -p 5000:5000 marafonet:latest

dockerize:
	docker build -f deployment/Dockerfile -t marafonet:latest .

etcd:
	docker run -d --name etcd-test -p 2379:2379 -e ALLOW_NONE_AUTHENTICATION=yes -e ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:2379" -e ETCD_ADVERTISE_CLIENT_URLS="http://0.0.0.0:2379" quay.io/coreos/etcd:v3.6.8

destroy:
	docker stop etcd-test && docker rm etcd-test