package sdk

import (
	"log"

	hclvetPlugin "github.com/clintjedwards/hclvet/internal/plugin"
	proto "github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// GetRuleInfo returns information about the rule itself.
func (rule *Rule) GetRuleInfo(request *proto.GetRuleInfoRequest) (*proto.GetRuleInfoResponse, error) {
	ruleInfo := proto.GetRuleInfoResponse{
		RuleInfo: &proto.RuleInfo{
			Name:    rule.Name,
			Short:   rule.Short,
			Long:    rule.Long,
			Link:    rule.Link,
			Enabled: rule.Enabled,
		},
	}

	return &ruleInfo, nil
}

// ExecuteRule runs the linting rule given a single file and returns any linting errors.
func (rule *Rule) ExecuteRule(request *proto.ExecuteRuleRequest) (*proto.ExecuteRuleResponse, error) {
	ruleErrors, err := rule.Check.Check(request.HclFile)

	return &proto.ExecuteRuleResponse{
		Errors: ruleErrorsToProto(ruleErrors),
	}, err
}

// ParseHCL parses the HCL file content and returns a simple data structure representing the file.
// It's safe to ignore the error from ParseHCL as it should have already been handled by the main
// process.
func ParseHCL(content []byte) *hclsyntax.Body {
	// TODO(clintjedwards): Having to reparse the file for every plugin is very slow, figure
	// out if there is a better way to transfer this information to the main binary and have
	// plugins consume that instead.
	parser := hclparse.NewParser()
	file, _ := parser.ParseHCL(content, "tmp")
	return file.Body.(*hclsyntax.Body)
}

func ruleErrorsToProto(ruleErrors []RuleError) []*proto.RuleError {
	protoRuleErrors := []*proto.RuleError{}

	for _, ruleError := range ruleErrors {
		protoRuleErrors = append(protoRuleErrors, &proto.RuleError{
			Location: &proto.Location{
				Start: &proto.Position{
					Line:   ruleError.Location.Start.Line,
					Column: ruleError.Location.Start.Column,
				},
				End: &proto.Position{
					Line:   ruleError.Location.End.Line,
					Column: ruleError.Location.End.Column,
				},
			},
			Suggestion:  ruleError.Suggestion,
			Remediation: ruleError.Remediation,
			Metadata:    ruleError.Metadata,
		})
	}

	return protoRuleErrors
}

// validates a new rule has at least the basic information
func (rule *Rule) isValid() bool {
	if rule.Short == "" {
		return false
	}

	if rule.Name == "" {
		return false
	}

	if rule.Check == nil {
		return false
	}

	return true
}

// NewRule registers a new linting rule. This function must be included inside a rule.
func NewRule(rule *Rule) {
	if !rule.isValid() {
		log.Fatalf("%s is not valid", rule.Name)
		return
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: hclvetPlugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			// The key here is to enable different plugins to be served by one binary
			"hclvet-sdk": &hclvetPlugin.HCLvetRulePlugin{Impl: rule},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
