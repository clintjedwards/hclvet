package rule

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/clintjedwards/hclvet/internal/utils"
	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"
)

// cmdRuleCreate creates a skeleton ruleset
var cmdRuleCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new rule",
	Long: `Creates the files and folders needed to create a new rule.

Navigate to the ruleset folder in which you mean to create the rule. From there, simply run this command
to create all files and folders required for a hclvet rule.

The rule name should short, alphanumeric, and have no spaces. It will be used as the directory name
and hashed to provide the user a quick way to target said specific rule.
`,
	Example: `$ hclvet rule create example_rule_name`,
	RunE:    runCreate,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdRule.AddCommand(cmdRuleCreate)
}

func runCreate(_ *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	err := validateName(name)
	if err != nil {
		log.Println(err)
		return err
	}

	err = createRuleDir(name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func validateName(name string) error {
	err := validation.Validate(name,
		validation.Required,      // not empty
		validation.Length(3, 70), // within length reqs
		is.ASCII,
	)
	if err != nil {
		return fmt.Errorf("rule name malformed: %w", err)
	}

	return nil
}

func createRuleDir(name string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	//TODO(clintjedwards): Take this from the appcfg package and stop declaring it everywhere
	rulesDirName := "rules"
	ruleDirPath := fmt.Sprintf("%s/%s/%s", currentDir, rulesDirName, name)

	err = utils.CreateDir(ruleDirPath)
	if err != nil {
		return err
	}

	err = createMainFile(name)
	if err != nil {
		return err
	}

	return nil
}

func createMainFile(name string) error {
	const mainFileContent = `package main

import hclvet "github.com/clintjedwards/hclvet/sdk"

// Check is constructed so that we can fulfill the interface for the NewRule function below.
type Check struct{}

// Check is the logic of the linting rule. Consume the hclContent object and produce lint errors
// as your linting rule sees fit.
func (c *Check) Check(content []byte) ([]hclvet.RuleError, error) {
	// We declare lintErrors here so that we can append to it as we find errors within the file.
	var lintErrors []hclvet.RuleError

	// All HCL files are comprised of two components: "blocks" and within those blocks, "attributes".
	// The strategy for most rules is simply cycle through the blocks you're interested
	// in and perform some logic to make sure its in the state you expect.
	//
	// ParseHCL gives us back our HCL file neatly parsed into a struct representing those
	// nested blocks and attributes.
	hclContent := hclvet.ParseHCL(content)

	// This is where the actual linting logic is applied. Everytime we find an error we add
	// it to the errors list with its location.
	//
	// The example below parses through labels to find the name for the correct hcl resource
	// once it has found the correct resource it can perform some logic to make sure its in
	// a certain state and then continue if it is or append a new lint error if it isn't.
	//
	// The below is just an example and many parts can/should be changed for your use case.
	for _, block := range hclContent.Blocks {
		for _, label := range block.Labels {


			// <linting logic belongs here>

			// Once we find an error we log its location
			location := hclvet.Range{
				Start: hclvet.Position{
					Line:   uint32(block.DefRange().Start.Line),
					Column: uint32(block.DefRange().Start.Column),
				},
				End: hclvet.Position{
					Line:   uint32(block.DefRange().End.Line),
					Column: uint32(block.DefRange().End.Column),
				},
			}

			// Every error we find we construct a "RuleError" struct and add it to our errors list.
			lintErrors = append(lintErrors, hclvet.RuleError{
				Suggestion:  "Use a different resource name than example",
				Remediation: "resource \"google_compute_instance\" \"<new_name>\" {",
				Location:    location,
				Metadata: map[string]string{
					"severity": "warning",
					"example":  "Lorem ipsum dolor sit amet",
				},
			})
		}
	}

	return lintErrors, nil
}


func main() {
	// We instantiate an instance of our check interface we filled out above so we can register
	// it into the rule below.
	newCheck := Check{}

	// Here we can fill out more information about the rule, it's purpose, and where to find more
	// documentation.
	// The documentation for each of these fields can be found looking at the sdk documentation
	// here: https://pkg.go.dev/github.com/clintjedwards/hclvet/sdk#Rule
	newRule := &hclvet.Rule{
		Name:  "{{.Name}}",
		Short: "<Short description on what this rule is for, shown to user whenever rule finds an error>",
		Long: "<A longer description about what this rule is for. This is used as documentation.>",
		Enabled: true,
		Link:    "<This should be a hyperlink to additional documentation>",
		Check:   &newCheck,
	}

	// Lastly we add our new rule so that it is properly registered.
	hclvet.NewRule(newRule)
}
`

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	const mainFileName = "main.go"
	const rulesDirName = "rules"
	ruleDirPath := fmt.Sprintf("%s/%s/%s", currentDir, rulesDirName, name)
	mainFilePath := fmt.Sprintf("%s/%s", ruleDirPath, mainFileName)

	file, err := os.Create(mainFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("").Parse(mainFileContent))
	err = tmpl.Execute(file, struct {
		Name string
	}{
		Name: name,
	})
	if err != nil {
		return err
	}

	return nil
}
