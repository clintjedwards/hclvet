// Package appcfg controls actions that can be performed around the app's configuration file and
// config directory. Currently we use HCL as the config file format.
package appcfg

import (
	"errors"
	"os"
	"strings"

	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Appcfg represents the parsed hcl config of the main app configuration.
// We wrap this so that we can add other attributes in here.
type Appcfg struct {
	Rulesets []models.Ruleset `hcl:"ruleset,block"`
}

// CreateNewFile creates a new empty config file
func CreateNewFile() error {
	cfgFile := hclwrite.NewEmptyFile()

	f, err := os.Create(ConfigFilePath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(cfgFile.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// GetConfig parses the on disk config file and returns its representation in golang.
func GetConfig() (*Appcfg, error) {
	hclFile := &Appcfg{}

	err := hclsimple.DecodeFile(ConfigFilePath(), nil, hclFile)
	if err != nil {
		return nil, err
	}

	return hclFile, nil
}

// RepositoryExists checks to see if the config already has an entry for the repository in the config.
func (appcfg *Appcfg) RepositoryExists(repo string) bool {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Repository != repo {
			continue
		}

		return true
	}

	return false
}

// AddRuleset adds a new ruleset. Returns error if ruleset already exists.
func (appcfg *Appcfg) AddRuleset(rs models.Ruleset) error {
	if appcfg.RulesetExists(rs.Name) {
		return errors.New("ruleset already exists")
	}

	// cast ruleset name to lower
	rs.Name = strings.ToLower(rs.Name)

	appcfg.Rulesets = append(appcfg.Rulesets, rs)
	err := appcfg.writeConfig()
	if err != nil {
		return err
	}

	return nil
}

// UpdateRuleset updates an existing ruleset. Returns an error if the ruleset could not be found.
func (appcfg *Appcfg) UpdateRuleset(rs models.Ruleset) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rs.Name {
			continue
		}

		// cast ruleset name to lower
		rs.Name = strings.ToLower(rs.Name)

		appcfg.Rulesets[index] = rs
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("could not find ruleset")
}

// RulesetExists determines if a ruleset has already been added.
func (appcfg *Appcfg) RulesetExists(name string) bool {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name == name {
			return true
		}
	}

	return false
}

// UpsertRule adds a new rule to an already established ruleset if it does not exist. If the rule
// already exists it simply updates the rule with the newer information.
// Returns an error if the ruleset is not found or there was an error writing to the file.
func (appcfg *Appcfg) UpsertRule(rulesetName string, newRule models.Rule) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for ruleIndex, rule := range ruleset.Rules {
			if rule.ID == newRule.ID {

				// Keep user settings for updated rule
				newRule.Enabled = rule.Enabled

				appcfg.Rulesets[index].Rules[ruleIndex] = newRule
				err := appcfg.writeConfig()
				if err != nil {
					return err
				}
				return nil
			}
		}

		ruleset.Rules = append(ruleset.Rules, newRule)
		appcfg.Rulesets[index] = ruleset
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

// SetRulesetEnabled changes the enabled attribute on a ruleset.
// Returns an error if the ruleset isn't found.
func (appcfg *Appcfg) SetRulesetEnabled(name string, enabled bool) error {
	for index, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		appcfg.Rulesets[index].Enabled = enabled
		err := appcfg.writeConfig()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("ruleset not found")
}

// SetRuleEnabled changes the enabled attribute on a rule.
// Returns an error if the ruleset or rule isn't found.
func (appcfg *Appcfg) SetRuleEnabled(ruleset, rule string, enabled bool) error {
	for _, rs := range appcfg.Rulesets {
		if rs.Name != ruleset {
			continue
		}

		for index, r := range rs.Rules {
			if r.ID != rule {
				continue
			}

			rs.Rules[index].Enabled = enabled
			err := appcfg.writeConfig()
			if err != nil {
				return err
			}

			return nil
		}
		return errors.New("rule not found")
	}

	return errors.New("ruleset not found")
}

// GetRuleset returns the ruleset object of a given name.
// Returns an error if ruleset isn't found.
func (appcfg *Appcfg) GetRuleset(name string) (models.Ruleset, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != name {
			continue
		}

		return ruleset, nil
	}

	return models.Ruleset{}, errors.New("ruleset not found")
}

// GetRule returns the rule object of a given name.
// Returns an error if ruleset or rule isn't found.
func (appcfg *Appcfg) GetRule(rulesetName, ruleID string) (models.Rule, error) {
	for _, ruleset := range appcfg.Rulesets {
		if ruleset.Name != rulesetName {
			continue
		}

		for _, rule := range ruleset.Rules {
			if rule.ID != ruleID {
				continue
			}

			return rule, nil
		}

		return models.Rule{}, errors.New("rule not found")
	}

	return models.Rule{}, errors.New("ruleset not found")
}

// writeConfig takes the current representation of config and writes it to the file.
func (appcfg *Appcfg) writeConfig() error {
	f := hclwrite.NewEmptyFile()

	gohcl.EncodeIntoBody(appcfg, f.Body())

	err := os.WriteFile(ConfigFilePath(), f.Bytes(), 0o644)
	if err != nil {
		return err
	}

	return nil
}
