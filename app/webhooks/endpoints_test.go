package webhooks

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"
)

func TestPublicURL(t *testing.T) {
	ctx := context.Background()
	public := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	}
	private := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.8")}, nil
	}

	if err := PublicURL(ctx, "https://1.1.1.1/hook", public); err != nil {
		t.Fatal(err)
	}

	if err := PublicURL(ctx, "https://hooks.example/hook", public); err != nil {
		t.Fatal(err)
	}

	refused := []string{
		"http://1.1.1.1/hook",
		"https://127.0.0.1/hook",
		"https://10.0.0.5/hook",
		"https://169.254.169.254/hook",
		"https://user:pw@1.1.1.1/hook",
		"https://hooks.example/hook",
		"",
	}
	for _, raw := range refused {
		resolve := public
		if raw == "https://hooks.example/hook" {
			resolve = private
		}

		if err := PublicURL(ctx, raw, resolve); !errorsIsURL(err) {
			t.Fatalf("PublicURL(%q) = %v, want ErrURL", raw, err)
		}
	}
}

func TestPublicURLLookupTimeout(t *testing.T) {
	resolve := func(ctx context.Context, _, _ string) ([]netip.Addr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	start := time.Now()
	err := PublicURL(context.Background(), "https://hooks.example/hook", resolve)
	if !errorsIsURL(err) {
		t.Fatalf("PublicURL = %v, want ErrURL", err)
	}

	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("lookup took %s, want it bounded by lookupTimeout", elapsed)
	}
}

func errorsIsURL(err error) bool {
	return errors.Is(err, ErrURL)
}
