package fileutil

import (
	"log"
	"os"
)

// WriteToFile writes lines to a new file or overwrites an existing file
func WriteToFile(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("error creating file %s: %s\n", filename, err.Error())
		return err
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			log.Printf("error writing to file %s: %s\n", filename, err.Error())
			return err
		}
	}

	return nil
}
