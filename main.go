package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Config struct {
	From    int
	To      int
	Workers int
	Timeout time.Duration
}

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type Result struct {
	Movie *Movie
	Error error
}

func fetchMovie(ctx context.Context, client *http.Client, id int) Result {
	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{Error: fmt.Errorf("id %d: request error: %w", id, err)}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{Error: fmt.Errorf("id %d: network error: %w", id, err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{Error: fmt.Errorf("id %d: HTTP status %d", id, resp.StatusCode)}
	}

	var movie Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return Result{Error: fmt.Errorf("id %d: json error: %w", id, err)}
	}

	return Result{Movie: &movie}
}

func worker(ctx context.Context, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	client := &http.Client{Timeout: timeout}

	for id := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		results <- fetchMovie(ctx, client, id)
	}
}

func parseFlags() (Config, error) {
	var cfg Config

	flag.IntVar(&cfg.From, "from", 0, "first movie id (required)")
	flag.IntVar(&cfg.To, "to", 0, "last movie id (required)")
	flag.IntVar(&cfg.Workers, "workers", 10, "number of workers")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "request timeout")

	flag.Parse()

	if cfg.From <= 0 {
		return cfg, errors.New("flag --from is required")
	}
	if cfg.To <= 0 {
		return cfg, errors.New("flag --to is required")
	}
	if cfg.From > cfg.To {
		return cfg, errors.New("the value --from cannot be greater than --to")
	}
	if cfg.Workers <= 0 {
		return cfg, errors.New("the number of workers must be positive")
	}
	if cfg.Timeout <= 0 {
		return cfg, errors.New("timeout must be positive")
	}

	return cfg, nil
}

func printResult(res Result) {
	if res.Error != nil {
		fmt.Fprintln(os.Stderr, "Error:", res.Error)
		return
	}
	m := res.Movie
	fmt.Printf("%d — %s — %d — %s\n", m.ID, m.Title, m.Year, m.Director)
}

func run(ctx context.Context, cfg Config) {
	numJobs := cfg.To - cfg.From + 1
	jobs := make(chan int, numJobs)
	results := make(chan Result, numJobs)
	var wg sync.WaitGroup

	for range cfg.Workers {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg, cfg.Timeout)
	}
	for j := range numJobs {
		jobs <- cfg.From + j
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		printResult(res)
	}
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	run(ctx, cfg)
}
