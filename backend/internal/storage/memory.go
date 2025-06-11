package storage

import "sync"

var (
	urlStore = make(map[string]string)
	mu       sync.Mutex
)

func SaveURL(shortId, encryptedUrl string) {
	mu.Lock()
	defer mu.Unlock()
	urlStore[shortId] = encryptedUrl
}

func GetURL(shortId string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	originalUrl, exists := urlStore[shortId]
	return originalUrl, exists
}
