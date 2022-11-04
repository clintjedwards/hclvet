package ruleset

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/clintjedwards/polyfmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var cmdRulesetList = &cobra.Command{
	Use:   "list [ruleset]",
	Short: "Lists a ruleset and its rules",
	Long: `Allows the listing of a ruleset and its rules.

If no argument is provided, list will display all possible rulesets and relevant details.
If a ruleset is provided, list will display the ruleset's details and rules.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("", format)
	if err != nil {
		return err
	}

	const fullListTmpl = `== Summary ==

{{.RulesetList}}

== Rulesets ==

{{.Rulesets}}
`

	if len(args) == 0 {
		tmplRulesetList := formatAllRulesets(state.cfg.Rulesets)
		tmplRulesets := ""
		for _, ruleset := range state.cfg.Rulesets {
			tmplRulesets += formatRuleset(ruleset)
		}

		var tpl bytes.Buffer
		t := template.Must(template.New("tmp").Parse(fullListTmpl))
		_ = t.Execute(&tpl, struct {
			RulesetList string
			Rulesets    string
		}{
			RulesetList: tmplRulesetList,
			Rulesets:    tmplRulesets,
		})

		state.fmt.Println(tpl.String(), polyfmt.Pretty)
		state.fmt.Println(state.cfg.Rulesets, polyfmt.JSON)

		return nil
	}

	rulesetName := args[0]
	ruleset, err := state.cfg.GetRuleset(rulesetName)
	if err != nil {
		state.fmt.PrintErr(fmt.Sprintf("could not find ruleset %s", rulesetName))
		state.fmt.Finish()
		return err
	}
	state.fmt.Println(formatRuleset(ruleset), polyfmt.Pretty)
	state.fmt.Println(ruleset, polyfmt.JSON)
	state.fmt.Finish()
	return nil
}

func formatAllRulesets(rulesets []models.Ruleset) string {
	headers := []string{"Name", "Version", "Repository", "Enabled", "Rules"}
	data := [][]string{}

	for _, ruleset := range rulesets {
		data = append(data, []string{
			ruleset.Name,
			ruleset.Version,
			ruleset.Repository,
			strconv.FormatBool(ruleset.Enabled),
			strconv.Itoa(len(ruleset.Rules)),
		})
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader(headers)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func formatRuleset(ruleset models.Ruleset) string {
	enabledStr := ""
	if ruleset.Enabled {
		enabledStr = "enabled"
	} else {
		enabledStr = "disabled"
	}

	// Example v1.0.0 (enabled) [2 rule(s)]
	title := fmt.Sprintf("%s %s (%s) [%d rule(s)]\n\n",
		cases.Title(language.AmericanEnglish).String(ruleset.Name),
		ruleset.Version, enabledStr, len(ruleset.Rules))

	headers := []string{"Rule", "Name", "Description", "Enabled"}
	data := [][]string{}

	for _, rule := range ruleset.Rules {
		data = append(data, []string{
			rule.ID,
			rule.Name,
			rule.Short,
			strconv.FormatBool(rule.Enabled),
		})
	}

	tableString := &strings.Builder{}
	tableString.WriteString(title)
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader(headers)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetList)
}
