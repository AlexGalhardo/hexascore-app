// Package llm is the last-resort gap filler of the catalogue worker (RFC-001 section 5 F, revised 2026-11-04). It is
// called only for items that still lack a field after every other source: it reads text the worker already fetched
// (the Wikipedia introduction, the English synopsis), asks Claude Haiku for the missing facts or translations,
// validates the answer against a strict schema and writes a value directly only when it is grounded (the number is in
// the source text, or the text is a translation of our own synopsis); everything else goes to the review queue.
// With LLM_WEB_SEARCH on, an item with no stored text may be looked up with the API's web search, and those answers
// always go to review. The whole path is off until FLAG_INGEST_LLM is on and stops for the month once
// LLM_MONTHLY_BUDGET_USD is spent.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

// The API and the model (OQ-05, 2026-11-02: Claude Haiku 5.5).
const (
	APIURL = "https://api.anthropic.com/v1/messages"
	Host   = "api.anthropic.com"
	Model  = "claude-haiku-5-5"

	apiVersion      = "2023-06-01"
	maxOutputTokens = 1024
	// WebSearchTool is the server tool version Claude Haiku 5.5 takes (the dynamic-filtering 20260209 version is for
	// the Opus and Sonnet lines).
	WebSearchTool = "web_search_20250305"
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

// Completion is one answer with what it used.
type Completion struct {
	Text                string
	TokensIn, TokensOut int
	// Searches is how many web searches the API ran (billed per search).
	Searches int
	// Sources are the URLs the answer cited, from web search results.
	Sources []string
}

// Request is one call. MaxTokens 0 takes the default; WebSearches > 0 lets the model search the web that many times.
type Request struct {
	System, User string
	MaxTokens    int
	WebSearches  int
}

// IsPermanent reports whether retrying cannot help (no key, a rejected key, a malformed request).
func IsPermanent(err error) bool { return errors.Is(err, queue.ErrPermanent) }

// Complete sends one user message with a system prompt and returns the answer.
func (c *Client) Complete(ctx context.Context, system, user string) (Completion, error) {
	return c.Do(ctx, Request{System: system, User: user})
}

// Do sends one request. The model runs with thinking off and low effort: the tasks are short extractions and
// translations, and Claude Haiku 5.5 rejects sampling parameters such as temperature (a 400 on every call before).
func (c *Client) Do(ctx context.Context, r Request) (Completion, error) {
	if c.APIKey == "" {
		return Completion{}, fmt.Errorf("llm: ANTHROPIC_API_KEY is not set: %w", queue.ErrPermanent)
	}

	maxTokens := r.MaxTokens
	if maxTokens <= 0 {
		maxTokens = maxOutputTokens
	}

	payload := map[string]any{
		"model": Model, "max_tokens": maxTokens, "system": r.System,
		"thinking":      map[string]string{"type": "disabled"},
		"output_config": map[string]string{"effort": "low"},
		"messages":      []map[string]string{{"role": "user", "content": r.User}},
	}

	if r.WebSearches > 0 {
		payload["tools"] = []map[string]any{{"type": WebSearchTool, "name": "web_search", "max_uses": r.WebSearches}}
	}

	body, err := json.Marshal(payload)
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

	return parseResponse(raw)
}

// parseResponse reads an answer: the last text block is the reply (with web search, earlier text blocks narrate the
// searches), the cited URLs are kept, and the usage counts tokens and searches.
func parseResponse(raw []byte) (Completion, error) {
	var doc struct {
		StopReason string `json:"stop_reason"`
		Content    []struct {
			Type      string `json:"type"`
			Text      string `json:"text"`
			Citations []struct {
				URL string `json:"url"`
			} `json:"citations"`
		} `json:"content"`
		Usage struct {
			In            int `json:"input_tokens"`
			Out           int `json:"output_tokens"`
			ServerToolUse struct {
				WebSearch int `json:"web_search_requests"`
			} `json:"server_tool_use"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(raw, &doc); err != nil {
		return Completion{}, fmt.Errorf("llm: decode response: %w", err)
	}

	out := Completion{TokensIn: doc.Usage.In, TokensOut: doc.Usage.Out, Searches: doc.Usage.ServerToolUse.WebSearch}

	for _, block := range doc.Content {
		if block.Type != "text" {
			continue
		}

		if block.Text != "" {
			out.Text = block.Text
		}

		for _, c := range block.Citations {
			if c.URL != "" && !slices.Contains(out.Sources, c.URL) {
				out.Sources = append(out.Sources, c.URL)
			}
		}
	}

	switch {
	case doc.StopReason == "refusal":
		return out, errors.New("llm: the model declined the request")
	case out.Text == "":
		// the tokens were spent even though nothing usable came back; the caller still records them
		return out, errors.New("llm: the answer has no text")
	}

	return out, nil
}
