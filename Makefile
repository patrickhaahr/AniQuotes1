.PHONY: run build test clean

all: build

run: build
	GOTMPDIR=$(GOTMPDIR) ./bin/aniquotes

build: bin
	GOTMPDIR=$(GOTMPDIR) go build -o bin/aniquotes

test:
	GOTMPDIR=$(GOTMPDIR) go test -v ./...

clean:
	rm -f bin/aniquotes

bin:
	@mkdir -p bin
