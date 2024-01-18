package web

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetResponseBody получает на вход ссылку, возвращает тело ответа в байтах и ошибку
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

// Ping проверяет доступность ресурса и возвращает код статуса и ошибку
func Ping(url string, timeout time.Duration) (int, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "RomanianBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error connecting to %s: %v", url, err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
