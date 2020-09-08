package util

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var file *os.File
var mut sync.Mutex

func NewTempLogger() (*os.File, error) {
	mut.Lock()
	defer mut.Unlock()
	if file == nil {
		filename := fmt.Sprintf("/tmp/rmck-%d", time.Now().Unix())
		var err error
		file, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}

func GetLogger(forComponent string) (*log.Logger, error) {
	debug := os.Getenv("RMCK_DEBUG")
	flags := log.Ldate | log.Ltime
	if debug != "" {
		return log.New(os.Stdout, forComponent, flags), nil
	}

	rmckLogPath := os.Getenv("RMCK_LOG_FILE")
	if rmckLogPath != "" {
		stat, err := os.Stat(rmckLogPath)
		if err == nil && stat.IsDir() {
			return nil, fmt.Errorf("cannot log to %s: directory", rmckLogPath)
		} else if err != nil {
			os.Remove(rmckLogPath)
		}

		file, err := os.Create(rmckLogPath)
		if err != nil {
			return nil, err
		}

		return log.New(file, forComponent, flags), nil
	}

	file, err := NewTempLogger()
	if err != nil {
		return nil, err
	}

	return log.New(file, forComponent, flags), nil
}

func LogReader(reader io.Reader, log *log.Logger) {
	go func() {
		ch := lines(reader)
		keepLogging := true
		for keepLogging {
			line, ok := <-ch
			if !ok {
				keepLogging = false
			} else {
				log.Println(line)
			}
		}
	}()
}

func LogProcess(cmd *exec.Cmd, log *log.Logger) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		stdoutCh := lines(stdout)
		stderrCh := lines(stderr)
		keepLogging := true
		for keepLogging {
			select {
			case line, ok := <-stdoutCh:
				if !ok {
					keepLogging = false
					break
				}
				log.Println(line)
			case line, ok := <-stderrCh:
				if !ok {
					keepLogging = false
					break
				}
				log.Println(line)
			}
		}
	}()
	return nil
}

func lines(stream io.Reader) <-chan string {
	ch := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		close(ch)
	}()

	return ch
}
