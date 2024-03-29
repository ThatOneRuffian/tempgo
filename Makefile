# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
RESOURCES = "resources"
BINARY_NAME = tempgo
MAIN_PATH = .

all: test build

build:
	~/go/bin/fyne bundle -o bundled.go $(RESOURCES)
	CGO_ENABLED=1 $(GOBUILD) -buildmode=pie -v  -o $(BINARY_NAME) $(MAIN_PATH)

build-arm:
	~/go/bin/fyne bundle -o bundled.go $(RESOURCES)
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) -v -buildmode=pie -o $(BINARY_NAME)_linux_arm64 $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

.PHONY: all build test clean bundle run
