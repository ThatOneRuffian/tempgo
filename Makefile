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
	fyne bundle -o bundled.go $(RESOURCES)
	CGO_ENABLED=1 $(GOBUILD) -buildmode=pie -v  -o $(BINARY_NAME) $(MAIN_PATH)

build-arm:
	fyne bundle -o bundled.go $(RESOURCES)
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)_arm64 -v $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

.PHONY: all build test clean bundle run
