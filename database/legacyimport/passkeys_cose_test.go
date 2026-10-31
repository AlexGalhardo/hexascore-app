package legacyimport

import "testing"

func TestCOSEAlgES256(t *testing.T) {
	t.Parallel()

	// Map of one pair: key 3, value -7 (ES256).
	alg, err := coseAlg([]byte{0xa1, 0x03, 0x26})
	if err != nil {
		t.Fatal(err)
	}

	if alg != -7 {
		t.Fatalf("alg = %d, want -7", alg)
	}

	// A longer definite map still yields key 3 and rejects trailing bytes.
	// {1: 2, 3: -7}
	alg, err = coseAlg([]byte{0xa2, 0x01, 0x02, 0x03, 0x26})
	if err != nil {
		t.Fatal(err)
	}

	if alg != -7 {
		t.Fatalf("alg = %d, want -7", alg)
	}

	if _, err := coseAlg([]byte{0xa1, 0x03, 0x26, 0x00}); err == nil {
		t.Fatal("leftover bytes accepted")
	}

	if _, err := coseAlg([]byte{0xbf}); err == nil {
		t.Fatal("indefinite map accepted")
	}

	if _, err := coseAlg([]byte{0xa1, 0x01, 0x02}); err == nil {
		t.Fatal("map without key 3 accepted")
	}
}
