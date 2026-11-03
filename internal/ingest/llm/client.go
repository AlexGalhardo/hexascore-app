// Package llm is the last-resort text normaliser of the catalogue worker (RFC-001 section 5 F). It reads text the
// worker already fetched legally, asks Claude Haiku for a few facts, validates the answer against a strict schema and
// queues every value for a person to review: a model never writes to items. The whole path is off until FLAG_INGEST_LLM
// is on and stops for the month once LLM_MONTHLY_BUDGET_USD is spent.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

// The API and the model (OQ-05, 2026-11-02: Claude Haiku 5.5).
const (
	APIURL = "https://api.anthropic.com/v1/messages"
	Host   = "api.anthropic.com"
	Model  = "claude-haiku-5-5"

	apiVersion      = "2023-06-01"
	maxOutputTokens = 300
	// maxResponse caps what is read back: the answer is a few dozen tokens.
	maxResponse = 1 << 20
)

// HostLimits is the allowlist and rate table of the client the step uses.
func HostLimits() map[string]httpx.HostLimit {
	return map[string]httpx.HostLimit{Host: {PerSecond: 5, InFlight: 2}}
}

// Client calls the Messages API. APIKey comes from ANTHROPIC_API_KEY and is never logged. HTTP is the ingestion
// client (identified User-Agent, throttle, breaker, SSRF-safe dialer).
type Client struct {
	APIKey string
	// APIURL overrides APIURL; tests only.
	APIURL string
	HTTP   *http.Client
}

// Completion is one answer with the tokens it used.
type Completion struct {
	Text                string
	TokensIn, TokensOut int
}

// IsPermanent reports whether retrying cannot help (no key, a rejected key, a malformed request).
func IsPermanent(err error) bool { return errors.Is(err, queue.ErrPermanent) }

// Complete sends one user message with a system prompt and returns the first text block.
func (c *Client) Complete(ctx context.Context, system, user string) (Completion, error) {
	if c.APIKey == "" {
		return Completion{}, fmt.Errorf("llm: ANTHROPIC_API_KEY is not set: %w", queue.ErrPermanent)
	}

	body, err := json.Marshal(map[string]any{
		"model": Model, "max_tokens": maxOutputTokens, "temperature": 0, "system": system,
		"messages": []map[string]string{{"role": "user", "content": user}},
	})
	if err != nil {
		return Completion{}, err
	}

	url := c.APIURL
	if url == "" {
		url = APIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Completion{}, fmt.Errorf("llm: build request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Completion{}, fmt.Errorf("llm: request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return Completion{}, fmt.Errorf("llm: read response: %w", err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return Completion{}, fmt.Errorf(
			"llm: the API rejected the key (status %d): %w",
			resp.StatusCode,
			queue.ErrPermanent,
		)
	case resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests:
		return Completion{}, fmt.Errorf(
			"llm: the API refused the request (status %d): %w",
			resp.StatusCode,
			queue.ErrPermanent,
		)
	default:
		return Completion{}, fmt.Errorf("llm: status %d", resp.StatusCode)
	}

	var doc struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			In  int `json:"input_tokens"`
			Out int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(raw, &doc); err != nil {
		return Completion{}, fmt.Errorf("llm: decode response: %w", err)
	}

	for _, block := range doc.Content {
		if block.Type == "text" && block.Text != "" {
			return Completion{Text: block.Text, TokensIn: doc.Usage.In, TokensOut: doc.Usage.Out}, nil
		}
	}

	// the tokens were spent even though nothing usable came back; the caller still records them
	return Completion{TokensIn: doc.Usage.In, TokensOut: doc.Usage.Out}, errors.New("llm: the answer has no text")
}
