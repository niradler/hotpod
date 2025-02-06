package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

type ProcessManager struct {
	Process   *exec.Cmd
	KeepAlive bool
	Command   string
	Shell     string
	Args      []string
	ExitCode  chan int
	LogStream chan string
}

type NewProcessOpt struct {
	Command   string   `json:"command"`
	KeepAlive bool     `json:"keep_alive"`
	Shell     string   `json:"shell"`
	Args      []string `json:"args"`
}

func NewProcessManager(opt NewProcessOpt) *ProcessManager {
	return &ProcessManager{
		KeepAlive: opt.KeepAlive,
		Shell:     opt.Shell,
		Command:   opt.Command,
		Args:      opt.Args,
		ExitCode:  make(chan int, 1),
		LogStream: make(chan string, 100), // Buffered channel to prevent blocking
	}
}

func (pm *ProcessManager) StartProcess(command string, replace bool) error {
	if pm.Process != nil {
		if replace {
			pm.StopProcess()
		} else {
			return errors.New("process already running")
		}
	}

	pm.Command = command
	pm.Process = exec.Command(pm.Shell, append(pm.Args, command)...)

	stdout, err := pm.Process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stdout: %v", err)
	}
	stderr, err := pm.Process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stderr: %v", err)
	}

	// Start the process before streaming logs
	if err := pm.Process.Start(); err != nil {
		pm.Process = nil // Ensure cleanup if failed
		return fmt.Errorf("failed to start process: %v", err)
	}

	// Stream logs asynchronously
	go pm.streamOutput(stdout)
	go pm.streamOutput(stderr)

	// Wait for process completion
	go func() {
		err := pm.Process.Wait()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}
		pm.ExitCode <- exitCode
		pm.Process = nil
	}()

	return nil
}

func (pm *ProcessManager) streamOutput(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		pm.LogStream <- line
	}
	if err := scanner.Err(); err != nil {
		pm.LogStream <- fmt.Sprintf("error reading output: %v", err)
	}
}

func (pm *ProcessManager) StopProcess() error {
	if pm.Process == nil {
		return errors.New("no process running")
	}

	if err := pm.Process.Process.Signal(syscall.SIGTERM); err != nil {
		if err := pm.Process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
	}

	pm.Process = nil
	return nil
}
