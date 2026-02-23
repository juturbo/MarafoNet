all: react dockerize run

marafonet: dockerize run

react:
	cd frontend && npm run build

run:
	docker run -p 5000:5000 marafonet:latest

dockerize:
	docker build -f deployment/Dockerfile -t marafonet:latest .


clean:
	cd cmd/marafonet && rm -f marafonet