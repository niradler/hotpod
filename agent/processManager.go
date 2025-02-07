package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
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

func (pm *ProcessManager) StartProcess(command string, replace bool, customEnv *[]string) error {
	pm.mu.Lock()

	if pm.Process != nil {
		if replace {
			log.Println("Replacing existing process (PID:", pm.Process.Process.Pid, ")...")

			// Unlock before stopping the process
			pm.mu.Unlock()
			if err := pm.StopProcess(); err != nil {
				return fmt.Errorf("failed to stop existing process: %v", err)
			}
			pm.mu.Lock() // Reacquire the lock
		} else {
			pm.mu.Unlock()
			return errors.New("process already running")
		}
	}

	log.Printf("Starting process: %s %v %s\n", pm.Shell, pm.Args, command)
	pm.Command = command
	pm.Process = exec.Command(pm.Shell, append(pm.Args, command)...)
	pm.Process.Stdout = os.Stdout
	pm.Process.Stderr = os.Stderr
	env := os.Environ()
	if customEnv != nil {
		log.Println("Custom environment variables detected, applying...", customEnv)
		env = append(env, *customEnv...)
	}
	pm.Process.Env = env

	if err := pm.Process.Start(); err != nil {
		pm.Process = nil
		pm.mu.Unlock()
		return fmt.Errorf("failed to start process: %v", err)
	}
	log.Println("Process started with PID:", pm.Process.Process.Pid)
	// Handle process lifecycle in a separate goroutine
	go func(proc *exec.Cmd, pid int) {
		err := proc.Wait()

		pm.mu.Lock()
		defer pm.mu.Unlock()

		exitCode := 1
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			log.Printf("Process (PID: %d) exited with status code: %d, reason: %v\n", pid, exitCode, err)
		}

		// Ensure process is properly cleaned up
		if pm.Process == proc {
			pm.Process = nil
		}

		if pm.KeepAlive {
			log.Println("Process exited but KeepAlive is enabled.")
		} else {
			os.Exit(exitCode)
		}
	}(pm.Process, pm.Process.Process.Pid)

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

	pid := pm.Process.Process.Pid
	log.Println("Stopping process (PID:", pid, ")...")

	if err := pm.Process.Process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("Failed to send SIGTERM to PID %d: %v, trying SIGKILL", pid, err)
		if err := pm.Process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %v", pid, err)
		}
	}

	// Wait for process to exit completely
	done := make(chan struct{})
	go func() {
		pm.Process.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Process (PID:", pid, ") stopped successfully")
	case <-time.After(5 * time.Second):
		log.Println("Process (PID:", pid, ") did not exit in time, force killing...")
		pm.Process.Process.Kill()
	}

	pm.Process = nil
	return nil
}
