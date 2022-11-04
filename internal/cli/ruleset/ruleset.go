package ruleset

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/clintjedwards/hclvet/internal/cli/appcfg"
	"github.com/clintjedwards/polyfmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/spf13/cobra"
)

// CmdRuleset is a subcommand for ruleset
var CmdRuleset = &cobra.Command{
	Use:   "ruleset",
	Short: "Manage linting rulesets",
	Long: `Manage linting rulesets.

Rulesets are a grouping of rules that are used to lint documents.

The ruleset subcommand allows you to retrieve, remove, and otherwise manipulate particular rulesets.`,
}

// state tracks application state over the time it takes a command to run.
type state struct {
	fmt polyfmt.Formatter
	cfg *appcfg.Appcfg
}

// rulesetInfo is the struct representation of the ruleset.hcl file included in all ruleset repos.
type rulesetInfo struct {
	Name    string `hcl:"name"`
	Version string `hcl:"version"`
}

// newState returns a new initialized state object
func newState(initialFmtMsg, format string) (*state, error) {
	clifmt, err := polyfmt.NewFormatter(polyfmt.Mode(format), false)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	clifmt.Print(initialFmtMsg, polyfmt.Pretty)

	cfg, err := appcfg.GetConfig()
	if err != nil {
		errText := fmt.Sprintf("error reading config file %q: %v", appcfg.ConfigFilePath(), err)
		clifmt.PrintErr(errText)
		clifmt.Finish()
		return nil, errors.New(errText)
	}

	return &state{
		fmt: clifmt,
		cfg: cfg,
	}, nil
}

// getRemoteRuleset is used to retrieve a ruleset from the path given
// It supports a wide range of remote and local paths
//
// See https://github.com/hashicorp/go-getter#url-format for accepted formats.
func getRemoteRuleset(srcPath, dstPath string) error {
	// For some reason the go-getter file resolution seems to be not working as documented.
	// Particularly when we call a relative path we get a failure.
	// To mitigate this somewhat we check if the user is trying to call a relative path and if so
	// then just use go's way of getting the absolute path.
	if strings.HasPrefix(srcPath, "./") || strings.HasPrefix(srcPath, "../") {
		absSrcPath, err := filepath.Abs(srcPath)
		if err == nil {
			srcPath = absSrcPath
		}
	}

	_, err := getter.Get(context.Background(), dstPath, srcPath)
	if err != nil {
		return err
	}

	return nil
}

// getRemoteRulesetInfo parses the ruleset.hcl file that must be included in all ruleset repos.
func getRemoteRulesetInfo(repoPath string) (rulesetInfo, error) {
	var info rulesetInfo

	rulesetFilePath := fmt.Sprintf("%s/%s", repoPath, "ruleset.hcl")
	err := hclsimple.DecodeFile(rulesetFilePath, nil, &info)
	if err != nil {
		return rulesetInfo{}, err
	}

	return info, nil
}

// buildAllRules builds the plugins(rules are plugins) and places the binary
// underneath the correct ruleset directory.
func buildAllRules(s *state, ruleset string) error {
	s.fmt.Print("Opening rules directory")

	file, err := os.Open(appcfg.RepoRulesPath(ruleset))
	if err != nil {
		errText := fmt.Sprintf("could not open rules folder: %v", err)
		s.fmt.PrintErr(errText)
		s.fmt.Finish()
		return errors.New(errText)
	}
	defer file.Close()

	// get a list of all folders within the rules directory, which should represent rules.
	fileList, err := file.Readdir(0)
	if err != nil {
		errText := fmt.Sprintf("could not read rules folder: %v", err)
		s.fmt.PrintErr(errText)
		s.fmt.Finish()
		return errors.New(errText)
	}

	startTime := time.Now()
	count := 0

	// Rules are separated into directories. We iterate through directories and build whats inside
	// them.
	for _, file := range fileList {
		if !file.IsDir() {
			continue
		}

		// Get the dirname and not the full path.
		// Sometimes file.Name will return the full path based on what is passed to file.Open.
		dirName := filepath.Base(file.Name())

		s.fmt.Print(fmt.Sprintf("Compiling %s", dirName))

		rawRulePath := fmt.Sprintf("%s/%s", appcfg.RepoRulesPath(ruleset), dirName)

		// We take the hash of the dirname(aka the rule folder name) and make it the rule ID.
		// This allows us to have consistent ids for rules without having to have the ruleset
		// (aka the user) define them.
		//
		// TODO(clintjedwards): Collision detection isn't built in and will probably break
		// things if it ever does happen.
		ruleID := generateHash(dirName)

		// We build here by pointing the golang binary on the user's computer to the rule path.
		// This causes the compiler to compile whatever is in that path and spit out a binary
		// where ever we want.
		_, err := buildRule(rawRulePath, appcfg.RulePath(ruleset, ruleID))
		if err != nil {
			errText := fmt.Sprintf("could not build rule %s: %v", dirName, err)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return errors.New(errText)
		}

		s.fmt.Print(fmt.Sprintf("Collecting rule info for: %s", dirName))
		newRule, err := getRuleInfo(ruleset, ruleID)
		if err != nil {
			errText := fmt.Sprintf("could not build rule %s: %v", dirName, err)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return errors.New(errText)
		}

		err = s.cfg.UpsertRule(ruleset, newRule)
		if err != nil {
			errText := fmt.Sprintf("could not upsert rule %s to config file: %v", dirName, err)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return errors.New(errText)
		}
		count++
	}

	duration := time.Since(startTime)
	durationSeconds := float64(duration) / float64(time.Second)
	timePerRule := float64(duration) / float64(count)

	s.fmt.PrintSuccess(fmt.Sprintf("Compiled %d rule(s) in %.2fs (average %.2fms/rule)",
		count, durationSeconds, timePerRule/float64(time.Millisecond)))

	return nil
}

// verifyRuleset makes sure a downloaded ruleset has the correct structure.
//   - Makes sure the ruleset has a proper version and name.
//   - Makes sure the ruleset has a rules folder.
func verifyRuleset(path string, info rulesetInfo) error {
	isValidNameErr := validation.NewError("validation_is_valid_name", "must be a valid name; alphanumeric characters and _ or - only")
	isValidName := validation.NewStringRuleWithError(isValidName, isValidNameErr)

	err := validation.Validate(info.Name,
		validation.Required,      // not empty
		validation.Length(3, 20), // within length reqs
		isValidName,
	)
	if err != nil {
		return fmt.Errorf("ruleset name malformed: %w", err)
	}

	_, err = semver.NewVersion(info.Version)
	if err != nil {
		return fmt.Errorf("ruleset version text malformed; should be in semvar notation: %v", err)
	}

	// Must have a /rules directory
	rulesDirPath := fmt.Sprintf("%s/%s", path, "rules")
	if _, err := os.Stat(rulesDirPath); os.IsNotExist(err) {
		return errors.New("no rules directory found; all rulesets must have a rules directory")
	}

	return nil
}

func isValidName(value string) bool {
	const validNameCharSet string = "^[a-zA-Z0-9_-]+$"
	validNameRegexp := regexp.MustCompile(validNameCharSet)
	return validNameRegexp.MatchString(value)
}
