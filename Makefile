run:
	go run ./src

debug:
	go run ./src -v

trace:
	go run ./src -vv

setup:
	go get ./...
	go install github.com/dmarkham/enumer@latest
	go generate ./...

build:
	go build ./...

test:
	go test ./...

update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run
