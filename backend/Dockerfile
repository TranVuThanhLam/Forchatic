FROM golang:1.24

WORKDIR /app

RUN apt-get update && apt-get install -y netcat-openbsd

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN chmod +x ./scripts/wait-for-it.sh

RUN go build -o main ./cmd/web

CMD ["./scripts/wait-for-it.sh", "db", "3306", "./main"]

EXPOSE 8080

