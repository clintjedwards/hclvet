package ruleset

import (
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	"github.com/clintjedwards/hclvet/internal/cli/appcfg"
	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/spf13/cobra"
)

var cmdRulesetUpdate = &cobra.Command{
	Use:   "update [ruleset]",
	Short: "Retrieves most recent rules for a ruleset",
	Long: `Update will attempt to download rule updates for a specific ruleset.

Running without arguments will update all rulesets.

It works by comparing the version of the ruleset currently installed with the remote ruleset.
If the version downloaded is different than the local version this will trigger a
recompilation of all rules.

The resolution process is very basic and does not perform any more than a rudimentary check for diffs
and as such, for sufficiently large repositories this might be a heavy operation.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetUpdate)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Updating ruleset", format)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		for _, ruleset := range state.cfg.Rulesets {
			state.fmt.Print(fmt.Sprintf("Updating ruleset %s", ruleset.Name))
			err := updateRuleset(state, ruleset)
			if err != nil {
				state.fmt.PrintErr(fmt.Sprintf("could not update ruleset %s", ruleset.Name))
				state.fmt.Finish()
				return err
			}
		}
		state.fmt.PrintSuccess("Updated all rulesets")
		state.fmt.Finish()
		return nil
	}

	rulesetName := args[0]
	state.fmt.Print(fmt.Sprintf("Updating ruleset %s", rulesetName))
	ruleset, err := state.cfg.GetRuleset(rulesetName)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not find ruleset %s", rulesetName))
		state.fmt.Finish()
		return err
	}
	err = updateRuleset(state, ruleset)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not update ruleset %s", ruleset.Name))
		state.fmt.Finish()
		return err
	}
	state.fmt.PrintSuccess("Updated all rulesets")
	state.fmt.Finish()

	return nil
}

func updateRuleset(s *state, ruleset models.Ruleset) error {

	s.fmt.Print("Retrieveing ruleset")
	err := getRemoteRuleset(ruleset.Repository, appcfg.RepoPath(ruleset.Name))
	if err != nil {
		return err
	}

	s.fmt.Print("Parsing remote info")
	info, err := getRemoteRulesetInfo(appcfg.RepoPath(ruleset.Name))
	if err != nil {
		return err
	}

	s.fmt.Print("Verifying ruleset")
	err = verifyRuleset(appcfg.RepoPath(ruleset.Name), info)
	if err != nil {
		return err
	}

	newSemver, err := semver.NewVersion(info.Version)
	if err != nil {
		return err
	}

	oldSemver, err := semver.NewVersion(ruleset.Version)
	if err != nil {
		return err
	}

	if !newSemver.GreaterThan(oldSemver) {
		s.fmt.PrintSuccess(fmt.Sprintf("Ruleset %s at newest version (%s)", ruleset.Name, ruleset.Version))
		return nil
	}

	s.fmt.PrintSuccess(fmt.Sprintf("Found newer ruleset for %s (current: %s, remote: %s)",
		ruleset.Name, ruleset.Version, info.Version))

	s.fmt.Print("Updating ruleset")
	err = s.cfg.UpdateRuleset(models.Ruleset{
		Name:       info.Name,
		Version:    info.Version,
		Repository: ruleset.Repository,
		Enabled:    ruleset.Enabled,
		Rules:      ruleset.Rules,
	})
	if err != nil {
		return err
	}

	err = buildAllRules(s, info.Name)
	if err != nil {
		return err
	}

	return nil
}
