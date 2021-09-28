package cli

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/olekukonko/tablewriter"
)

// PrintLintError formats and prints details from a lint error.
//
// It borrows(blatantly copies) from rust style errors:
// https://doc.rust-lang.org/edition-guide/rust-2018/the-compiler/improved-error-messages.html
func formatLintError(lintErr models.LintError) string {
	const lintErrorTmpl = `Error[{{.ID}}]: {{.Short}}
  --> {{.Filepath}}:{{.StartLine}}:{{.StartColumn}}
{{.LineText}}
  = additional information:
{{.Metadata}}
For more information about this error, try running ` + "`hclvet rule describe {{.Ruleset}} {{.ID}}`."

	var tpl bytes.Buffer
	t := template.Must(template.New("tmp").Parse(lintErrorTmpl))
	_ = t.Execute(&tpl, struct {
		ID          string
		Short       string
		Filepath    string
		StartLine   int
		StartColumn int
		LineText    string
		Metadata    string
		Ruleset     string
	}{
		ID:          lintErr.Rule.ID,
		Short:       lintErr.Rule.Short,
		Filepath:    lintErr.Filepath,
		StartLine:   int(lintErr.RuleErr.Location.Start.Line),
		StartColumn: int(lintErr.RuleErr.Location.Start.Column),
		LineText:    formatLineTable(lintErr.Line, int(lintErr.RuleErr.Location.Start.Line)),
		Metadata:    formatAdditionalInfo(lintErr),
		Ruleset:     lintErr.Ruleset,
	})

	return tpl.String()
}

// formatLineTable returns a pretty printed string of an error line
func formatLineTable(line string, lineNum int) string {
	data := [][]string{
		{"", "|", ""},
		{" " + strconv.Itoa(lineNum), "|", line},
		{"", "|", ""},
		{"", "|", ""},
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetColMinWidth(0, 3)
	table.SetRowSeparator("")
	table.SetBorder(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT})
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}

func formatAdditionalInfo(lintErr models.LintError) string {
	data := [][]string{
		{" ", "• name:", lintErr.Rule.Name},
		{" ", "• link:", lintErr.Rule.Link},
	}

	if lintErr.RuleErr.Suggestion != "" {
		data = append(data, []string{" ", "• suggestion:", lintErr.RuleErr.Suggestion})
	}
	if lintErr.RuleErr.Remediation != "" {
		data = append(data,
			[]string{" ", "• remediation:", fmt.Sprintf("`%s`", lintErr.RuleErr.Remediation)})
	}

	if len(lintErr.RuleErr.Metadata) != 0 {
		for key, value := range lintErr.RuleErr.Metadata {
			data = append(data,
				[]string{" ", fmt.Sprintf("• %s:", key), value},
			)
		}
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}
