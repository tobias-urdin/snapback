GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: compile

compile:
	GCGO_ENABLE=1 DOOS=linux GOARCH=amd64 $(GOBUILD) -o build/snapback-amd64 cmd/main.go

test:
	$(GOTEST) ./...
