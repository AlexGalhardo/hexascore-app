package security_test

import (
	"errors"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"

	"github.com/AlexGalhardo/hexascore-app/app/security"
)

func newPoW(t *testing.T, now *time.Time) *security.PoW {
	t.Helper()

	c := cache.New(cache.NewMemory(1000), "test:")

	return security.New([]byte("0123456789abcdef0123456789abcdef"), func() (*cache.Cache, error) { return c, nil },
		func() time.Time { return *now })
}

func TestPoW(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	p := newPoW(t, &now)
	ctx := t.Context()

	ch, err := p.Issue("signup", 12)
	if err != nil {
		t.Fatal(err)
	}

	n := security.Solve(ch)

	if err := p.Verify(ctx, "login", ch.C, n, 12); !errors.Is(err, security.ErrInvalid) {
		t.Errorf("other route: %v, want ErrInvalid", err)
	}

	if err := p.Verify(ctx, "signup", ch.C, n, 13); !errors.Is(err, security.ErrInvalid) {
		t.Errorf("below the required difficulty: %v, want ErrInvalid", err)
	}

	if err := p.Verify(ctx, "signup", ch.C+"x", n, 12); !errors.Is(err, security.ErrInvalid) {
		t.Errorf("tampered challenge: %v, want ErrInvalid", err)
	}

	for _, bad := range []string{"", "x", "99999999999999999999999999999999999"} {
		if err := p.Verify(ctx, "signup", ch.C, bad, 12); !errors.Is(err, security.ErrInvalid) {
			t.Errorf("nonce %q: %v, want ErrInvalid", bad, err)
		}
	}

	if err := p.Verify(ctx, "signup", ch.C, n, 12); err != nil {
		t.Fatalf("valid solution: %v", err)
	}

	if err := p.Verify(ctx, "signup", ch.C, n, 12); !errors.Is(err, security.ErrReused) {
		t.Errorf("reuse: %v, want ErrReused", err)
	}

	late, _ := p.Issue("signup", 12)
	now = now.Add(security.TTL + time.Second)

	if err := p.Verify(ctx, "signup", late.C, security.Solve(late), 12); !errors.Is(err, security.ErrExpired) {
		t.Errorf("expired: %v, want ErrExpired", err)
	}
}
