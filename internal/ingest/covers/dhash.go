package covers

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math/bits"
)

// NearDuplicate is the largest Hamming distance between two dHashes still counted as the same picture.
const NearDuplicate = 4

// DHash is the 64-bit difference hash of img: the picture is reduced to 9x8 grey cells and each bit says whether a
// cell is brighter than its right neighbour. It survives rescaling and recompression, which is what two sources
// serving the same poster differ by.
func DHash(img image.Image) uint64 {
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return 0
	}

	var grey [8][9]uint64

	for y := range 8 {
		y0, y1 := b.Min.Y+y*b.Dy()/8, b.Min.Y+max(y*b.Dy()/8+1, (y+1)*b.Dy()/8)

		for x := range 9 {
			x0, x1 := b.Min.X+x*b.Dx()/9, b.Min.X+max(x*b.Dx()/9+1, (x+1)*b.Dx()/9)

			var sum, n uint64

			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					r, g, bl, _ := img.At(sx, sy).RGBA()
					sum += (299*uint64(r) + 587*uint64(g) + 114*uint64(bl)) / 1000
					n++
				}
			}

			grey[y][x] = sum / n
		}
	}

	var h uint64

	for y := range 8 {
		for x := range 8 {
			h <<= 1

			if grey[y][x] > grey[y][x+1] {
				h |= 1
			}
		}
	}

	return h
}

// Distance is the Hamming distance of two dHashes.
func Distance(a, b uint64) int { return bits.OnesCount64(a ^ b) }

// Fingerprint decodes a JPEG or PNG source and returns its dHash and size.
func Fingerprint(src []byte) (hash uint64, w, h int, err error) {
	var img image.Image

	_, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: %w", ErrUnsupported, err)
	}

	switch format {
	case "jpeg":
		img, err = jpeg.Decode(bytes.NewReader(src))
	case "png":
		img, err = png.Decode(bytes.NewReader(src))
	default:
		return 0, 0, 0, fmt.Errorf("%w: %s", ErrUnsupported, format)
	}

	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: %w", ErrUnsupported, err)
	}

	b := img.Bounds()

	return DHash(img), b.Dx(), b.Dy(), nil
}
