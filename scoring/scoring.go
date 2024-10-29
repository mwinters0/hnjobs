package scoring

import (
	"fmt"
	"github.com/mwinters0/hnjobs/config"
	"github.com/mwinters0/hnjobs/db"
	"regexp"
	"slices"
)

type RuleType int

const (
	TextFound RuleType = iota
	TextMissing
)

func (rt RuleType) String() string {
	switch rt {
	case TextFound:
		return "TextFound"
	case TextMissing:
		return "TextMissing"
	default:
		panic(fmt.Errorf("unhandled rule type %d", rt))
	}
}

type Rule struct {
	config.ScoringRule
	RuleType RuleType
	Regex    *regexp.Regexp
}

func newRuleFromConf(confRule *config.ScoringRule) (*Rule, error) {
	var rt RuleType
	if confRule.TextFound != "" {
		rt = TextFound
	} else if confRule.TextMissing != "" {
		rt = TextMissing
	}
	r := &Rule{
		*confRule,
		rt,
		nil,
	}
	switch rt {
	case TextFound:
		r.Regex = regexp.MustCompile(confRule.TextFound)
	case TextMissing:
		r.Regex = regexp.MustCompile(confRule.TextMissing)
	default:
		return nil, fmt.Errorf("unhandled rule type %d", rt)
	}
	return r, nil
}

var rules []*Rule

func ReloadRules() error {
	// create Rule list from config
	confRules := config.GetConfig().Scoring.Rules
	rules = make([]*Rule, len(confRules))
	for i, confRule := range confRules {
		r, err := newRuleFromConf(&confRule)
		if err != nil {
			return err
		}
		rules[i] = r
	}
	return nil
}

func GetRules() []*Rule {
	if len(rules) == 0 {
		err := ReloadRules()
		if err != nil {
			panic(err)
		}
	}
	return rules
}

func ScoreDBComment(dbc *db.Job) int {
	if len(rules) == 0 {
		err := ReloadRules()
		if err != nil {
			panic(err)
		}
	}
	dbc.Score = 0
	for _, r := range rules {
		applyRule(r, dbc)
	}
	return dbc.Score
}

func applyRule(rule *Rule, dbc *db.Job) {
	shouldMatch := true //is this a regular Rule (Regex should return true) or an inverse Rule (should return false)?
	if rule.RuleType == TextMissing {
		shouldMatch = false
	}
	matched := rule.Regex.MatchString(dbc.Text)
	if matched == shouldMatch {
		//Rule applies
		dbc.Score = dbc.Score + rule.Score
		for _, y := range rule.TagsWhy {
			if !slices.Contains(dbc.Why, y) {
				dbc.Why = append(dbc.Why, y)
			}
		}
		for _, n := range rule.TagsWhyNot {
			if !slices.Contains(dbc.WhyNot, n) {
				dbc.WhyNot = append(dbc.WhyNot, n)
			}
		}
	}
}
