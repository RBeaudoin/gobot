.PHONY: build clean doc fmt get lint test vet install bench


all: clean \
	get \
	lint \
	vet \
	test \
	install

get:
	go get -t -v ./...

install:
	go install

build: 
	go build

clean:
	go clean

doc:
	godoc -http=:6060

lint:
	golint ./...

test:
	go test ./... -cover

bench:
	go test ./... -bench=.

vet:
	go vet ./...
