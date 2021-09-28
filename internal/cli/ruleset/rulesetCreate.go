package ruleset

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/clintjedwards/hclvet/internal/utils"
	"github.com/spf13/cobra"
)

// cmdRulesetCreate creates a skeleton ruleset
var cmdRulesetCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new ruleset",
	Long: `Creates the files and folders needed to create a new ruleset.

Navigate to the folder in which you mean to create the ruleset. From there simply run this command
to create all files and folders required for a hclvet ruleset.
`,
	Example: `$ hclvet ruleset create example`,
	RunE:    runCreate,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetCreate)
}

func runCreate(_ *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	// Create ./ruleset.hcl
	err := createRulesetConfigFile(name)
	if err != nil {
		log.Println(err)
		return err
	}

	// Create ./README.md
	err = createReadmeFile(name)
	if err != nil {
		log.Println(err)
		return err
	}

	// Create ./rules dir
	err = createRulesDir()
	if err != nil {
		return err
	}

	return nil
}

func createRulesetConfigFile(name string) error {
	const rulesetFileContent = `// short (between 3 and 20 char) name for the ruleset
name = "{{.Name}}"
// bumping the version causes downstream clients to detect that there has been an update.
version = "0.0.0"
`

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	//TODO(clintjedwards): Take this from the appcfg package and stop declaring it everywhere
	const rulesetFileName = "ruleset.hcl"
	rulesetFilePath := fmt.Sprintf("%s/%s", currentDir, rulesetFileName)

	file, err := os.Create(rulesetFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("").Parse(rulesetFileContent))
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

func createReadmeFile(name string) error {
	const readmeFileContent = `# New HCLvet ruleset created

## You're ready to start creating rules!

<br />

## What's next?

1) Look [here](https://github.com/clintjedwards/hclvet-ruleset-example)
to get an idea on how to start creating your own rules.
2) When you're ready, you can use the ` + "`hclvet rule create <rule name>`" +
		` command to generate the template for your rule (or just copy the example).
3) Update the current version and check the name given in the ` + "`ruleset.hcl`" + " file."

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	const readmeFileName = "README.md"
	readmeFilePath := fmt.Sprintf("%s/%s", currentDir, readmeFileName)

	file, err := os.Create(readmeFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("").Parse(readmeFileContent))
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

func createRulesDir() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	const rulesDirName = "rules"
	rulesDirPath := fmt.Sprintf("%s/%s", currentDir, rulesDirName)

	err = utils.CreateDir(rulesDirPath)
	if err != nil {
		return err
	}

	return nil
}
