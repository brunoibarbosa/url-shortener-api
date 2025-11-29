package cache

import (
	"context"
	"fmt"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/redis/go-redis/v9"
)

const StateExpiration = 2 * time.Minute

type StateRepository struct {
	client *redis.Client
}

func NewStateRepository(client *redis.Client) *StateRepository {
	return &StateRepository{
		client: client,
	}
}

func (r *StateRepository) GenerateState(ctx context.Context) (string, error) {
	state, err := session_domain.GenerateRandomState()
	if err != nil {
		return "", err
	}

	key := r.getKey(state)
	err = r.client.Set(ctx, key, "1", StateExpiration).Err()
	if err != nil {
		return "", session_domain.ErrStateGeneration
	}

	return state, nil
}

func (r *StateRepository) ValidateState(ctx context.Context, state string) error {
	if state == "" {
		return session_domain.ErrInvalidState
	}

	key := r.getKey(state)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return session_domain.ErrInvalidState
	}

	if exists == 0 {
		return session_domain.ErrInvalidState
	}

	return nil
}

func (r *StateRepository) DeleteState(ctx context.Context, state string) error {
	if state == "" {
		return nil
	}

	key := r.getKey(state)
	return r.client.Del(ctx, key).Err()
}

func (r *StateRepository) getKey(state string) string {
	key := fmt.Sprintf("oauth:state:%s", state)
	return key
}
