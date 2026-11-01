package covers_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

// TestMain turns the test binary into a fake cwebp: the Encoder under test runs it as the child process. The fake
// writes "RIFF<size>WEBP", the line "ARGS:<args>" and then the PNG it was given, so a test sees both.
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_CWEBP") == "" {
		os.Exit(m.Run())
	}

	if d := os.Getenv("FAKE_CWEBP_SLEEP"); d != "" {
		dur, _ := time.ParseDuration(d)
		time.Sleep(dur)
	}

	args := os.Args[1:]
	in, err := os.ReadFile(args[len(args)-3])
	if err != nil {
		os.Exit(2)
	}

	out := args[len(args)-1]
	body := append([]byte("ARGS:"+strings.Join(args, " ")+"\n"), in...)
	head := append([]byte("RIFF"), binary.LittleEndian.AppendUint32(nil, uint32(len(body)+4))...)

	if err := os.WriteFile(out, append(append(head, "WEBP"...), body...), 0o600); err != nil {
		os.Exit(3)
	}

	os.Exit(0)
}

func fake(t *testing.T, env ...string) *covers.Cwebp {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	return &covers.Cwebp{Path: exe, Timeout: 20 * time.Second, Env: append([]string{"FAKE_CWEBP=1"}, env...)}
}

// fed splits the fake's output into the arguments and the PNG that cwebp received.
func fed(t *testing.T, webp []byte) (string, image.Image) {
	t.Helper()

	if len(webp) < 12 || string(webp[:4]) != "RIFF" || string(webp[8:12]) != "WEBP" {
		t.Fatalf("output is not a RIFF/WEBP container: %q", webp[:min(len(webp), 12)])
	}

	line, rest, _ := bytes.Cut(webp[12:], []byte("\n"))

	img, err := png.Decode(bytes.NewReader(rest))
	if err != nil {
		t.Fatalf("cwebp input is not a PNG: %v", err)
	}

	return string(line), img
}

func gradient(w, h int, alpha uint8) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 255 / w), uint8(y * 255 / h), 90, alpha})
		}
	}

	return img
}

func jpegBytes(t *testing.T, img image.Image) []byte {
	t.Helper()

	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func pngBytes(t *testing.T, img image.Image) []byte {
	t.Helper()

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

// withExifOrientation inserts an APP1 Exif segment (big-endian TIFF, one IFD0 entry: orientation) after SOI.
func withExifOrientation(src []byte, o uint16) []byte {
	tiff := []byte{
		'M',
		'M',
		0,
		42,
		0,
		0,
		0,
		8,
		0,
		1,
		0x01,
		0x12,
		0,
		3,
		0,
		0,
		0,
		1,
		byte(o >> 8),
		byte(o),
		0,
		0,
		0,
		0,
		0,
		0,
	}
	seg := append([]byte("Exif\x00\x00"), tiff...)
	app1 := append([]byte{0xFF, 0xE1}, binary.BigEndian.AppendUint16(nil, uint16(len(seg)+2))...)

	return append(append(append([]byte{}, src[:2]...), append(app1, seg...)...), src[2:]...)
}

func TestToWebpResizesAndStripsMetadata(t *testing.T) {
	src := withExifOrientation(jpegBytes(t, gradient(1200, 1800, 255)), 1)

	out, err := fake(t).Encode(t.Context(), src, covers.Full)
	if err != nil {
		t.Fatal(err)
	}

	args, img := fed(t, out)
	if b := img.Bounds(); b.Dx() != 400 || b.Dy() != 600 {
		t.Errorf("size = %dx%d, want 400x600 (fit in 600x600)", b.Dx(), b.Dy())
	}

	for _, want := range []string{"-q 78", "-m 6", "-metadata none"} {
		if !strings.Contains(args, want) {
			t.Errorf("args %q lack %q", args, want)
		}
	}
}

func TestToWebpKeepsAlphaAndNeverUpscales(t *testing.T) {
	out, err := fake(t).Encode(t.Context(), pngBytes(t, gradient(50, 80, 128)), covers.Full)
	if err != nil {
		t.Fatal(err)
	}

	_, img := fed(t, out)
	if b := img.Bounds(); b.Dx() != 50 || b.Dy() != 80 {
		t.Errorf("size = %dx%d, want 50x80 (never upscale)", b.Dx(), b.Dy())
	}

	if _, _, _, a := img.At(10, 10).RGBA(); a == 0xFFFF {
		t.Error("alpha was flattened")
	}
}

func TestToWebpAppliesExifOrientation(t *testing.T) {
	// 200x100 with orientation 6 (rotate 90 CW): the left half is red, so after rotating it is the top half.
	src := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	for y := range 100 {
		for x := range 200 {
			c := color.NRGBA{0, 0, 255, 255}
			if x < 100 {
				c = color.NRGBA{255, 0, 0, 255}
			}

			src.SetNRGBA(x, y, c)
		}
	}

	out, err := fake(t).Encode(t.Context(), withExifOrientation(jpegBytes(t, src), 6), covers.Full)
	if err != nil {
		t.Fatal(err)
	}

	_, img := fed(t, out)
	if b := img.Bounds(); b.Dx() != 100 || b.Dy() != 200 {
		t.Fatalf("size = %dx%d, want 100x200", b.Dx(), b.Dy())
	}

	top, _, _, _ := img.At(50, 20).RGBA()
	bottom, _, _, _ := img.At(50, 180).RGBA()

	if top < 0xC000 || bottom > 0x4000 {
		t.Errorf("red channel top=%#x bottom=%#x, want red on top", top, bottom)
	}
}

func TestToWebpThumbFitsListSize(t *testing.T) {
	out, err := fake(t).Encode(t.Context(), pngBytes(t, gradient(1000, 1000, 255)), covers.Thumb)
	if err != nil {
		t.Fatal(err)
	}

	args, img := fed(t, out)
	if b := img.Bounds(); b.Dx() != 96 || b.Dy() != 96 {
		t.Errorf("size = %dx%d, want 96x96 (fit in 96x144)", b.Dx(), b.Dy())
	}

	if !strings.Contains(args, "-q 70") {
		t.Errorf("args %q lack -q 70", args)
	}
}

func TestEncodeRejectsBombsBeforeDecoding(t *testing.T) {
	// A PNG whose header declares 6000x5000 (30 MP) and nothing else: a decode would fail, DecodeConfig would not.
	hdr := pngBytes(t, gradient(2, 2, 255))[:33]
	binary.BigEndian.PutUint32(hdr[16:], 6000)
	binary.BigEndian.PutUint32(hdr[20:], 5000)
	binary.BigEndian.PutUint32(hdr[29:], crc32.ChecksumIEEE(hdr[12:29]))

	_, err := fake(t).Encode(t.Context(), hdr, covers.Full)
	if !errors.Is(err, covers.ErrTooLarge) {
		t.Errorf("30 MP error = %v, want ErrTooLarge", err)
	}

	_, err = fake(t).Encode(t.Context(), make([]byte, covers.MaxSourceBytes+1), covers.Full)
	if !errors.Is(err, covers.ErrTooLarge) {
		t.Errorf("16 MB error = %v, want ErrTooLarge", err)
	}

	_, err = fake(t).Encode(t.Context(), []byte("GIF89a not supported"), covers.Full)
	if !errors.Is(err, covers.ErrUnsupported) {
		t.Errorf("unknown format error = %v, want ErrUnsupported", err)
	}
}

func TestEncodeTimeoutKillsTheChild(t *testing.T) {
	e := fake(t, "FAKE_CWEBP_SLEEP=30s")
	e.Timeout = 300 * time.Millisecond

	start := time.Now()

	_, err := e.Encode(context.Background(), pngBytes(t, gradient(10, 10, 255)), covers.Full)
	if err == nil || time.Since(start) > 10*time.Second {
		t.Errorf("err = %v after %v, want a timeout error well before 30 s", err, time.Since(start))
	}
}

func TestMissingBinaryIsReported(t *testing.T) {
	e := &covers.Cwebp{Path: "cwebp-does-not-exist"}

	_, err := e.Encode(t.Context(), pngBytes(t, gradient(10, 10, 255)), covers.Full)
	if !errors.Is(err, covers.ErrNoEncoder) {
		t.Errorf("err = %v, want ErrNoEncoder", err)
	}
}

// TestWebpSizeParity compares the Encoder output with the sizes the pinned cwebp produces on its own for the same
// fixtures (testdata/covers/reference_sizes.json): ±10 %. It needs the real cwebp, so it skips where there is none;
// run it in the app image (see the P6.B.2.1 report).
func TestWebpSizeParity(t *testing.T) {
	path, err := exec.LookPath("cwebp")
	if err != nil {
		t.Skip("cwebp is not installed; run in the app image")
	}

	raw, err := os.ReadFile("../testdata/covers/reference_sizes.json")
	if err != nil {
		t.Fatal(err)
	}

	var ref struct {
		Sizes map[string]map[string]int `json:"sizes"`
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}

	enc := &covers.Cwebp{Path: path, Env: []string{}}
	variants := map[string]covers.Variant{"full": covers.Full, "thumb": covers.Thumb}

	for name, bySize := range ref.Sizes {
		src, err := os.ReadFile("../testdata/covers/" + name)
		if err != nil {
			t.Fatal(err)
		}

		for tag, want := range bySize {
			out, err := enc.Encode(t.Context(), src, variants[tag])
			if err != nil {
				t.Fatalf("%s %s: %v", name, tag, err)
			}

			if d := math.Abs(float64(len(out))-float64(want)) / float64(want); d > 0.10 {
				t.Errorf("%s %s: %d bytes, reference %d (%.1f %% off, limit 10 %%)", name, tag, len(out), want, d*100)
			}
		}
	}
}
