package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type PodManagerClient struct {
	PodName   string
	Namespace string
	Port      int
}

func NewPodManagerClient(podName, namespace string, port int) *PodManagerClient {
	return &PodManagerClient{
		PodName:   podName,
		Namespace: namespace,
		Port:      port,
	}
}

func (c *PodManagerClient) StartPortForward() (*exec.Cmd, error) {
	cmd := exec.Command("kubectl", "port-forward", "-n", c.Namespace, c.PodName, fmt.Sprintf("%d:%d", c.Port, c.Port))
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start port-forward: %v", err)
	}
	return cmd, nil
}

func (c *PodManagerClient) SendRequest(endpoint string, method string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:%d/%s", c.Port, endpoint)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func main() {
	var podName, namespace string
	var port int

	var rootCmd = &cobra.Command{
		Use:   "pod-manager-cli",
		Short: "A CLI tool to manage processes in Kubernetes pods",
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the main process in the pod",
		Run: func(cmd *cobra.Command, args []string) {
			client := NewPodManagerClient(podName, namespace, port)
			pfCmd, err := client.StartPortForward()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer pfCmd.Process.Kill()

			requestBody, _ := json.Marshal(map[string]string{
				"command": strings.Join(args, " "),
			})

			resp, err := client.SendRequest("start", http.MethodPost, bytes.NewBuffer(requestBody))
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Fatalf("Failed to start process: %s", resp.Status)
			}

			fmt.Println("Process started successfully")
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the main process in the pod",
		Run: func(cmd *cobra.Command, args []string) {
			client := NewPodManagerClient(podName, namespace, port)
			pfCmd, err := client.StartPortForward()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer pfCmd.Process.Kill()

			resp, err := client.SendRequest("stop", http.MethodPost, nil)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Fatalf("Failed to stop process: %s", resp.Status)
			}

			fmt.Println("Process stopped successfully")
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Get the status of the main process in the pod",
		Run: func(cmd *cobra.Command, args []string) {
			client := NewPodManagerClient(podName, namespace, port)
			pfCmd, err := client.StartPortForward()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer pfCmd.Process.Kill()

			resp, err := client.SendRequest("status", http.MethodGet, nil)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			defer resp.Body.Close()

			var status map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
				log.Fatalf("Error decoding response: %v", err)
			}

			fmt.Println("Process status:")
			for k, v := range status {
				fmt.Printf("%s: %v\n", k, v)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&podName, "pod", "p", "", "Name of the pod")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the pod")
	rootCmd.PersistentFlags().IntVar(&port, "port", 8080, "Port for the Pod Manager API")

	rootCmd.AddCommand(startCmd, stopCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
