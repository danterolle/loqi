.PHONY: run build stop clean

BINARY = voca

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

stop:
	-pkill -f "$(BINARY)" 2>/dev/null; pkill ollama 2>/dev/null; echo "Stopped."

clean:
	rm -f $(BINARY)
