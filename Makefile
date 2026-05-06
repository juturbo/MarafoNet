all: react dockerize run

# Development targets

dev:
	npm --prefix frontend start

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
	npm --prefix frontend run build

build: test react
	docker build -f deployment/marafonet/Dockerfile -t marafonet:latest .
	docker build -f deployment/matchmaking/Dockerfile -t matchmaking:latest .

push:
	docker tag marafonet:latest ghcr.io/juturbo/marafonet:latest
	docker push ghcr.io/juturbo/marafonet:latest
	docker tag matchmaking:latest ghcr.io/juturbo/mf-matchmaking:latest
	docker push ghcr.io/juturbo/mf-matchmaking:latest

# Test targets

test:
	@set -e; \
	$(MAKE) destroy-etcd >/dev/null 2>&1 || true; \
	trap '$(MAKE) destroy-etcd >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) etcd; \
	$(MAKE) test-commands

test-commands:
	go test -v ./...

# Deploy targets

deploy: test etcd cluster certs secrets kube tunnel

cluster:
	minikube start --nodes 2 --driver=docker -p marafonet-cluster
	minikube addons enable ingress -p marafonet-cluster
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s

kube:
	kubectl get nodes
	kubectl apply -f deployment/kubernetes/
	kubectl get pods

tunnel: 
	kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:443

# Run once: TLS certificate generation and K8s secret creation
certs: 
	openssl req -x509 -newkey rsa:4096 -sha256 -nodes -keyout ./deployment/kubernetes/certs/tls.key -out ./deployment/kubernetes/certs/tls.crt -subj "/CN=localhost" -days 365

secrets:	
	kubectl create secret tls marafonet-tls --key=./deployment/kubernetes/certs/tls.key --cert=./deployment/kubernetes/certs/tls.crt

# Add minikube's IP to /etc/hosts
hosts:
	echo "$(minikube ip -p marafonet-cluster) marafo.net" | sudo tee -a /etc/hosts

# Cleanup targets
destroy-kube:
	kubectl delete -f deployment/kubernetes/
	kubectl delete secret marafonet-tls

destroy-cluster: 
	minikube delete -p marafonet-cluster

cleanup: destroy-kube destroy-cluster
	

