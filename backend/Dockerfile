FROM golang:1.23.4-alpine AS base

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# Install air for hot reloading
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Create tmp directory for air
RUN mkdir -p tmp

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]