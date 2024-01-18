package downloaders

import (
	"fmt"
	"io"
	"net/http"
	"romaniabot/pkg/fileutil"
	"romaniabot/pkg/web"
	"sync"
	"time"
)

var mu sync.Mutex

// []filename
// Проверяем, скачаны ли файлы. Возвращает срез скачанных файлов
func CheckDownloadedFiles(pathForSave string, filesToCheck []string) []string {
	// Проверяем существование папки. Если не существует, создается автоматически и наличие файлов в ней пропускается.
	var downloadedFiles []string
	if fileutil.CheckDir(pathForSave) {
		for _, k := range filesToCheck {
			if fileutil.CheckFile(pathForSave + k) {
				downloadedFiles = append(downloadedFiles, k)
			}
		}
	}
	return downloadedFiles
}

// []url
// Проверяем доступность ссылки и возвращаем мапу со сломанными ссылками
func CheckBrokenURLs(URLS []string, maxRetries int, timeout time.Duration) []string {
	var brokenURLs []string
	var wg sync.WaitGroup
	// Используем канал для синхронизации горутин
	ch := make(chan string)

	for _, v := range URLS {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			retries := maxRetries
			for retries > 0 {
				status, err := web.Ping(url, timeout)
				if err != nil || status == 0 || status > 399 {
					retries--
					time.Sleep(timeout)
				} else {
					// Успешный запрос - отправляем URL в канал
					ch <- ""
					return
				}
			}
			// Если все попытки неудачны, добавляем URL в сломанные
			ch <- url
		}(v)
	}

	// Горутины завершили работу, закрываем канал после завершения всех
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Собираем результаты из канала
	for result := range ch {
		if result != "" {
			mu.Lock()
			brokenURLs = append(brokenURLs, result)
			mu.Unlock()
		}
	}

	return brokenURLs
}




// map[filename]url
// Скачивает файлы. Получает путь для сохранения файлов и карту, состоящую из наименования файла для сохранения и ссылки на скачивание
func Downloader(pathForSave string, filesURLS map[string]string) {
	download(pathForSave, filesURLS)
}

// map[filename]url
func download(pathForSave string, filesURLS map[string]string) {
	wg := sync.WaitGroup{}
	wg.Add(len(filesURLS))

	mu := sync.Mutex{} // Мьютекс для синхронизации доступа к файловой системе

	for fname, url := range filesURLS {
		go func(fname, url string) {
			defer wg.Done()

			// Создаем запрос к URL
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("error during request: %v\n", err)
				return
			}

			// Устанавливаем User-Agent для HTTPS
			req.Header.Set("User-Agent", "RomanianBot/1.0")

			// Создаем клиент и выполняем запрос
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("error during connect to %s: %v\n", url, err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("error closing response body: %v\n", err)
				}
			}()

			// Читаем тело ответа
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("error reading response body: %v\n", err)
				return
			}

			// Используем мьютекс для синхронизации записи в файл
			mu.Lock()
			defer mu.Unlock()

			// Записываем тело ответа в файл
			err = fileutil.WriteToFile(pathForSave, fname, body)
			if err != nil {
				fmt.Printf("error during writing file: %v\n", err)
				return
			}
		}(fname, url)
	}

	wg.Wait()
	fmt.Println("All files downloaded and saved.")
}

