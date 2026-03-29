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
