package config

import (
	"reflect"
	"testing"
)

func TestDefaultContents(t *testing.T) {
	expected := ConfigObj{
		Version: 1,
		Cache: CacheConfig{
			TTLSecs: 86400,
		},
		Scoring: ScoringConfig{
			Rules: []ScoringRule{
				{
					TextFound: "(?i)sre",
					Score:     1,
					TagsWhy:   []string{"career"},
				},
				{
					TextFound: "(?i)reliability",
					Score:     1,
					TagsWhy:   []string{"career"},
				},
				{
					TextFound: "(?i)resiliency",
					Score:     1,
					TagsWhy:   []string{"career"},
				},
				{
					TextFound: "(?i)principal",
					Score:     1,
					TagsWhy:   []string{"level"},
				},
				{
					TextFound: "(?i)staff\\b",
					Score:     1,
					TagsWhy:   []string{"level"},
				},
				{
					TextFound: "(?i)aws",
					Score:     1,
					TagsWhy:   []string{"tech"},
				},
				{
					TextFound: "(?i)\\brust\\b",
					Score:     1,
					TagsWhy:   []string{"tech", "memecred"},
				},
				{
					TextFound: "(?i)golang",
					Score:     1,
					TagsWhy:   []string{"tech"},
				},
				{
					TextFound: "(?i)\\bgo\\b",
					Score:     1,
					TagsWhy:   []string{"tech"},
				},
				{
					TextFound: "(?i)open.?source",
					Score:     2,
					TagsWhy:   []string{"values"},
				},
				{
					TextFound: "(?i)education",
					Score:     2,
					TagsWhy:   []string{"values"},
				},
				{
					TextFound: "(?i)\\bheal",
					Score:     2,
					TagsWhy:   []string{"values"},
				},
				{
					TextMissing: "(?i)remote",
					Score:       -100,
					TagsWhyNot:  []string{"onsite", "fsckbezos"},
				},
			},
		},
		Display: DisplayConfig{
			ScoreThreshold: 1,
			Theme:          "default",
		},
	}

	err := loadConfigJSON(DefaultConfigFileContents())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, config) {
		t.Errorf(
			"Config contents are not the same.\n  Expected:\n%v\n  Got:\n%v",
			expected, config,
		)
	}
}
