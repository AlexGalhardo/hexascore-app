package covers

import (
	"encoding/binary"
	"image"
)

// exifOrientation reads the Orientation tag (1..8) from a JPEG's APP1 Exif segment; anything unreadable is 1.
func exifOrientation(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}

	for i := 2; i+4 <= len(b) && b[i] == 0xFF; {
		marker, n := b[i+1], int(binary.BigEndian.Uint16(b[i+2:]))
		if n < 2 || i+2+n > len(b) || marker == 0xDA {
			return 1
		}

		if seg := b[i+4 : i+2+n]; marker == 0xE1 && len(seg) > 14 && string(seg[:6]) == "Exif\x00\x00" {
			return tiffOrientation(seg[6:])
		}

		i += 2 + n
	}

	return 1
}

func tiffOrientation(t []byte) int {
	var bo binary.ByteOrder

	switch string(t[:2]) {
	case "MM":
		bo = binary.BigEndian
	case "II":
		bo = binary.LittleEndian
	default:
		return 1
	}

	off := int(bo.Uint32(t[4:]))
	if off < 8 || off+2 > len(t) {
		return 1
	}

	for n, p := int(bo.Uint16(t[off:])), off+2; n > 0 && p+12 <= len(t); n, p = n-1, p+12 {
		if bo.Uint16(t[p:]) == 0x0112 {
			if o := int(bo.Uint16(t[p+8:])); o >= 1 && o <= 8 {
				return o
			}

			return 1
		}
	}

	return 1
}

// orient returns src turned upright for EXIF orientation o (the 8 TIFF cases), as an NRGBA image.
func orient(src image.Image, o int) *image.NRGBA {
	sb := src.Bounds()
	w, h := sb.Dx(), sb.Dy()

	dw, dh := w, h
	if o >= 5 {
		dw, dh = h, w
	}

	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))

	for y := range h {
		for x := range w {
			var dx, dy int

			switch o {
			case 2:
				dx, dy = w-1-x, y
			case 3:
				dx, dy = w-1-x, h-1-y
			case 4:
				dx, dy = x, h-1-y
			case 5:
				dx, dy = y, x
			case 6:
				dx, dy = h-1-y, x
			case 7:
				dx, dy = h-1-y, w-1-x
			case 8:
				dx, dy = y, w-1-x
			default:
				dx, dy = x, y
			}

			dst.Set(dx, dy, src.At(sb.Min.X+x, sb.Min.Y+y))
		}
	}

	return dst
}
