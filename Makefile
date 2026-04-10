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

destroy-etcd:
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

# Deploy targets

deploy: etcd cluster kube

cluster:
	minikube start --nodes 2 --driver=docker -p marafonet-cluster
	minikube addons enable ingress -p marafonet-cluster

kube:
	kubectl get nodes
	kubectl apply -f deployment/kubernetes/
	kubectl get pods

tunnel: 
	kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80

destroy-kube:
	kubectl delete -f deployment/kubernetes/

destroy-cluster: 
	minikube delete -p marafonet-cluster
	

