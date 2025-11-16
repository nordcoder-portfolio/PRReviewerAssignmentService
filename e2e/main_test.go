package e2e

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var httpClient *Client

func TestMain(m *testing.M) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient = NewClient(baseURL)

	if err := waitForHealth(baseURL, 30*time.Second); err != nil {
		log.Printf("service did not become healthy on %s: %v", baseURL, err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func waitForHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("service not healthy within %s", timeout)
}
