PROGRAM_NAME=starlight-nick-server

all: test build

build:
	mkdir -p build
	go build -o ./build/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

build-race:
	mkdir -p build
	go build -race -o ./build/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/${PROGRAM_NAME}/"
	godoc -http=:6060

install-tools:
	go get -v -u honnef.co/go/tools/cmd/staticcheck

analyze: analyze-vet analyze-staticcheck

analyze-vet:
	# go vet
	go vet github.com/boreq/${PROGRAM_NAME}/...

analyze-staticcheck:
	# https://github.com/dominikh/go-tools/tree/master/cmd/staticcheck
	staticcheck github.com/boreq/${PROGRAM_NAME}/...

test:
	go test ./...

test-verbose:
	go test -v -count 1 ./...

clean:
	rm -rf ./build

.PHONY: all build build-race doc install-tools analyze analyze-vet analyze-staticcheck test test-verbose clean
