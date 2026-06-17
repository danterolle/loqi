.PHONY: run build stop clean

BINARY = voca
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

stop:
	-pkill -f "$(BINARY)" 2>/dev/null; pkill ollama 2>/dev/null; echo "Stopped."

clean:
	rm -f $(BINARY)
