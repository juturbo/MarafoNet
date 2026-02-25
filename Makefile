all:
	cd frontend && npm run build
	go run cmd/marafonet/main.go

react:
	cd frontend && npm start