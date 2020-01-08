.PHONY: vet

all: debuggo

debuggo: go.mod go.sum $(wildcard *.go **/*.go)
	go build -o debuggo

vet:
	go vet ./...

test: vet
	go test ./...

coverage.txt: vet go.mod go.sum
	go test ./... -race -coverprofile=$@ -covermode=atomic

clean:
	rm -f debuggo
