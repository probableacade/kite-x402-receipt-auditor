package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	endpoint := flag.String("endpoint", "http://localhost:8080", "auditor base URL")
	file := flag.String("file", "", "receipt JSON file (defaults to stdin)")
	flag.Parse()
	if err := run(*endpoint, *file, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(endpoint, file string, stdin io.Reader, stdout io.Writer) error {
	var input io.Reader = stdin
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}
	body, err := io.ReadAll(io.LimitReader(input, 1<<20))
	if err != nil {
		return err
	}
	if !json.Valid(body) {
		return fmt.Errorf("input is not valid JSON")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(endpoint+"/v1/audits", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, response, "", "  "); err != nil {
		return fmt.Errorf("unexpected server response: %s", response)
	}
	_, _ = fmt.Fprintln(stdout, formatted.String())
	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit rejected with HTTP %d", resp.StatusCode)
	}
	return nil
}
