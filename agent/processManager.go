package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

type ProcessManager struct {
	mu        sync.Mutex
	Process   *exec.Cmd
	KeepAlive bool
	Command   string
	Shell     string
	Args      []string
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
	}
}

func (pm *ProcessManager) StartProcess(command string, replace bool) error {
	pm.mu.Lock()

	// If replace is true, stop the current process outside of the lock
	if pm.Process != nil {
		if replace {
			pm.mu.Unlock() // Unlock before stopping the process
			if err := pm.StopProcess(); err != nil {
				return fmt.Errorf("failed to stop existing process: %v", err)
			}
			pm.mu.Lock() // Reacquire the lock
		} else {
			pm.mu.Unlock() // Unlock before returning
			return errors.New("process already running")
		}
	}

	pm.Command = command
	log.Printf("Running command: %s %v\n", pm.Shell, append(pm.Args, command))

	pm.Process = exec.Command(pm.Shell, append(pm.Args, command)...)
	pm.Process.Stdout = os.Stdout
	pm.Process.Stderr = os.Stderr

	if err := pm.Process.Start(); err != nil {
		pm.Process = nil
		pm.mu.Unlock()
		return fmt.Errorf("failed to start process: %v", err)
	}

	// Handle process lifecycle in a separate goroutine
	go func() {
		err := pm.Process.Wait()

		pm.mu.Lock()
		defer pm.mu.Unlock()

		exitCode := 1
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			log.Printf("Process exited with status code: %d, reason: %v\n", exitCode, err)
		}

		// Clean up process reference
		pm.Process = nil

		if pm.KeepAlive {
			log.Println("Waiting for process...")
		} else {
			os.Exit(exitCode)
		}
	}()

	pm.mu.Unlock()
	return nil
}

func (pm *ProcessManager) StopProcess() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.Process == nil {
		log.Println("No process to stop")
		return nil
	}

	log.Println("Stopping process...")

	// gracefully ? pm.Process.Process.Signal(syscall.SIGTERM)
	if err := pm.Process.Process.Kill(); err != nil {
		return fmt.Errorf("Failed to send SIGKILL: %v", err)
	}

	pm.Process = nil

	return nil
}
