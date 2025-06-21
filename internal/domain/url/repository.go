package url

import "time"

type URLRepository interface {
	Save(url *URL, expires time.Duration) error
	FindByShortCode(shortCode string) (*URL, error)
}
