FROM golang:1.22-bookworm

# Install development & text editor tools
RUN apt-get update && apt-get install -y \
    curl \
    git \
    bash \
    vim \
    && rm -rf /var/lib/apt/lists/*

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
