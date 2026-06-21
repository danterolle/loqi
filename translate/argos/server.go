package argos

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const DefaultBaseURL = "http://localhost:5000"

//go:embed argos_server.py
var serverScript string

const scriptName = "loqi_argos_server.py"

func venvDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "loqi-argos-venv")
	}
	return filepath.Join(home, ".cache", "loqi", "argos-venv")
}

func venvPython() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvDir(), "Scripts", "python.exe")
	}
	return filepath.Join(venvDir(), "bin", "python3")
}

func findPython() (string, error) {
	if runtime.GOOS == "windows" {
		p, err := exec.LookPath("python")
		if err != nil {
			return "", fmt.Errorf("argos: python not found in PATH — install Python from https://python.org")
		}
		return p, nil
	}
	p, err := exec.LookPath("python3")
	if err != nil {
		p, err = exec.LookPath("python")
		if err != nil {
			return "", fmt.Errorf("argos: python3 not found in PATH — install Python from https://python.org")
		}
	}
	return p, nil
}

func ensureVenv() error {
	vp := venvPython()
	if _, err := os.Stat(vp); err == nil {
		return nil
	}

	python, err := findPython()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "  ◆ Creating Python virtual environment...\n")
	if out, err := exec.Command(python, "-m", "venv", venvDir()).CombinedOutput(); err != nil {
		return fmt.Errorf("argos: create venv: %w\n%s", err, out)
	}

	fmt.Fprintf(os.Stderr, "  ◆ Installing argostranslate...\n")
	if out, err := exec.Command(vp, "-m", "pip", "install", "argostranslate").CombinedOutput(); err != nil {
		return fmt.Errorf("argos: pip install failed: %w\n%s", err, out)
	}

	return nil
}

func Reachable(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	conn, err := net.DialTimeout("tcp", u.Host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func StartServer(baseURL string) (*exec.Cmd, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("argos: invalid base_url: %w", err)
	}

	if err := ensureVenv(); err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(os.TempDir(), scriptName)
	if err := os.WriteFile(scriptPath, []byte(serverScript), 0644); err != nil {
		return nil, fmt.Errorf("argos: write server script: %w", err)
	}

	port := u.Port()
	cmd := exec.Command(venvPython(), "-W", "ignore::UserWarning", scriptPath, port)
	cmd.Stderr, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("argos: start server: %w", err)
	}

	for range 120 {
		if Reachable(baseURL) {
			return cmd, nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	cmd.Process.Kill()
	cmd.Wait()
	return nil, fmt.Errorf("argos: server did not start after 60s — check pip install output above")
}
