FROM golang:1.25.7-trixie
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/marafonet/main.go ./
COPY frontend/build ./frontend

RUN CGO_ENABLED=0 GOOS=linux go build -o /marafonet main.go

CMD ["/marafonet"]