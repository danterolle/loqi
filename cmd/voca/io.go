package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func readInput(args []string) (string, error) {
	if len(args) > 0 {
		path := args[0]
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("reading file %q: %w", path, err)
			}
			return strings.TrimSpace(string(data)), nil
		} else if err != nil && looksLikeFilePath(path) {
			return "", fmt.Errorf("file not found: %q", path)
		}
		return strings.Join(args, " "), nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", nil
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", nil
}

func readStdinOrFile(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("stdin not available: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input file specified and stdin is a terminal; pipe data or provide a file path")
	}
	return io.ReadAll(os.Stdin)
}

func looksLikeFilePath(path string) bool {
	for i := len(path) - 1; i >= 0; i-- {
		switch path[i] {
		case '/', '\\':
			return false
		case '.':
			return i > 0 && i < len(path)-1
		}
	}
	return false
}
