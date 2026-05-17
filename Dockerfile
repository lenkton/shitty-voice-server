FROM golang:1.23.3

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY frontend/ frontend/

RUN go build -o app ./cmd/server/main.go

EXPOSE 8080
EXPOSE 8443

CMD ["./app"]
