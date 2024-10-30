package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"os"
	"strings"
)

var config ConfigObj
var configLoaded bool

type ConfigObj struct {
	Version int
	Cache   CacheConfig   `json:"cache"`
	Scoring ScoringConfig `json:"scoring"`
	Display DisplayConfig `json:"display"`
}

type CacheConfig struct {
	TTLSecs int64 `json:"ttl_secs"`
}

type ScoringConfig struct {
	Rules []ScoringRule `json:"rules"`
}

type DisplayConfig struct {
	ScoreThreshold int    `json:"score_threshold"`
	Theme          string `json:"theme"`
}

type ScoringRule struct {
	TextFound   string   `json:"text_found,omitempty"`
	TextMissing string   `json:"text_missing,omitempty"`
	Score       int      `json:"score"`
	TagsWhy     []string `json:"tags_why,omitempty"`
	TagsWhyNot  []string `json:"tags_why_not,omitempty"`
	Colorize    *bool    `json:"colorize,omitempty"` // pointer for default nil instead of false
}

func GetConfig() ConfigObj {
	if !configLoaded {
		err := Reload()
		if err != nil {
			panic(err)
		}
	}
	return config
}

func GetPath() (string, error) {
	// ~/.config/
	return xdg.ConfigFile("hnjobs/config.json") //creates path if needed
}

func Reload() error {
	configPath, err := GetPath()
	if err != nil {
		return err
	}
	return loadConfigFile(configPath)
}

func loadConfigFile(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file \"%s\": %v", filename, err)
	}
	err = loadConfigJSON(contents)
	if err != nil {
		return fmt.Errorf("error loading config file \"%s\": %v", filename, err)
	}
	return nil
}

func loadConfigJSON(j []byte) error {
	config = ConfigObj{}
	err := json.Unmarshal(j, &config)
	if err != nil {
		return err
	}
	// validate
	for i, r := range config.Scoring.Rules {
		if r.TextFound == "" && r.TextMissing == "" {
			return errors.New("scoring rules must have either `text_found` or `text_missing`")
		}
		if r.TextFound != "" && r.TextMissing != "" {
			return errors.New("scoring rules cannot have both `text_found` and `text_missing`")
		}
		if r.TextFound != "" {
			config.Scoring.Rules[i].TextFound = strings.ToLower(r.TextFound)
		}
		if r.TextMissing != "" {
			config.Scoring.Rules[i].TextMissing = strings.ToLower(r.TextMissing)
		}
	}
	// TODO validate that tags are sane (json-compliant, sqlite-compliant, no spaces)

	configLoaded = true
	return nil
}

func DefaultConfigFileContents() []byte {
	// I want the generated rules to be pleasant to edit, which means one rule per line.
	// There's no way to get the standard json package to marshal this way.

	// Don't forget to update the tests
	rules := []ScoringRule{
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
	}

	escapeString := func(s string) string {
		// see json/encode.go:appendString()
		s = strings.ReplaceAll(s, "\\b", "\\\\b")
		s = strings.ReplaceAll(s, "\\f", "\\\\f")
		s = strings.ReplaceAll(s, "\\n", "\\\\n")
		s = strings.ReplaceAll(s, "\\r", "\\\\r")
		s = strings.ReplaceAll(s, "\\t", "\\\\t")
		return s
	}

	prettyMarshalRule := func(sr *ScoringRule) ([]byte, error) {
		var elems []string
		if !(sr.TextFound == "") {
			elems = append(elems, fmt.Sprintf(
				`"text_found": "%s"`,
				escapeString(sr.TextFound),
			))
		}
		if !(sr.TextMissing == "") {
			elems = append(elems, fmt.Sprintf(
				`"text_missing": "%s"`,
				escapeString(sr.TextMissing),
			))
		}
		elems = append(elems, fmt.Sprintf(`"score": %d`, sr.Score))
		if len(sr.TagsWhy) > 0 {
			var quoted []string
			for _, tag := range sr.TagsWhy {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, tag))
			}
			elems = append(elems, fmt.Sprintf(
				`"tags_why": [%s]`,
				escapeString(strings.Join(quoted, ", ")),
			))
		}
		if len(sr.TagsWhyNot) > 0 {
			var quoted []string
			for _, tag := range sr.TagsWhyNot {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, tag))
			}
			elems = append(elems, fmt.Sprintf(
				`"tags_why_not": [%s]`,
				escapeString(strings.Join(quoted, ", ")),
			))
		}
		return []byte(fmt.Sprintf(`{%s}`, strings.Join(elems, ", "))), nil
	}

	var renderedRules []string
	for _, r := range rules {
		j, err := prettyMarshalRule(&r)
		if err != nil {
			panic(err)
		}
		renderedRules = append(renderedRules, "      "+string(j))
	}
	rulesOut := strings.Join(renderedRules, ",\n")
	out := fmt.Sprintf(`{
  "version": 1,
  "cache": {
    "ttl_secs": 86400
  },
  "scoring": {
    "rules": [
%s
    ]
  },
  "display": {
    "theme": "default",
    "score_threshold": 1
  }
}
`, rulesOut)

	return []byte(out)
}
