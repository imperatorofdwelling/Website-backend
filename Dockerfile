FROM golang:1.21.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o migrate ./cmd/migrator/main.go
RUN go build -o bin ./cmd/app/main.go

CMD migrate

ENTRYPOINT ["/app/bin"]
