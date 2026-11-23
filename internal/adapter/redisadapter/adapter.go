package redisadapter

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"film-searcher/internal/domain"
	"film-searcher/internal/usecase"

	"github.com/redis/go-redis/v9"
)

const (
	indexPrefix = "idx:"
	cachePrefix = "search:"
)

type Adapter struct {
	client *redis.Client
}

func New(addr string) *Adapter {
	return &Adapter{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (a *Adapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}

func (a *Adapter) Close() error {
	return a.client.Close()
}

func (a *Adapter) Build(ctx context.Context, movies []domain.Movie) error {
	if err := a.deleteByPattern(ctx, indexPrefix+"*"); err != nil {
		return err
	}

	pipe := a.client.Pipeline()
	for _, m := range movies {
		words := usecase.Tokenize(m.Title + " " + m.Director)
		seen := make(map[string]struct{}, len(words))
		for _, w := range words {
			if _, ok := seen[w]; ok {
				continue
			}
			seen[w] = struct{}{}
			pipe.SAdd(ctx, indexPrefix+w, m.ID)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (a *Adapter) SearchIDs(ctx context.Context, words []string) ([]int, error) {
	if len(words) == 0 {
		return []int{}, nil
	}

	keys := make([]string, len(words))
	for i, w := range words {
		keys[i] = indexPrefix + w
	}

	var members []string
	var err error
	if len(keys) == 1 {
		members, err = a.client.SMembers(ctx, keys[0]).Result()
	} else {
		members, err = a.client.SInter(ctx, keys...).Result()
	}
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(members))
	for _, member := range members {
		id, err := strconv.Atoi(member)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (a *Adapter) Get(ctx context.Context, key string) ([]domain.Movie, bool, error) {
	data, err := a.client.Get(ctx, cachePrefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var movies []domain.Movie
	if err := json.Unmarshal(data, &movies); err != nil {
		return nil, false, err
	}
	return movies, true, nil
}

func (a *Adapter) Set(ctx context.Context, key string, movies []domain.Movie, ttl time.Duration) error {
	data, err := json.Marshal(movies)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, cachePrefix+key, data, ttl).Err()
}

func (a *Adapter) Clear(ctx context.Context) error {
	return a.deleteByPattern(ctx, cachePrefix+"*")
}

func (a *Adapter) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, next, err := a.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := a.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}
