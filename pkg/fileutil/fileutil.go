package fileutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
)

// WriteToFile writes lines to a new file or overwrites an existing file
func WriteToFileStrings(filename string, lines []string) error {
	options := os.O_WRONLY | os.O_CREATE
	mode := int(0777)
	file, err := os.OpenFile(filename, options, os.FileMode(mode))
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

// Чтение из файлов строк. Каждая строка - элемент среза
func ReadfromFile(filename string) []string {
	var lines []string

	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil
	}
	ErrChecking(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	ErrChecking(scanner.Err())

	return lines
}

func ReadBytesfromFile(filename string) []byte {
	var bytes []byte

	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil
	}
	ErrChecking(err)
	defer file.Close()

	bytes, err = io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	return bytes
}

func CheckDir(path string) bool {
	if len(path) > 0 {
		err := os.Mkdir(path, 0750)
		if err != nil && !os.IsExist(err) {
			ErrChecking(err)
		}
	}
	return true
}

// Если файл существует, возвращается true
func CheckFile(fname string) bool {
	_, err := os.Stat(fname)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
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

// Проверяем ошибки отдельной функцией чтобы не плодить код
func ErrChecking(err error) {
	if err != nil {
		fmt.Println("oops")
		log.Fatal(err)
	}
}

func IsExtension(fname string, ext string) bool {
	// Открываем файл
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	// Проверяем тип файла
	mimeType := mime.TypeByExtension(ext)
	return mimeType == ext
}
