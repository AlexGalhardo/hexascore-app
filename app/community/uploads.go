package community

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/httpclient"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/upload"
	"github.com/AlexGalhardo/vukem/web"
)

const (
	minCoverSide = 300
	coverDir     = "community/covers"
)

// Cover is a stored community cover.
type Cover struct {
	Key    string `json:"key"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// CoverRules are JPEG, PNG and WebP up to 5 MiB under community/covers/.
// upload.Images also allows GIF; this set does not. AVIF is undetectable
// by that package, so it is refused with the other unknown types.
func CoverRules() upload.Rules {
	rules := upload.Images(5 << 20)
	rules.Types = []string{"image/jpeg", "image/png", "image/webp"}
	rules.Dir = coverDir
	rules.MaxFiles = 1

	return rules
}

// UploadCover reads one multipart file named cover. AVIF is 422, a body
// over 5 MiB is 413, and a short side under 300 px is 422.
func UploadCover(w http.ResponseWriter, r *http.Request, disk storage.Disk) {
	c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))
	form, err := upload.Receive(c, disk, CoverRules())
	if err != nil {
		failUpload(w, err)
		return
	}

	if len(form.Files) != 1 {
		failUpload(w, errs.NewValidation().Add("cover", "One cover is required."))
		return
	}

	cover, err := measureCover(r.Context(), disk, form.Files[0])
	if err != nil {
		failUpload(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cover)
}

// SaveCover stores one already-open image with the cover rules.
func SaveCover(ctx context.Context, disk storage.Disk, filename string, src io.Reader) (Cover, error) {
	file, err := upload.Save(ctx, disk, CoverRules(), "cover", filename, src)
	if err != nil {
		return Cover{}, err
	}

	return measureCover(ctx, disk, file)
}

// FetchCover downloads an https cover through the safe client and stores it.
// A blocked or refused URL returns ErrCoverBlocked or ErrCoverRefused.
func FetchCover(ctx context.Context, disk storage.Disk, raw string) (Cover, error) {
	if err := allowedCover(raw); err != nil || !strings.HasPrefix(raw, "https://") {
		return Cover{}, ErrCoverRefused
	}

	client := httpclient.Safe(httpclient.Options{
		Timeout:      5 * time.Second,
		MaxRedirects: 3,
		MaxBody:      (5 << 20) + 1,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, http.NoBody)
	if err != nil {
		return Cover{}, ErrCoverRefused
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, httpclient.ErrBlockedAddress) {
			return Cover{}, ErrCoverBlocked
		}

		return Cover{}, fmt.Errorf("community: fetch cover: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Cover{}, ErrCoverRefused
	}

	name := "cover.jpg"
	switch {
	case strings.Contains(resp.Header.Get("Content-Type"), "png"):
		name = "cover.png"
	case strings.Contains(resp.Header.Get("Content-Type"), "webp"):
		name = "cover.webp"
	}

	return SaveCover(ctx, disk, name, resp.Body)
}

// RememberCover writes the fetch outcome into suggestion meta.
func RememberCover(ctx context.Context, db *database.DB, suggestionID, status string, cover Cover) error {
	payload := map[string]any{"cover_status": status}
	if cover.Key != "" {
		payload["cover_key"] = cover.Key
		payload["cover_width"] = cover.Width
		payload["cover_height"] = cover.Height
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("community: encode cover status: %w", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE suggestions SET meta = meta || $2::jsonb, updated_at = now() WHERE id = $1`,
		suggestionID,
		body,
	); err != nil {
		return fmt.Errorf("community: store cover status: %w", err)
	}

	return nil
}

func measureCover(ctx context.Context, disk storage.Disk, file upload.File) (Cover, error) {
	width, height, err := imageSize(ctx, disk, file.Path, file.Type)
	if err != nil || width < minCoverSide || height < minCoverSide {
		if delErr := disk.Delete(ctx, file.Path); delErr != nil && !errors.Is(delErr, storage.ErrNotFound) {
			return Cover{}, fmt.Errorf("community: delete small cover: %w", delErr)
		}

		if err != nil {
			return Cover{}, errs.NewValidation().Add("cover", "The image could not be read.")
		}

		return Cover{}, errs.NewValidation().Add("cover", "The short side must be at least 300 px.")
	}

	return Cover{Key: file.Path, Width: width, Height: height}, nil
}

func failUpload(w http.ResponseWriter, err error) {
	status := errs.Status(err)
	if status < 400 || status > 599 {
		status = http.StatusInternalServerError
	}

	w.WriteHeader(status)
}

func imageSize(ctx context.Context, disk storage.Disk, path, typ string) (int, int, error) {
	rc, err := disk.Open(ctx, path)
	if err != nil {
		return 0, 0, err
	}
	defer rc.Close()

	var cfg image.Config
	switch typ {
	case "image/jpeg":
		cfg, err = jpeg.DecodeConfig(rc)
	case "image/png":
		cfg, err = png.DecodeConfig(rc)
	case "image/webp":
		cfg, err = webpSize(rc)
	default:
		return 0, 0, errors.New("community: unsupported cover type")
	}
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, errors.New("community: unreadable cover")
	}

	return cfg.Width, cfg.Height, nil
}

// webpSize reads a VP8, VP8L or VP8X canvas. The standard library has no WebP decoder.
func webpSize(r io.Reader) (image.Config, error) {
	var b [30]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return image.Config{}, err
	}

	if string(b[0:4]) != "RIFF" || string(b[8:12]) != "WEBP" {
		return image.Config{}, errors.New("community: not a webp")
	}

	var w, h int
	switch string(b[12:16]) {
	case "VP8 ":
		if b[23] != 0x9d || b[24] != 0x01 || b[25] != 0x2a {
			return image.Config{}, errors.New("community: bad webp")
		}
		w = int(binary.LittleEndian.Uint16(b[26:]) & 0x3fff)
		h = int(binary.LittleEndian.Uint16(b[28:]) & 0x3fff)
	case "VP8L":
		if b[20] != 0x2f {
			return image.Config{}, errors.New("community: bad webp")
		}
		bits := binary.LittleEndian.Uint32(b[21:])
		w = int(bits&0x3fff) + 1
		h = int(bits>>14&0x3fff) + 1
	case "VP8X":
		w = (int(b[24]) | int(b[25])<<8 | int(b[26])<<16) + 1
		h = (int(b[27]) | int(b[28])<<8 | int(b[29])<<16) + 1
	default:
		return image.Config{}, errors.New("community: bad webp")
	}

	return image.Config{Width: w, Height: h}, nil
}
