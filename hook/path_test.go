package hook

import (
	"testing"
)

func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain", "hello world", "hello-world"},
		{"mixed punctuation", "Fix the bug — now!", "fix-the-bug-now"},
		{"already-slug", "already-a-slug", "already-a-slug"},
		{"trim hyphens", "---foo---bar---", "foo-bar"},
		{"unicode dropped", "café résumé", "caf-r-sum"},
		{"emoji dropped", "ship it 🚢🚢🚢", "ship-it"},
		{"all symbols → empty", "!@#$%^&*()", ""},
		{"cap at 50 boundary", "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda", "alpha-beta-gamma-delta-epsilon-zeta-eta-theta"},
		{"kelvin sign treated as non-ASCII", "K K K", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugifyTitle(tt.in)
			if got != tt.want {
				t.Errorf("SlugifyTitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
