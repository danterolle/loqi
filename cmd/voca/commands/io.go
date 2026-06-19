package commands

import (
	"io"
	"os"
	"strings"
)

func ReadInput(args []string) (string, error) {
	if len(args) > 0 {
		data, err := ReadStdinOrFile(args)
		if err == nil {
			return strings.TrimSpace(string(data)), nil
		}
		if os.IsNotExist(err) && isTerminal() {
			return strings.TrimSpace(args[0]), nil
		}
		return "", err
	}

	data, err := readStdin()
	if err != nil || data == nil {
		return "", nil
	}
	return strings.TrimSpace(string(data)), nil
}

func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func ReadStdinOrFile(args []string) ([]byte, error) {
	if len(args) > 0 {
		path := args[0]
		if path == "-" {
			return io.ReadAll(os.Stdin)
		}
		return os.ReadFile(path)
	}

	return readStdin()
}

func readStdin() ([]byte, error) {
	if isTerminal() {
		return nil, nil
	}
	return io.ReadAll(os.Stdin)
}
