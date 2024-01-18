package fileutil

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
)

// WriteToFileStrings writes lines to a new file or overwrites an existing file
func WriteToFileStrings(filename string, lines []string) error {
	// Open the file with write-only and create mode, using 0777 as the file permission
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.FileMode(0777))
	if err != nil {
		// Return an error if there was a problem creating the file
		return fmt.Errorf("error creating file %s: %w", filename, err)
	}
	defer file.Close()

	// Create a buffered writer to improve write performance
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		// Write each line to the file, appending a newline character
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			// Return an error if there was a problem writing to the file
			return fmt.Errorf("error writing to file %s: %w", filename, err)
		}
	}
	// Flush the writer to ensure all buffered data is written to the file
	err = writer.Flush()
	if err != nil {
		// Return an error if there was a problem flushing the writer
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

// ReadFromFile reads lines from a file and returns them as a slice of strings.
func ReadFromFile(filename string) ([]string, error) {
	// Open the file with the given filename
	file, err := os.Open(filename)
	if err != nil {
		// If the file does not exist, return an empty slice
		if os.IsNotExist(err) {
			return nil, nil
		}
		// If there was an error opening the file, return the error
		return nil, err
	}
	// Close the file when we're done
	defer file.Close()

	// Initialize an empty slice to store the lines
	var lines []string
	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	// Read the file line by line until there are no more lines
	for scanner.Scan() {
		// Append each line to the slice
		lines = append(lines, scanner.Text())
	}

	// Check if there was an error scanning the file
	if err := scanner.Err(); err != nil {
		// If there was an error, return the error
		return nil, err
	}

	// Return the lines as a slice and no error
	return lines, nil
}

// ReadBytesfromFile reads the contents of a file and returns them as a byte slice.
// If the file does not exist or fails to open, it returns an error.
func ReadBytesfromFile(filename string) ([]byte, error) {
	// Open the file with the given filename
	file, err := os.Open(filename)
	// If the file does not exist, return an error
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}
	// If there was an error opening the file, return an error
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	// Ensure the file is closed when this function returns
	defer file.Close()

	// Read the contents of the file into a byte slice
	bytes, err := io.ReadAll(file)
	// If there was an error reading the file, return an error
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	// Return the byte slice containing the file contents
	return bytes, nil
}

// CheckDir checks if a directory exists at the given path, and creates it if it doesn't.
// Returns true if the directory exists or was successfully created.
func CheckDir(path string) bool {
	// Check if the path is not empty
	if len(path) > 0 {
		// Create the directory with the given path and permissions 0750
		err := os.Mkdir(path, 0750)
		// Check if there was an error creating the directory and it's not because the directory already exists
		if err != nil && !os.IsExist(err) {
			// Handle the error by calling ErrChecking function
			ErrChecking(err)
		}
	}
	// Return true to indicate that the directory exists or was successfully created
	return true
}

// CheckFile checks if a file exists and returns true if it does, false otherwise.
func CheckFile(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

// Сохранение в файл байтовой информацией. Файл создается с нуля, не дополняется
func WriteToFile(pathForSave, filename string, b []byte) error {
	options := os.O_WRONLY | os.O_CREATE
	mode := int(0777)
	file, err := os.OpenFile(pathForSave+filename, options, os.FileMode(mode))
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	fmt.Printf("File %s saved.\n", filename)
	return nil
}

// The ErrChecking function is used to check if an error occurred.
// If an error is present, it prints "oops" and exits the program with the error message.
func ErrChecking(err error) {
	// Check if an error occurred
	if err != nil {
		// Print "oops" to indicate an error
		fmt.Println("oops")
		// Exit the program with the error message
		log.Fatal(err)
	}
}

// IsExtension checks if the given file has the specified extension.
func IsExtension(fname string, ext string) bool {
	// Open the file with the given filename.
	file, err := os.Open(fname)
	if err != nil {
		log.Println(err)
		return false
	}
	defer file.Close()

	// Get the MIME type based on the provided extension.
	mimeType := mime.TypeByExtension(ext)

	// Check if the MIME type matches the provided extension.
	return mimeType == ext
}
