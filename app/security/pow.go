// Package security holds the bot-protection building blocks of ADR 0016: a
// stateless proof-of-work challenge and the honeypot check.
package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
)

// Challenge parameters (SDD 03 section 7.3).
const (
	DefaultBits = 18
	AttackBits  = 21
	TTL         = 120 * time.Second
	maxBits     = 28
	maxNonceLen = 32
)

// Verification errors; the middleware answers all but ErrStore with a fresh challenge.
var (
	ErrInvalid = errors.New("pow: invalid challenge or solution")
	ErrExpired = errors.New("pow: challenge expired")
	ErrReused  = errors.New("pow: solution already used")
	// ErrStore means the single-use store failed; callers answer 503.
	ErrStore = errors.New("pow: store unavailable")
)

// Challenge is what the client solves: find n so that SHA-256(C || n) has
// Bits leading zero bits. C is signed, bound to a route and expires at Exp.
type Challenge struct {
	C    string `json:"c"`
	Bits int    `json:"bits"`
	Exp  int64  `json:"exp"`
}

// PoW issues and verifies challenges with an HMAC key (app.DeriveKey("pow", 32)).
// A solution works once: its hash is claimed in the cache until the challenge expires.
type PoW struct {
	key   []byte
	cache func() (*cache.Cache, error)
	now   func() time.Time
}

// New builds a PoW; now is the clock (time.Now in production).
func New(key []byte, c func() (*cache.Cache, error), now func() time.Time) *PoW {
	return &PoW{key: key, cache: c, now: now}
}

func (p *PoW) mac(payload []byte, route string) []byte {
	h := hmac.New(sha256.New, p.key)
	h.Write(payload)
	h.Write([]byte{0})
	h.Write([]byte(route))

	return h.Sum(nil)
}

// Issue returns a new challenge for route with the given difficulty.
func (p *PoW) Issue(route string, difficulty int) (Challenge, error) {
	if difficulty < 1 || difficulty > maxBits {
		return Challenge{}, fmt.Errorf("pow: bits %d out of range", difficulty)
	}

	exp := p.now().Add(TTL).Unix()
	payload := make([]byte, 16+8+1)

	if _, err := rand.Read(payload[:16]); err != nil {
		return Challenge{}, err
	}

	binary.BigEndian.PutUint64(payload[16:24], uint64(exp)) //nolint:gosec // exp is a positive Unix time
	payload[24] = byte(difficulty)

	c := base64.RawURLEncoding.EncodeToString(
		payload,
	) + "." + base64.RawURLEncoding.EncodeToString(
		p.mac(payload, route),
	)

	return Challenge{C: c, Bits: difficulty, Exp: exp}, nil
}

// Verify checks that c was issued for route, has not expired, asks for at
// least minBits, and that n solves it; then it burns the solution.
func (p *PoW) Verify(ctx context.Context, route, c, n string, minBits int) error {
	if n == "" || len(n) > maxNonceLen {
		return ErrInvalid
	}

	body, sig, ok := strings.Cut(c, ".")
	if !ok {
		return ErrInvalid
	}

	payload, err := base64.RawURLEncoding.Strict().DecodeString(body)
	if err != nil || len(payload) != 25 {
		return ErrInvalid
	}

	got, err := base64.RawURLEncoding.Strict().DecodeString(sig)
	if err != nil || !hmac.Equal(got, p.mac(payload, route)) {
		return ErrInvalid
	}

	if uint64(p.now().Unix()) > binary.BigEndian.Uint64(payload[16:24]) { //nolint:gosec // Unix time is positive
		return ErrExpired
	}

	if need := int(payload[24]); need < minBits || need > maxBits || leadingZeroBits(c, n) < need {
		return ErrInvalid
	}

	store, err := p.cache()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStore, err)
	}

	sum := sha256.Sum256([]byte(c))
	fresh, err := cache.Add(ctx, store, "pow:"+base64.RawURLEncoding.EncodeToString(sum[:]), true, TTL+time.Minute)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStore, err)
	}

	if !fresh {
		return ErrReused
	}

	return nil
}

func leadingZeroBits(c, n string) int {
	sum := sha256.Sum256([]byte(c + n))
	total := 0

	for _, b := range sum {
		if b == 0 {
			total += 8
			continue
		}

		return total + bits.LeadingZeros8(b)
	}

	return total
}

// Solve finds a nonce for ch by brute force. Clients do this in a worker; Go
// callers (tests, tools) use it directly.
func Solve(ch Challenge) string {
	for i := uint64(0); ; i++ {
		n := strconv.FormatUint(i, 10)
		if leadingZeroBits(ch.C, n) >= ch.Bits {
			return n
		}
	}
}
