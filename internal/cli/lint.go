package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/clintjedwards/hclvet/internal/cli/appcfg"
	hclvetPlugin "github.com/clintjedwards/hclvet/internal/plugin"
	"github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/clintjedwards/hclvet/internal/utils"
	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/clintjedwards/polyfmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/mitchellh/go-homedir"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

// cmdLint is a subcommand that controls the actual act of running the linter
var cmdLint = &cobra.Command{
	Use:   "lint [paths...]",
	Short: "Runs the hcl linter",
	Long: `Runs the hcl linter for all enabled rules, grabbing all hcl files in current
directory by default.

Accepts multiple paths delimited by a space.
`,
	RunE: runLint,
	Example: `$ hclvet lint
$ hclvet lint myfile.tf
$ hclvet line somefile.tf manyfilesfolder/*`,
}

// state contains a bunch of useful state information for the add cli function. This is mostly
// just for convenience.
type state struct {
	fmt polyfmt.Formatter
	cfg *appcfg.Appcfg
}

// newState returns a new state object with the fmt initialized
func newState(initialFmtMsg, format string) (*state, error) {
	clifmt, err := polyfmt.NewFormatter(polyfmt.Mode(format))
	if err != nil {
		return nil, err
	}

	clifmt.Print(initialFmtMsg, polyfmt.Pretty)

	cfg, err := appcfg.GetConfig()
	if err != nil {
		errText := fmt.Sprintf("error reading config file %q: %v", appcfg.ConfigFilePath(), err)
		clifmt.PrintErr(errText)
		return nil, errors.New(errText)
	}

	return &state{
		fmt: clifmt,
		cfg: cfg,
	}, nil
}

// getHCLFiles returns the paths of all hcl files within the path given.
// If the given path is not a directory, but a hcl file instead, it will return a list with
// only that file included.
func (s *state) getHCLFiles(paths []string) ([]string, error) {

	tfFiles := []string{}

	for _, path := range paths {
		// Resolve home directory
		path, err := homedir.Expand(path)
		if err != nil {
			errText := fmt.Sprintf("could not parse path %s", path)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return nil, errors.New(errText)
		}

		// Get full path for file
		path, err = filepath.Abs(path)
		if err != nil {
			errText := fmt.Sprintf("could not parse path %s", path)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return nil, errors.New(errText)
		}

		// Check that the path exists
		_, err = os.Stat(filepath.Dir(path))
		if err != nil {
			errText := fmt.Sprintf("could not open path: %v", err)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return nil, errors.New(errText)
		}

		// Return all hcl files
		globFiles, err := filepath.Glob(path)
		if err != nil {
			errText := fmt.Sprintf("could match on glob pattern %s", path)
			s.fmt.PrintErr(errText)
			s.fmt.Finish()
			return nil, errors.New(errText)
		}

		for _, file := range globFiles {
			if strings.HasSuffix(file, ".tf") {
				tfFiles = append(tfFiles, file)
			}
		}
	}

	return tfFiles, nil
}

func runLint(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Print(err)
		return err
	}

	state, err := newState("Running Linter", format)
	if err != nil {
		log.Print(err)
		return err
	}

	// Get paths from arguments, if no arguments were given attempt to get files from current dir.
	var paths []string
	if len(args) == 0 {
		defaultPath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return err
		}
		defaultPath = defaultPath + "/*"

		paths = []string{defaultPath}
	} else {
		paths = args
	}

	files, err := state.getHCLFiles(paths)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		state.fmt.PrintErr("No hcl files found")
		state.fmt.Finish()
		return errors.New("no hcl files found")
	}

	startTime := time.Now()
	numFiles := 0   // how many files we've ran through
	numErrors := 0  // how many errors we've found
	numSkipped := 0 // how many files we've skipped

	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			state.fmt.PrintErr(
				fmt.Sprintf("Skipped file %s; could not open: %v\n", filepath.Base(file.Name()), err),
				polyfmt.Pretty)
			state.fmt.PrintErr(map[string]interface{}{
				"skipped_file": fmt.Sprintf("Skipped file %s; could not open: %v\n", filepath.Base(file.Name()), err),
			}, polyfmt.JSON)
			numSkipped++
			continue
		}

		errorsFound, err := state.lintFile(file)
		if err != nil {
			file.Close() // Close the file handle if we're not going to process it.
			state.fmt.PrintErr(
				fmt.Sprintf("Skipped file %s; could not open: %v\n", filepath.Base(file.Name()), err),
				polyfmt.Pretty)
			state.fmt.PrintErr(map[string]interface{}{
				"skipped_file": fmt.Sprintf("Skipped file %s; could not open: %v\n", filepath.Base(file.Name()), err),
			}, polyfmt.JSON)
			numSkipped++
			continue
		}
		numErrors = numErrors + errorsFound
		numFiles++
		file.Close() // don't defer things that are in loops
	}

	duration := time.Since(startTime)
	durationSeconds := float64(duration) / float64(time.Second)
	timePerFile := float64(duration) / float64(numFiles)

	state.fmt.PrintSuccess(fmt.Sprintf("Found %d error(s) and skipped %d file(s)", numErrors, numSkipped))
	state.fmt.PrintSuccess(fmt.Sprintf("Linted %d file(s) in %.2fs (avg %.2fms/file)",
		numFiles, durationSeconds, timePerFile/float64(time.Millisecond)))
	state.fmt.Finish()

	return nil
}

// lintFile orchestrates the process of linting the given file.
func (s *state) lintFile(file *os.File) (int, error) {

	// Check we have enough memory to store file
	err := checkAvailMemory(file)
	if err != nil {
		return 0, err
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}

	_, diags := hclparse.NewParser().ParseHCL(contents, file.Name())
	if diags.HasErrors() {
		return 0, diags
	}

	rulesets := s.cfg.Rulesets
	errorsFound := 0

	// For each ruleset we need to run each one of the enabled rules against the given file.
	for _, ruleset := range rulesets {
		if !ruleset.Enabled {
			continue
		}

		for _, rule := range ruleset.Rules {
			if !rule.Enabled {
				continue
			}

			s.fmt.Print(fmt.Sprintf("%q ruleset linting %q for rule %q",
				strings.ToLower(ruleset.Name), filepath.Base(file.Name()), strings.ToLower(rule.Name)))

			numErrors, err := s.runRule(ruleset.Name, rule, file.Name(), contents)
			if err != nil {
				s.fmt.PrintErr(fmt.Sprintf("Rule failed %s; encountered an error while running: %v",
					rule.Name, err))
				continue
			}

			errorsFound = errorsFound + numErrors
		}
	}

	return errorsFound, nil
}

// runRule runs the rule plugin and returns the number of errors found.
func (s *state) runRule(ruleset string, rule models.Rule, filepath string, rawHCLFile []byte) (int, error) {
	tmpPluginName := "hclvetPlugin"

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: hclvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			tmpPluginName: &hclvetPlugin.HCLvetRulePlugin{},
		},
		Cmd: exec.Command(appcfg.RulePath(ruleset, rule.ID)),
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: ioutil.Discard,
			Level:  0,
			Name:   "plugin",
		}),
		Stderr:           nil,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return 0, fmt.Errorf("could not create rpc client: %w", err)
	}

	raw, err := rpcClient.Dispense(tmpPluginName)
	if err != nil {
		return 0, fmt.Errorf("could not connect to rule plugin: %w", err)
	}

	plugin, ok := raw.(hclvetPlugin.RuleDefinition)
	if !ok {
		return 0, fmt.Errorf("could not convert rule interface: %w", err)
	}

	response, err := plugin.ExecuteRule(&proto.ExecuteRuleRequest{
		HclFile: rawHCLFile,
	})
	if err != nil {
		return 0, fmt.Errorf("could not execute linting rule: %w", err)
	}

	ruleErrs := response.Errors
	for _, ruleError := range ruleErrs {
		line, _, err := utils.ReadLine(bytes.NewBuffer(rawHCLFile), int(ruleError.Location.Start.Line))
		if err != nil {
			return 0, fmt.Errorf("could not get line from file: %w", err)
		}

		s.fmt.PrintErr(formatLintError(models.LintError{
			Filepath: filepath,
			Line:     line,
			Ruleset:  ruleset,
			Rule:     rule,
			RuleErr:  *models.ProtoToRuleError(ruleError),
		})+"\n", polyfmt.Pretty)

		s.fmt.PrintErr(struct {
			LintError models.LintError `json:"lint_error"`
		}{
			LintError: models.LintError{
				Filepath: filepath,
				Line:     line,
				Ruleset:  ruleset,
				Rule:     rule,
				RuleErr:  *models.ProtoToRuleError(ruleError),
			},
		}, polyfmt.JSON)
	}

	return len(ruleErrs), nil
}

// checkAvailMemory compares the file size of a given file vs the available
// memory of the OS. If the OS does not have enough memory to read the
// file entirely this will return an error.
func checkAvailMemory(file *os.File) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	fileInfo, _ := file.Stat()
	if v.Available < uint64(fileInfo.Size()) {
		return fmt.Errorf("not enough available memory. avail: %d; file: %d",
			v.Available, fileInfo.Size())
	}

	return nil
}

func init() {
	RootCmd.AddCommand(cmdLint)
}
