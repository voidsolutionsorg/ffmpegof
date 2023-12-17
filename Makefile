setup:
	go get ./...

build:
	go build ./...

test:
	go test ./...

update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run