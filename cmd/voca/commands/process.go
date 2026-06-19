package commands

import (
	"os/exec"
	"syscall"
	"time"
)

func stopProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Signal(syscall.SIGTERM)

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		cmd.Process.Kill()
		<-ch
	}
}
