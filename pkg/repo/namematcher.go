package repo

import (
	"fmt"
	"strings"
)

type NameMatcher struct {
	parts  []string
	weight int
}

type NameMatchers []NameMatcher

func (nm *NameMatchers) Set(mapping []string) error {
	for _, m := range mapping {
		parts := strings.Split(m, "/")
		if len(parts) > 0 {
			weight := 1
			for _, part := range parts {
				if part != "*" {
					weight++ // Increment weight for wildcard parts
				}
			}
			*nm = append(*nm, NameMatcher{
				parts:  parts,
				weight: weight,
			})
		} else {
			return fmt.Errorf("invalid mapping '%s' in config", m)
		}
	}
	return nil
}

func (nm NameMatchers) GetMatchWeight(name []string) int {
	bestWeight := 0
	for _, matcher := range nm {
		if len(name) != len(matcher.parts) {
			continue // Not enough parts to match
		}
		match := true
		for i, part := range matcher.parts {
			if part != "*" && part != name[i] {
				match = false
				break
			}
		}
		if match {
			if matcher.weight > bestWeight {
				bestWeight = matcher.weight
			}
		}
	}
	return bestWeight
}
