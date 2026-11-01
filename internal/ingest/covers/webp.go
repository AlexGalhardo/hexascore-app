// Package covers prepares cover images: it fits them into the two stored sizes and encodes WebP (SDD 06 §7.5).
package covers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// MaxSourceBytes is the largest source image accepted (SDD 06 §7.5).
	MaxSourceBytes = 15 << 20
	// MaxPixels is the largest decoded image accepted: 25 MP.
	MaxPixels = 25_000_000
	// DefaultTimeout bounds one cwebp run.
	DefaultTimeout = 30 * time.Second
)

var (
	// ErrTooLarge means the source is over 15 MB or 25 MP; it is rejected before it is decoded.
	ErrTooLarge = errors.New("covers: source image too large")
	// ErrUnsupported means the source is neither JPEG nor PNG.
	ErrUnsupported = errors.New("covers: unsupported image format")
	// ErrNoEncoder means the cwebp binary is not installed (VF-004: the caller keeps the JPEG and says so).
	ErrNoEncoder = errors.New("covers: cwebp not found")
)

// Variant is one stored size: the source is fitted into MaxW x MaxH, never upscaled.
type Variant struct {
	MaxW, MaxH, Quality int
}

var (
	// Full is the detail-page cover: 600x600 at q78.
	Full = Variant{600, 600, 78}
	// Thumb is the list cover: 96x144 at q70.
	Thumb = Variant{96, 144, 70}
)

// Encoder turns a source JPEG or PNG into the WebP of a variant, without metadata.
type Encoder interface {
	Encode(ctx context.Context, src []byte, v Variant) ([]byte, error)
}

// encodeSlot lets one cwebp run at a time (NFR-06-2).
var encodeSlot = make(chan struct{}, 1)

// Cwebp is the Encoder that runs the pinned cwebp binary (Alpine libwebp-tools) as a child process.
//
// shortcut: the child gets a 30 s timeout, a scratch directory and an empty environment, but no address-space or
// CPU limit. contrib/engine (Limits{CPUTime, AddressSpace}) does that and is a separate module the app does not
// require yet; swap run for engine.Tool.Run once the owner approves the dependency.
type Cwebp struct {
	// Path is the binary; empty means "cwebp" on PATH.
	Path string
	// Timeout bounds one run; zero means DefaultTimeout.
	Timeout time.Duration
	// Env is the whole environment of the child (tests); nil means empty.
	Env []string
}

// Encode fits src into v and returns the WebP bytes.
func (c *Cwebp) Encode(ctx context.Context, src []byte, v Variant) ([]byte, error) {
	if len(src) > MaxSourceBytes {
		return nil, fmt.Errorf("%w: %d bytes", ErrTooLarge, len(src))
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnsupported, err)
	}

	if format != "jpeg" && format != "png" {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, format)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width*cfg.Height > MaxPixels {
		return nil, fmt.Errorf("%w: %dx%d", ErrTooLarge, cfg.Width, cfg.Height)
	}

	bin := c.Path
	if bin == "" {
		bin = "cwebp"
	}

	path, err := exec.LookPath(bin)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoEncoder, err)
	}

	img, err := fit(src, format, v)
	if err != nil {
		return nil, err
	}

	select {
	case encodeSlot <- struct{}{}:
		defer func() { <-encodeSlot }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return c.run(ctx, path, img, v)
}

func (c *Cwebp) run(ctx context.Context, bin string, img image.Image, v Variant) ([]byte, error) {
	dir, err := os.MkdirTemp("", "cwebp-")
	if err != nil {
		return nil, fmt.Errorf("covers: scratch dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	var in bytes.Buffer
	if err := png.Encode(&in, img); err != nil {
		return nil, fmt.Errorf("covers: encode intermediate png: %w", err)
	}

	inPath, outPath := filepath.Join(dir, "in.png"), filepath.Join(dir, "out.webp")
	if err := os.WriteFile(inPath, in.Bytes(), 0o600); err != nil {
		return nil, fmt.Errorf("covers: write intermediate png: %w", err)
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// bin is operator configuration and every argument is built here, none from the image.
	cmd := exec.CommandContext( //nolint:gosec // see above
		ctx, bin, "-q", strconv.Itoa(v.Quality), "-m", "6", "-metadata", "none", inPath, "-o", outPath,
	)
	cmd.Dir = dir
	cmd.Env = append([]string{}, c.Env...)

	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("covers: cwebp: %w: %.200s", err, out)
	}

	return os.ReadFile(outPath)
}

// fit decodes src, applies its EXIF orientation and scales it down into v with a box filter.
func fit(src []byte, format string, v Variant) (image.Image, error) {
	var (
		img image.Image
		err error
	)

	if format == "jpeg" {
		img, err = jpeg.Decode(bytes.NewReader(src))
	} else {
		img, err = png.Decode(bytes.NewReader(src))
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnsupported, err)
	}

	o := 1
	if format == "jpeg" {
		o = exifOrientation(src)
	}

	up := orient(img, o)
	b := up.Bounds()

	if b.Dx() <= v.MaxW && b.Dy() <= v.MaxH {
		return up, nil
	}

	scale := min(float64(v.MaxW)/float64(b.Dx()), float64(v.MaxH)/float64(b.Dy()))
	w, h := max(1, int(float64(b.Dx())*scale)), max(1, int(float64(b.Dy())*scale))

	return boxScale(up, w, h), nil
}

// boxScale averages the source pixels under each destination pixel, with alpha-weighted colour.
func boxScale(src *image.NRGBA, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()

	for y := range h {
		y0, y1 := y*sh/h, max(y*sh/h+1, (y+1)*sh/h)

		for x := range w {
			x0, x1 := x*sw/w, max(x*sw/w+1, (x+1)*sw/w)

			var r, g, bl, a, n uint64

			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					p := src.NRGBAAt(sx, sy)
					al := uint64(p.A)
					r, g, bl, a, n = r+uint64(p.R)*al, g+uint64(p.G)*al, bl+uint64(p.B)*al, a+al, n+1
				}
			}

			if a == 0 {
				continue
			}

			dst.SetNRGBA(x, y, color.NRGBA{avg8(r, a), avg8(g, a), avg8(bl, a), avg8(a, n)})
		}
	}

	return dst
}

// avg8 is sum/div for sums of bytes, so the result always fits a byte.
func avg8(sum, div uint64) uint8 { return uint8(min(sum/div, 255)) }
