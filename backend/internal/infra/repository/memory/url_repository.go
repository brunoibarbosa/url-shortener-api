package memory_repo

import (
	"sync"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
)

var (
	urlStore = make(map[string]string)
	mu       sync.Mutex
)

type URLRepository struct {
}

func NewURLRepository() *URLRepository {
	return &URLRepository{}
}

func (r *URLRepository) Save(url *domain.URL) error {
	mu.Lock()
	defer mu.Unlock()
	urlStore[url.ShortCode] = url.EncryptedURL
	return nil
}

func (r *URLRepository) FindByShortCode(shortCode string) (*domain.URL, error) {
	mu.Lock()
	defer mu.Unlock()
	encryptedUrl, exists := urlStore[shortCode]

	if !exists {
		return nil, nil
	}

	return &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedUrl,
	}, nil
}
