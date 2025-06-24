package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg"
	"github.com/robfig/cron/v3"
)

func deleteExpiredURLs(urlRepository *pg_repo.URLRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	rows, err := urlRepository.DeleteExpiredURLs(ctx)

	if err != nil {
		log.Fatalf("Error deleting expired URLs: %v", err)
	}

	log.Printf("Cleanup completed successfully. Deleted %d expired URLs.", rows)
}

func main() {
	cfg := LoadAppConfig()

	postgres := pg.NewPostgres(cfg.Env.PostgresConn)
	defer postgres.Pool.Close()

	repo := pg_repo.NewURLRepository(postgres)

	task := func() {
		deleteExpiredURLs(repo)
	}

	c := cron.New()
	c.AddFunc(fmt.Sprintf("@every %v", cfg.Env.ExpiredURLCleanupInterval.String()), task)
	task()
	c.Start()
	select {}
}
