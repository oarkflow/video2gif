.PHONY: build run test clean docker

BINARY  = video2gif
MODULE  = github.com/oarkflow/video2gif
VERSION = 1.0.0

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/video2gif

run: build
	./$(BINARY) -config config.json

run-cli: build
	./$(BINARY) -cli -input $(INPUT) -output $(OUTPUT) -profile $(PROFILE)

test:
	go test ./... -v -race -timeout 60s

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf uploads/ outputs/ tmp/

tidy:
	go mod tidy

docker:
	docker build -t $(BINARY):$(VERSION) .

.PHONY: setup
setup:
	mkdir -p uploads outputs tmp
	go mod tidy
