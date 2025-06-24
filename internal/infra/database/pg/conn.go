package pg

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConnection struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     int
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(postgres PostgresConnection) *Postgres {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", postgres.User, postgres.Password, postgres.Host, postgres.Port, postgres.Name)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Unable to parse Postgres connection string: v%", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 6 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to create Postgres pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to create Postgres pool: %v", err)
	}

	log.Println("Postgres connection established successfully")
	return &Postgres{
		Pool: pool,
	}
}
