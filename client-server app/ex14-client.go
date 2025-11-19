package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Flags to allow configurable client name, input file, and max elements
	clientName := flag.String("name", "Elena", "client name to send in header")
	const inputFile string = "data/ex14-input.txt"
	maxElements := flag.Int("max", 0, "maximum number of elements to send (0 = send all)")
	flag.Parse()

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	lines := bytes.Split(content, []byte{'\n'})
	arr := make([]string, 0, len(lines))
	for _, l := range lines {
		v := bytes.TrimSpace(l)
		if len(v) == 0 {
			continue
		}
		arr = append(arr, string(v))
	}

	// Apply maxElements limit if requested
	if *maxElements > 0 && len(arr) > *maxElements {
		arr = arr[:*maxElements]
	}

	data, err := json.Marshal(arr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal json: %v\n", err)
		os.Exit(1)
	}

	// Client-side messages
	fmt.Printf("Client %s Connected.\n", *clientName)
	fmt.Printf("Client %s made a POST request to /ex14 with data %s\n", *clientName, string(data))

	// Build request so we can add headers with client metadata
	req, err := http.NewRequest("POST", "http://localhost:8080/ex14", bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Name", *clientName)
	req.Header.Set("X-Request-Type", "POST")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("status: %s\n", resp.Status)
	fmt.Printf("body: %s\n", string(body))

	// Try to parse messages from server response
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msgs, ok := parsed["messages"].([]interface{}); ok {
			for _, m := range msgs {
				fmt.Printf("Server: %v\n", m)
			}
		}
	}

	fmt.Printf("Client %s receives response from server\n", *clientName)
}
