package prompt

import (
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

// WithFuzzy picks using
func WithFuzzy(a Args) (int, error) {
	lines := make([]string, len(a.Options))
	for index, item := range a.Options {
		line := strings.Join(item, " ")
		lines[index] = line
	}

	return fuzzyfinder.Find(lines, func(i int) string {
		return lines[i]
	})
}
