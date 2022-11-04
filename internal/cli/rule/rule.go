package rule

import (
	"errors"
	"fmt"
	"log"

	"github.com/clintjedwards/hclvet/internal/cli/appcfg"
	"github.com/clintjedwards/polyfmt"
	"github.com/spf13/cobra"
)

// CmdRule is a subcommand for rule.
var CmdRule = &cobra.Command{
	Use:   "rule",
	Short: "Manage linting rules",
	Long: `Manage linting rules.

Rules are the constraints on which hclvet lints documents against.

The rule subcommand allows you to describe, enable, and otherwise manipulate particular rules.`,
}

// state tracks application state over the time it takes a command to run.
type state struct {
	fmt polyfmt.Formatter
	cfg *appcfg.Appcfg
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
