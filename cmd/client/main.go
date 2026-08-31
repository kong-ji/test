// Command client is a CLI client for the Banner fingerprint identification
// service. It reads a JSON array of [{ip,port,banner}], posts it to the
// server's /fingerprint endpoint, and prints the identification results
// as pretty-printed JSON.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Input is a single banner sample to identify.
type Input struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Banner string `json:"banner"`
}

// Result is a single identification result returned by the server.
type Result struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	OSHint     string  `json:"os_hint"`
	Confidence float64 `json:"confidence"`
}

func main() {
	os.Exit(run())
}

// envOrDefault returns the value of key, or fallback when unset/empty.
func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// run performs the actual work and returns the process exit code.
func run() int {
	var (
		server  = flag.String("server", envOrDefault("SERVER_URL", "http://localhost:8080"), "server base URL")
		file    = flag.String("file", envOrDefault("INPUT_FILE", ""), "input JSON file path (optional; defaults to stdin)")
		timeout = flag.Duration("timeout", 10*time.Second, "HTTP request timeout")
	)
	flag.Parse()

	var data []byte
	var err error
	if *file != "" {
		data, err = os.ReadFile(*file)
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "read input:", err)
		return 1
	}

	var inputs []Input
	if err := json.Unmarshal(data, &inputs); err != nil {
		fmt.Fprintln(os.Stderr, "decode input:", err)
		return 1
	}

	payload, err := json.Marshal(inputs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encode request:", err)
		return 1
	}

	endpoint := strings.TrimRight(*server, "/") + "/fingerprint"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintln(os.Stderr, "build request:", err)
		return 1
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: *timeout}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send request:", err)
		return 1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read response:", err)
		return 1
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "server returned %s: %s\n", resp.Status, strings.TrimSpace(string(body)))
		return 1
	}

	var results []Result
	if err := json.Unmarshal(body, &results); err != nil {
		fmt.Fprintln(os.Stderr, "decode response:", err)
		return 1
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "encode output:", err)
		return 1
	}
	fmt.Println(string(out))

	return 0
}
