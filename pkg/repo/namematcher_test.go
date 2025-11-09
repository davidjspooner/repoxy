package repo

import (
	"strings"
	"testing"
)

func TestNameMatchersPrefersMostSpecificMatch(t *testing.T) {
	t.Parallel()

	var matchers NameMatchers
	if err := matchers.Set([]string{"library/*", "library/alpine"}); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	wildcard := matchers.GetMatchWeight(strings.Split("library/bash", "/"))
	specific := matchers.GetMatchWeight(strings.Split("library/alpine", "/"))

	if specific <= wildcard {
		t.Fatalf("expected specific mapping weight (%d) to exceed wildcard (%d)", specific, wildcard)
	}
}
