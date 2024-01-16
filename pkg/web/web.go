package web

import (
	"fmt"
	"io"
	"net/http"
)

func GetResponseBody(url string) ([]byte, error) {
    // Создаем запрос к URL
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error during request: %v", err)
    }

    // Устанавливаем User-Agent для HTTPS
    req.Header.Set("User-Agent", "RomanianBot/1.0")

    // Создаем клиент и выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error during connect to %s: %v", url, err)
    }
    defer resp.Body.Close()

    // Читаем тело ответа
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %v", err)
    }

    return body, nil
}
