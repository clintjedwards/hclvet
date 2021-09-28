package ruleset

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRulesetEnable = &cobra.Command{
	Use:   "enable <ruleset>",
	Short: "Turns on a ruleset",
	Long:  `Turns on a particular ruleset, causing it to be skipped on all future linting runs.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Enabling ruleset", format)
	if err != nil {
		return err
	}

	err = state.cfg.SetRulesetEnabled(ruleset, true)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not enable ruleset: %v", err))
		state.fmt.Finish()
		return err
	}

	state.fmt.PrintSuccess(fmt.Sprintf("Enabled ruleset %s", ruleset))
	state.fmt.Finish()
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetEnable)
}
