package url

type URLRepository interface {
	Save(url *URL) error
	FindByShortCode(shortCode string) (*URL, error)
}
