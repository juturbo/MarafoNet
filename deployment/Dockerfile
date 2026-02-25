FROM golang:1.25.7-trixie
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY frontend/build ./frontend

RUN CGO_ENABLED=0 GOOS=linux go build -o /marafonet ./cmd/marafonet

CMD ["/marafonet"]