## Build the project
build:
	go build -o mk .

## Run all tests
test:
	go test ./...

## Run static analysis
vet:
	go vet ./...

## Clean build artifacts
clean:
	rm -f mk
	rm -rf dist/

## Build for all platforms
dist: clean
	GOOS=linux   GOARCH=amd64 go build -o dist/mk-linux .
	GOOS=darwin  GOARCH=amd64 go build -o dist/mk-mac .
	GOOS=darwin  GOARCH=arm64 go build -o dist/mk-mac-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/mk.exe .

## Format source code
fmt:
	go fmt ./...

## Run tests with coverage
coverage:
	go test -cover ./...

## Install to GOPATH/bin
install:
	go install .
