package appcfg

import (
	"fmt"
	"log"

	"github.com/clintjedwards/hclvet/internal/config"
	"github.com/mitchellh/go-homedir"
)

const (
	// rulesetsDirName is the name of the config directory that holds rulesets.
	rulesetsDirName string = "rulesets.d"

	// configFileName is the name of the config file that stores app information.
	configFileName string = ".hclvet.hcl"

	// repoDirName is the name of the directory that stores the raw ruleset folder.
	repoDirName string = "repo"

	// rulesDirName is the name of the directory
	rulesDirName string = "rules"
)

// Config paths

// ConfigPath returns the absolute config path determined by environment variable.
// The default is ~/.hclvet.d
func ConfigPath() string {
	config, err := config.FromEnv()
	if err != nil {
		log.Fatalf("could not access config: %v", err)
	}

	absConfigPath, err := homedir.Expand(config.ConfigPath)
	if err != nil {
		log.Fatalf("could not access config: %v", err)
	}

	return absConfigPath
}

// ConfigFilePath returns the absolute path of the configuration file.
// By default this is ~/.hclvet.d/.hclvet.hcl
func ConfigFilePath() string {
	return fmt.Sprintf("%s/%s", ConfigPath(), configFileName)
}

// RulesetsPath returns the absolute directory path of the directory that stores rulesets.
// By default this is ~/.hclvet.d/rulesets.d
//
// Note: this is not the ruleset itself just the parent folder.
// Use RulesetPath() to get a path to a specific ruleset.
func RulesetsPath() string {
	return fmt.Sprintf("%s/%s", ConfigPath(), rulesetsDirName)
}

// RulesetPath returns the directory path to a supplied ruleset name.
// By default this is ~/.hclvet.d/rulesets.d/<ruleset>
func RulesetPath(ruleset string) string {
	return fmt.Sprintf("%s/%s", RulesetsPath(), ruleset)
}

// RepoPath returns the absolute path for the repo directory inside of a specific ruleset.
// By default this is ~/.hclvet.d/rulesets.d/<ruleset>/repo
func RepoPath(ruleset string) string {
	return fmt.Sprintf("%s/%s", RulesetPath(ruleset), repoDirName)
}

// RepoRulesPath returns the absolute path for the rules directory inside of a repo directory of
// a ruleset
// By default this is ~/.hclvet.d/rulesets.d/<ruleset>/repo/rules
func RepoRulesPath(ruleset string) string {
	return fmt.Sprintf("%s/%s", RepoPath(ruleset), rulesDirName)
}

// RulePath returns the absolute path for a rule within a ruleset directory.
// By default this is ~/.hclvet.d/rulesets.d/<ruleset>/<ruleID>
func RulePath(ruleset, ruleID string) string {
	return fmt.Sprintf("%s/%s", RulesetPath(ruleset), ruleID)
}
