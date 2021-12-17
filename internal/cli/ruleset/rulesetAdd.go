package ruleset

import (
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/clintjedwards/hclvet/internal/cli/appcfg"
	hclvetPlugin "github.com/clintjedwards/hclvet/internal/plugin"
	"github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/clintjedwards/hclvet/internal/utils"
	models "github.com/clintjedwards/hclvet/sdk"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

const (
	golangBinaryName = "go"
)

var cmdRulesetAdd = &cobra.Command{
	Use:   "add <repository>",
	Short: "Retrieves and enables a new ruleset",
	Long: `The add command retrieves and enables a new hclvet ruleset.

Supports a wide range of sources including github url, fileserver path, or even just
the path to a local directory.

See https://github.com/hashicorp/go-getter#url-format for more information on all supported input
types.

Arguments:

• <repository> is the location of the ruleset repository. Ruleset repositories must adhere to the
following rules:

  • Repository must contain a ruleset.hcl file containing name and version.
  • Repository must contain a rules folder with rules plugins built with hclvet sdk.

For more information on hclvet ruleset repository requirements and structure see:
github.com/clintjedwards/hclvet-ruleset-example
`,
	Example: `$ hclvet add github.com/example/hclvet-ruleset-aws
$ hclvet add ~/tmp/hclvet-ruleset-example`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

// moveRepo copies a downloaded repo from the temporary download path
// to the well known repo path for the ruleset.
func moveRepo(ruleset, tmpPath string) error {
	err := utils.CreateDir(appcfg.RulesetPath(ruleset))
	if err != nil {
		return fmt.Errorf("could not create parent directory: %w", err)
	}

	err = copy.Copy(tmpPath, appcfg.RepoPath(ruleset))
	if err != nil {
		return fmt.Errorf("could not copy ruleset to config directory: %w", err)
	}

	return nil
}

// generateHash takes a string and returns it's hashed result.
// The hash used is non-cryptographic(because of speed) and as such it
// should not be used for anything expecting to be secure.
func generateHash(s string) string {
	digest := fnv.New32()
	_, _ = digest.Write([]byte(s))
	hash := hex.EncodeToString(digest.Sum(nil))
	return fmt.Sprintf(hash[0:5])
}

// getRulePluginClient returns the grpc plugin client, the rule plugin client, and a possible error.
// the rule plugin client is used to communicate with the rule plugins and from this client you can
// run commands that work just like regular methods against the plugins.
//
// YOU MUST call kill() on the returned plugin.Client object or it will cause memory leaks.
//
// TODO(clintjedwards): This by default just discards any logs from the client.
// Change this to only do this above the loglevel debug.
func getRulePluginClient(ruleset, ruleID string) (client *plugin.Client, rule hclvetPlugin.RuleDefinition, err error) {
	tmpPluginName := "hclvetPlugin"

	client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: hclvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			tmpPluginName: &hclvetPlugin.HCLvetRulePlugin{},
		},
		Cmd: exec.Command(appcfg.RulePath(ruleset, ruleID)),
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: ioutil.Discard,
			Level:  0,
			Name:   "plugin",
		}),
		Stderr:           nil,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("could not connect to rule plugin %s: %v", ruleID, err)
	}

	raw, err := rpcClient.Dispense(tmpPluginName)
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("could not connect to rule plugin %s: %v", ruleID, err)
	}

	plugin, ok := raw.(hclvetPlugin.RuleDefinition)
	if !ok {
		client.Kill()
		return nil, nil, fmt.Errorf("could not convert rule plugin %s: %v", ruleID, err)
	}

	return client, plugin, nil
}

// getRuleInfo retrieves information by calling the GetRuleInfo method on the rule plugin.
func getRuleInfo(ruleset, ruleID string) (models.Rule, error) {
	c, plugin, err := getRulePluginClient(ruleset, ruleID)
	if err != nil {
		return models.Rule{}, fmt.Errorf("could not get rule info for %s: %w", ruleID, err)
	}
	defer c.Kill()

	response, err := plugin.GetRuleInfo(&proto.GetRuleInfoRequest{})
	if err != nil {
		return models.Rule{}, fmt.Errorf("could not get rule info for %s: %w", ruleID, err)
	}

	return models.Rule{
		ID:      ruleID,
		Name:    response.RuleInfo.Name,
		Short:   response.RuleInfo.Short,
		Long:    response.RuleInfo.Long,
		Link:    response.RuleInfo.Link,
		Enabled: response.RuleInfo.Enabled,
	}, nil
}

// buildRule builds the rule/plugin from srcPath and stores it in dstPath
// with the provided name.
func buildRule(srcPath, dstPath string) ([]byte, error) {
	buildArgs := []string{"build", "-o", dstPath}

	golangBinaryPath, err := exec.LookPath(golangBinaryName)
	if err != nil {
		return nil, err
	}

	// go build <args> <path_to_plugin_src_files>
	output, err := utils.ExecuteCmd(golangBinaryPath, buildArgs, nil, srcPath)
	if err != nil {
		return output, err
	}

	return output, nil
}

func runAdd(cmd *cobra.Command, args []string) error {
	repoLocation := args[0]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	state, err := newState("Adding ruleset", format)
	if err != nil {
		return err
	}

	state.fmt.Print("Checking for duplicates")

	// Check that repository does not yet exist
	if state.cfg.RepositoryExists(repoLocation) {
		errText := "repository already exists; use `hclvet ruleset update`" +
			" to manipulate already added rulesets"
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}

	// Download remote repository
	state.fmt.Print(fmt.Sprintf("Retrieving %s", repoLocation))
	tmpDownloadPath := fmt.Sprintf("%s/hclvet_%s", os.TempDir(), generateHash(repoLocation))
	err = getRemoteRuleset(repoLocation, tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not download ruleset: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}
	defer os.RemoveAll(tmpDownloadPath) // Remove tmp dir in case we end early
	state.fmt.PrintSuccess(fmt.Sprintf("Retrieved %s", repoLocation))

	// Get the repository information from the repository itself.
	info, err := getRemoteRulesetInfo(tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not get ruleset info: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}

	// Verify that the repository has the correct elements.
	state.fmt.Print("Verifying ruleset")
	err = verifyRuleset(tmpDownloadPath, info)
	if err != nil {
		errText := fmt.Sprintf("could not verify ruleset: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}
	state.fmt.PrintSuccess("Verified ruleset")

	// Add new ruleset to configuration file.
	state.fmt.Print("Adding ruleset to config")
	err = state.cfg.AddRuleset(models.Ruleset{
		Name:       info.Name,
		Version:    info.Version,
		Repository: repoLocation,
		Enabled:    true,
	})
	if err != nil {
		errText := fmt.Sprintf("could not add ruleset: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}

	// Move downloaded ruleset repository to the permanent location within config.
	state.fmt.Print("Moving ruleset to permanent config location")
	err = moveRepo(info.Name, tmpDownloadPath)
	if err != nil {
		errText := fmt.Sprintf("could not move ruleset repository: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}
	state.fmt.PrintSuccess("New ruleset added")

	// Find all rules within the ruleset and build them using the go compiler.
	err = buildAllRules(state, info.Name)
	if err != nil {
		errText := fmt.Sprintf("could not build ruleset rules: %v", err)
		state.fmt.PrintErr(errText)
		state.fmt.Finish()
		return errors.New(errText)
	}

	state.fmt.PrintSuccess(fmt.Sprintf("Successfully added ruleset: %s v%s", info.Name, info.Version))
	state.fmt.Finish()
	return nil
}

func init() {
	CmdRuleset.AddCommand(cmdRulesetAdd)
}
