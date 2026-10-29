package i18n_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func TestCategoryKeys(t *testing.T) {
	t.Parallel()

	for _, locale := range []string{"en", "pt-BR", "es"} {
		raw, err := os.ReadFile("../../resources/lang/" + locale + ".json")
		if err != nil {
			t.Fatal(err)
		}

		var cat struct {
			Categories map[string]map[string]string `json:"categories"`
			Rating     map[string]string            `json:"rating"`
		}
		if err := json.Unmarshal(raw, &cat); err != nil {
			t.Fatal(err)
		}

		for typ, keys := range scoring.CategoryKeys {
			for _, k := range keys {
				if cat.Categories[typ][k] == "" {
					t.Errorf("%s: missing categories.%s.%s", locale, typ, k)
				}
			}
		}

		for _, k := range []string{"valuetext", "unset"} {
			if cat.Rating[k] == "" {
				t.Errorf("%s: missing rating.%s", locale, k)
			}
		}
	}
}
