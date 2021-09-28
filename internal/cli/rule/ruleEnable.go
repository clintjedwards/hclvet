package rule

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRuleEnable = &cobra.Command{
	Use:   "enable <ruleset> <rule>",
	Short: "Turns on a rule",
	Long:  `Turns on a particular rule, causing it to be skipped on all future linting runs.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	rule := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Enabling rule", format)
	if err != nil {
		return err
	}

	err = state.cfg.SetRuleEnabled(ruleset, rule, true)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not enable rule: %v", err))
		state.fmt.Finish()
		return err
	}

	state.fmt.PrintSuccess(fmt.Sprintf("Enabled rule %s", rule))
	state.fmt.Finish()
	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleEnable)
}
