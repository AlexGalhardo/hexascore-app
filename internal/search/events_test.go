package search_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

func TestRedact(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"call ①②③④⑤⑥⑦":                          "call [redacted]",
		"zelda":                                 "zelda",
		"mail me ana@example.com please":        "mail me [redacted] please",
		"123.456.789-09":                        "[redacted]",
		"12345678909 zelda":                     "[redacted] zelda",
		"call +55 (11) 91234-5678":              "call [redacted]",
		"blade runner 2049":                     "blade runner 2049",
		"fahrenheit 451":                        "fahrenheit 451",
		"ISBN 9780306406157":                    "ISBN [redacted]",
		"ana@example.com and 555-123-4567 here": "[redacted] and [redacted] here",
		"phone 11_9123_4567":                    "phone [redacted]",
		"cpf 123/456/789/09":                    "cpf [redacted]",
		"１１９１２３４５６７":                            "[redacted]",
		"call ①②③④⑤⑥⑦":                          "call [redacted]",
		"top 100 games of 2024":                 "top 100 games of 2024",
	}

	for in, want := range tests {
		if got := search.Redact(in); got != want {
			t.Errorf("Redact(%q) = %q, want %q", in, got, want)
		}
	}
}
