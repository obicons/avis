package util

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

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
