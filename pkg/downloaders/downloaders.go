package downloaders

import (
	"io"
	"log"
	"net/http"
	"romaniabot/pkg/fileutil"
	"romaniabot/pkg/web"
	"sync"
	"time"
)

// CheckDownloadedFiles is a function that checks if files have been downloaded.
// It takes a path to save the files and a list of files to check.
// It returns a slice of downloaded files.
func CheckDownloadedFiles(pathForSave string, filesToCheck []string) []string {
	// Check if the directory exists. If it doesn't exist, create it automatically and skip checking for files in it.
	var downloadedFiles []string
	if fileutil.CheckDir(pathForSave) {
		// Iterate over the files to check.
		for _, k := range filesToCheck {
			// Check if the file exists in the specified path.
			if fileutil.CheckFile(pathForSave + k) {
				// If the file exists, add it to the downloadedFiles slice.
				downloadedFiles = append(downloadedFiles, k)
			}
		}
	}
	// Return the slice of downloaded files.
	return downloadedFiles
}

// CheckBrokenURLs checks the availability of URLs and returns a slice of broken URLs
func CheckBrokenURLs(URLs []string, maxRetries int, timeout time.Duration) []string {
	// Create a slice to store the broken URLs
	var brokenURLs []string

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a channel to communicate the results of the goroutines
	ch := make(chan string, len(URLs))

	// Iterate over each URL in the input slice
	for _, url := range URLs {
		// Add 1 to the WaitGroup counter
		wg.Add(1)

		// Create a goroutine to check the availability of the URL
		go func(u string) {
			// Mark the goroutine as done when it finishes
			defer wg.Done()

			// Set the number of retries to the maximum value
			retries := maxRetries

			// Retry the request until the maximum number of retries is reached
			for retries > 0 {
				// Ping the URL and get the status code and error
				status, err := web.Ping(u, timeout)

				// Check if there was an error or if the status code indicates a broken URL
				if err != nil || status == 0 || status > 399 {
					// Decrement the number of retries
					retries--

					// Sleep for the specified timeout before retrying
					time.Sleep(timeout)
				} else {
					// If the URL is available, send an empty string to the channel and return
					ch <- ""
					return
				}
			}

			// If the URL is broken after all retries, send it to the channel
			ch <- u
		}(url)
	}

	// Wait until all goroutines have finished
	wg.Wait()

	// Close the channel to signal that no more values will be sent
	close(ch)

	// Iterate over the results received from the channel
	for result := range ch {
		// If the result is not an empty string, it means the URL is broken
		if result != "" {
			// Append the broken URL to the slice
			brokenURLs = append(brokenURLs, result)
		}
	}

	// Return the slice of broken URLs
	return brokenURLs
}

// map[filename]url
// Скачивает файлы. Получает путь для сохранения файлов и карту, состоящую из наименования файла для сохранения и ссылки на скачивание
func Downloader(pathForSave string, filesURLS map[string]string) {
	download(pathForSave, filesURLS)
}

// Refactored download function
func download(pathForSave string, filesURLS map[string]string) {
	var wg sync.WaitGroup // Create a wait group to wait for all goroutines to finish

	wg.Add(len(filesURLS)) // Add the number of files to the wait group

	var mu sync.Mutex // Create a mutex to synchronize access to shared resources

	// Iterate over each file URL in the map
	for fname, url := range filesURLS {
		go func(fname, url string) { // Create a goroutine to download and save the file
			defer wg.Done() // Notify the wait group that the goroutine has finished

			req, err := http.NewRequest("GET", url, nil) // Create a new GET request
			if err != nil {
				log.Printf("error during request: %v\n", err) // Log any errors during request creation
				return
			}
			req.Header.Set("User-Agent", "RomanianBot/1.0") // Set the User-Agent header

			client := &http.Client{}    // Create a new HTTP client
			resp, err := client.Do(req) // Send the request and get the response
			if err != nil {
				log.Printf("error during connect to %s: %v\n", url, err) // Log any errors during connection
				return
			}
			defer resp.Body.Close() // Close the response body when finished

			body, err := io.ReadAll(resp.Body) // Read the response body
			if err != nil {
				log.Printf("error reading response body: %v\n", err) // Log any errors during reading response body
				return
			}

			mu.Lock()         // Acquire the lock to synchronize access to shared resources
			defer mu.Unlock() // Release the lock when finished

			err = fileutil.WriteToFile(pathForSave, fname, body) // Write the file to disk
			if err != nil {
				log.Printf("error during writing file: %v\n", err) // Log any errors during writing file
				return
			}
		}(fname, url) // Pass the file name and URL to the goroutine
	}

	wg.Wait()                                      // Wait for all goroutines to finish
	log.Println("All files downloaded and saved.") // Log a message indicating that all files have been downloaded and saved
}
