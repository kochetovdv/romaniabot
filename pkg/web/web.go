package web

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetResponseBody makes an HTTP GET request to the specified URL and returns the response body as a byte slice.
func GetResponseBody(url string) ([]byte, error) {
    // Create a new GET request with the specified URL
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }
    
    // Set the User-Agent header to identify the client
    req.Header.Set("User-Agent", "RomanianBot/1.0")
    
    // Send the request using the default HTTP client
    client := http.DefaultClient
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error connecting to %s: %w", url, err)
    }
    
    // Close the response body after reading all the data
    defer resp.Body.Close()
    
    // Read the response body and return it as a byte slice
    return io.ReadAll(resp.Body)
}

// Ping checks the availability of a resource and returns the status code and error.
func Ping(url string, timeout time.Duration) (int, error) {
	// Create a new HTTP client with the specified timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create a new HTTP HEAD request to the specified URL
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Return an error if there was an issue creating the request
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "RomanianBot/1.0")

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		// Return an error if there was an issue connecting to the URL
		return 0, fmt.Errorf("error connecting to %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Return the status code from the response
	return resp.StatusCode, nil
}
