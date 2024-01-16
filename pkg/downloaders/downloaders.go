package downloaders

import (
	"fmt"
	"romaniabot/pkg/fileutil"
	"romaniabot/pkg/web"
	"sync"
)

// Скачивает файлы. Получает путь для сохранения файлов и карту, состоящую из наименования файла для сохранения и ссылки на скачивание
func Downloader(pathForSave string, filesURLS map[string]string) {
	_ = fileutil.CheckDir(pathForSave)
	for k:= range filesURLS {
		if fileutil.CheckFile(pathForSave + k) {
			delete(filesURLS, k)
			fmt.Printf("Deleted: %s\n", k)
		}
	}

	download(pathForSave, filesURLS)
}

func download(pathForSave string, filesURLS map[string]string){
	wg := sync.WaitGroup{}
	wg.Add(len(filesURLS))

	for fname, url := range filesURLS {
		go func(fname, url string) {
			defer wg.Done()

			buf, err := web.GetResponseBody(url)
			if err != nil {
				fmt.Printf("Error during reading file to buffer: %v\n", err)
				return
			}

			fileutil.WriteToFile(pathForSave, fname, buf)
			if err != nil {
				fmt.Printf("Error during writing file: %v\n", err)
				return
			}
		}(fname, url)
	}

	wg.Wait()
	fmt.Println("All files downloaded and saved.")
}