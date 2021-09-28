package rule

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/clintjedwards/polyfmt"
	"github.com/spf13/cobra"
)

var cmdRuleDescribe = &cobra.Command{
	Use:   "describe <ruleset> <rule>",
	Short: "Prints details about a rule",
	Long:  `Prints extended information about a particular rule.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDescribe,
}

func runDescribe(cmd *cobra.Command, args []string) error {
	ruleset := args[0]
	ruleID := args[1]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		return err
	}

	rule, err := state.cfg.GetRule(ruleset, ruleID)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not describe rule %v", err))
		state.fmt.Finish()
		return err
	}

	const describeTmpl = `[{{.ID}}] {{.Name}}

{{.Short}}

{{.Long}}
Enabled: {{.Enabled}} | Link: {{.Link}}`

	var tpl bytes.Buffer
	t := template.Must(template.New("tmp").Parse(describeTmpl))
	_ = t.Execute(&tpl, struct {
		ID      string
		Name    string
		Short   string
		Long    string
		Enabled bool
		Link    string
	}{
		ID:      rule.ID,
		Name:    rule.Name,
		Short:   rule.Short,
		Long:    strings.TrimPrefix(rule.Long, "\n"),
		Enabled: rule.Enabled,
		Link:    rule.Link,
	})

	state.fmt.Println(tpl.String(), polyfmt.Pretty)
	state.fmt.Println(rule, polyfmt.JSON)

	return nil
}

func init() {
	CmdRule.AddCommand(cmdRuleDescribe)
}
