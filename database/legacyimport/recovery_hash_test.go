package legacyimport

import (
	"testing"

	"github.com/AlexGalhardo/vukem/auth/totp"
)

func TestRecoveryHashesMatchVukemFormat(t *testing.T) {
	codes := []string{"ab12-cd34", "  Ef56-Gh78 "}
	hashes := recoveryHashes(codes)

	for _, c := range codes {
		left, ok := totp.UseRecoveryCode(c, hashes)
		if !ok || len(left) != len(hashes)-1 {
			t.Fatalf("code %q not accepted by totp.UseRecoveryCode", c)
		}
	}

	if got := recoveryHashes([]string{" ab12-cd34"})[0]; got != hashes[0] {
		t.Fatalf("trim/upper not normalised: %s != %s", got, hashes[0])
	}
}
