package adapter

import (
	"context"
	"encoding/json"
	"film-searcher/internal/domain"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Movie{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.Movie{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Movie{}, domain.ErrMovieNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return domain.Movie{}, fmt.Errorf("stand returned status %d: %s", resp.StatusCode, string(body))
	}

	var movie domain.Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return domain.Movie{}, err
	}
	if movie.ID == 0 {
		movie.ID = id
	}
	return movie, nil
}
