all: react dockerize run

marafonet: dockerize run

react:
	cd frontend && npm run build

straight:
	cd cmd/marafonet && go run main.go

run:
	docker run -p 5000:5000 marafonet:latest

dockerize:
	docker build -f deployment/Dockerfile -t marafonet:latest .
