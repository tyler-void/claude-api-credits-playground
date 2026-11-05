// Command hello-credits makes one Messages API call with a key read only from
// ARENA_ANTHROPIC_API_KEY and prints the reply, the token usage and the
// estimated cost from prices.yaml.
//
//	ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key)" go run . "Say hi in five words."
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"gopkg.in/yaml.v3"
)

// KeyEnv is the only place the key is read from.
const KeyEnv = "ARENA_ANTHROPIC_API_KEY"

// Price is USD per million tokens.
type Price struct {
	Input      float64 `yaml:"input"`
	Output     float64 `yaml:"output"`
	CacheRead  float64 `yaml:"cache_read"`
	CacheWrite float64 `yaml:"cache_write"`
}

// Cost estimates the USD of one request from its reported usage.
func (p Price) Cost(u anthropic.Usage) float64 {
	return (float64(u.InputTokens)*p.Input + float64(u.OutputTokens)*p.Output +
		float64(u.CacheReadInputTokens)*p.CacheRead + float64(u.CacheCreationInputTokens)*p.CacheWrite) / 1e6
}

// newClient uses only the explicit key: no environment variables, no
// profiles, no SDK retries. baseURL "" means the production API.
func newClient(key, baseURL string) anthropic.Client {
	opts := []option.RequestOption{
		option.WithoutEnvironmentDefaults(),
		option.WithHTTPClient(&http.Client{Timeout: 5 * time.Minute}),
		option.WithAPIKey(key),
		option.WithMaxRetries(0),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return anthropic.NewClient(opts...)
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, ""); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, out io.Writer, baseURL string) error {
	fs := flag.NewFlagSet("hello-credits", flag.ContinueOnError)
	model := fs.String("model", "claude-haiku-5-5", "model id")
	effort := fs.String("effort", "", "low|medium|high|xhigh|max (empty = model default)")
	maxTokens := fs.Int64("max-tokens", 4000, "max output tokens, including thinking")
	pricesPath := fs.String("prices", "prices.yaml", "price table (USD per million tokens)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	prompt := strings.Join(fs.Args(), " ")
	if prompt == "" {
		return errors.New("usage: hello-credits [flags] <prompt>")
	}
	key := getenv(KeyEnv)
	if key == "" {
		return fmt.Errorf("%s is not set (see docs/setup-keys.md)", KeyEnv)
	}
	raw, err := os.ReadFile(*pricesPath)
	if err != nil {
		return err
	}
	var prices map[string]Price
	if err := yaml.Unmarshal(raw, &prices); err != nil {
		return fmt.Errorf("%s: %w", *pricesPath, err)
	}
	price, ok := prices[*model]
	if !ok {
		return fmt.Errorf("no price for %s in %s", *model, *pricesPath)
	}

	client := newClient(key, baseURL)
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(*model),
		MaxTokens: *maxTokens,
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	}
	if *effort != "" {
		params.OutputConfig = anthropic.OutputConfigParam{Effort: anthropic.OutputConfigEffort(*effort)}
	}
	start := time.Now()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		var apierr *anthropic.Error
		if errors.As(err, &apierr) {
			return fmt.Errorf("API returned %d (request id %s)", apierr.StatusCode, apierr.RequestID)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Fprintln(out, t.Text)
		}
	}
	if string(resp.Model) != *model {
		fmt.Fprintf(out, "warning: answered by %s, asked for %s\n", resp.Model, *model)
	}
	u := resp.Usage
	fmt.Fprintf(out, "\nmodel %s  stop %s  latency %.1fs\n", resp.Model, resp.StopReason, time.Since(start).Seconds())
	fmt.Fprintf(out, "tokens: input %d, output %d, cache read %d, cache write %d\n",
		u.InputTokens, u.OutputTokens, u.CacheReadInputTokens, u.CacheCreationInputTokens)
	fmt.Fprintf(out, "estimated cost: $%.6f\n", price.Cost(u))
	return nil
}
