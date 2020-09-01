package util

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

// Waits for cmd to exit using the supplied context
func WaitWithContext(ctx context.Context, cmd *exec.Cmd) error {
	proc, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return err
	}

	isRunning, err := proc.IsRunning()
	if err != nil {
		return err
	}

	done := ctx.Done()
	for isRunning {
		select {
		case <-done:
			return ctx.Err()
		default:
			isRunning, err := proc.IsRunning()
			if err != nil {
				return err
			} else if isRunning {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	return cmd.Wait()
}

// Returns if a process matching name is running
func IsRunning(name string) (bool, error) {
	procs, err := process.Processes()
	if err != nil {
		return false, err
	}
	for _, proc := range procs {
		pname, err := proc.Name()
		if err != nil {
			continue
		}
		if strings.Contains(pname, name) {
			return proc.IsRunning()
		}
	}
	return false, fmt.Errorf("process %s not found", name)
}

// Attemps to cleanly shutdown the process
func GracefulStop(cmd *exec.Cmd, ctx context.Context) error {
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("GracefulStop(): SIGINT: %s", err)
	}
	nctx, cc := context.WithTimeout(ctx, time.Second)
	defer cc()
	if err := WaitWithContext(nctx, cmd); err == nil {
		return fmt.Errorf("GracefulStop(): Wait: %s", err)
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("GracefulStop(): SIGTERM: %s", err)
	}
	nctx, cc = context.WithTimeout(ctx, time.Second)
	defer cc()
	if err := WaitWithContext(nctx, cmd); err == nil {
		return nil
	}

	if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("GracefulStop(): SIGKILL: %s", err)
	}

	return cmd.Wait()
}
