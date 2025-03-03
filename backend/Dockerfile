FROM golang:1.23.4-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /go/src/bytecast

RUN go install github.com/cosmtrek/air@v1.49.0

COPY . .

RUN go mod download

RUN mkdir -p tmp

RUN echo '[build]' > .air.toml && \
    echo 'cmd = "go build -o ./tmp/main ./cmd/server/main.go"' >> .air.toml && \
    echo 'bin = "./tmp/main"' >> .air.toml && \
    echo 'include_ext = ["go", "mod"]' >> .air.toml && \
    echo 'exclude_dir = ["tmp", "vendor"]' >> .air.toml && \
    echo 'delay = 1000' >> .air.toml && \
    echo 'kill_delay = "0.5s"' >> .air.toml && \
    echo 'send_interrupt = true' >> .air.toml && \
    echo 'stop_on_error = true' >> .air.toml && \
    echo '[color]' >> .air.toml && \
    echo 'build = "yellow"' >> .air.toml && \
    echo 'main = "magenta"' >> .air.toml && \
    echo 'runner = "green"' >> .air.toml && \
    echo 'watcher = "cyan"' >> .air.toml

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]