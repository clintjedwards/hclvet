package rule

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cmdRuleDisable = &cobra.Command{
	Use:   "disable <ruleset> <rule>",
	Short: "Turns off a rule",
	Long:  `Turns off a particular rule, causing it to be skipped on all future linting runs.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDisable,
}

func runDisable(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	rule := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Disabling rule", format)
	if err != nil {
		return err
	}

	err = state.cfg.SetRuleEnabled(ruleset, rule, false)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not disable rule: %v", err))
		state.fmt.Finish()
		return err
	}

	state.fmt.PrintSuccess(fmt.Sprintf("Disabled rule %s", rule))
	state.fmt.Finish()
	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleDisable)
}
