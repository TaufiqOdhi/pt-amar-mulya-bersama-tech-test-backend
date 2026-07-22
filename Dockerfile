FROM golang:1.22-bookworm

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application binary
RUN go build -o /app/server ./cmd/api

EXPOSE 8080

CMD ["/app/server"]
