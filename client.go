package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("movie not found")

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) FetchMovie(ctx context.Context, id int) (*Movie, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stand returned status %d: %s", resp.StatusCode, string(body))
	}

	var movie Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}
	if movie.ID == 0 {
		movie.ID = id
	}
	return &movie, nil
}

func (c *Client) FindLastMovie(ctx context.Context) (*Movie, error) {
	left := 1
	right := 100000
	lastFound := -1

	for left <= right {
		mid := (left + right) / 2
		_, err := c.FetchMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				right = mid - 1
				continue
			}
			return nil, err
		}
		lastFound = mid
		left = mid + 1
	}

	if lastFound < 1 {
		return nil, ErrNotFound
	}

	return c.FetchMovie(ctx, lastFound)
}
