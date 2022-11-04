// package sdk is a convenience package enabling the easy downstream development around hclvet.
// It provides the primitives to allow for ruleset/rule creation and structs to help in parsing hclvet output.
package sdk

import "github.com/clintjedwards/hclvet/internal/plugin/proto"

// Ruleset represents a packaged set of rules that govern what hclvet checks for.
type Ruleset struct {
	Name       string `hcl:"name,label" json:"name"`
	Version    string `hcl:"version" json:"version"`
	Repository string `hcl:"repository" json:"repository"`
	Enabled    bool   `hcl:"enabled" json:"enabled"`
	Rules      []Rule `hcl:"rule,block" json:"rules"`
}

// Check provides an interface for the user to define their own check/lint method.
// This is the core of the pluggable interface pattern and allows the user to simply consume
// the hcl file and return linting errors.
//
// content is the full hclfile in byte format.
type Check interface {
	Check(content []byte) ([]RuleError, error)
}

// Rule is the representation of a single rule within hclvet.
// This just combines the rule with the check interface.
// This should be kept in lockstep with the Rule model from the hclvet package.
type Rule struct {
	// ID is used by the main hclvet program to uniquely identify rules. Should not be set if creating a rule.
	ID string `hcl:"id,label" json:"id"`
	// The name of the rule, it should be short and to the point of what the rule is for.
	Name string `hcl:"name" json:"name"`
	// A short description about the rule. This should be one line at most and will be shown
	// to the user when the rule finds errors.
	Short string `hcl:"short" json:"short"`
	// A longer description about the rule. This can be looked up by the user via command line.
	Long string `hcl:"long" json:"long"`
	// A link that pertains to the rule; usually additional documentation.
	Link string `hcl:"link" json:"link"`
	// Enabled controls whether the rule will be enabled by default on addition of a ruleset.
	// If enabled is set to false, the user will have to manually turn on the rule.
	Enabled bool `hcl:"enabled" json:"enabled"`
	// Check is a function which runs when the rule is called. This should contain the logic around
	// what the rule is checking.
	Check `json:"-"`
}

// Position represents location within a document.
type Position struct {
	// These are uint32 because that is what the protobuf requires
	Line   uint32 `json:"line"`
	Column uint32 `json:"column"`
}

// Range represents the starting and ending points on a specific line within a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// RuleError represents a single lint error's details
type RuleError struct {
	// Suggestion is a short text description on how to fix the error.
	Suggestion string `json:"suggestion"`
	// Remediation is a short snippet of code that can be used to fix the error.
	Remediation string `json:"remediation"`
	// The location of the error in the file.
	Location Range `json:"location"`
	// metadata is a key value store that allows the rule to include extra data,
	// that can be used by any tooling consuming said rule. For example "severity"
	// might be something included in metadata.
	Metadata map[string]string `json:"metadata"`
}

// LintErrorWrapper is a convenience struct so that json output is easier to programmatically read.
// Nesting the output of LintError as a pointer allows downstream programs to check if the line
// parses cleanly into the wrapper by simply checking if the resulting object is nil
// Example:
//
// A line that is not a LintError
// err := json.Unmarshal(logLine, &newError)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
// if newError.Data.LintError == nil {
// We know this is not a LintError because this is nil
// }
type LintErrorWrapper struct {
	Label string `json:"label"`
	Data  struct {
		LintError *LintError `json:"lint_error"`
	} `json:"data"`
}

// LintError is a harness for all the details that go into a lint error
type LintError struct {
	Filepath string    `json:"filepath"`
	Line     string    `json:"line"`
	RuleErr  RuleError `json:"rule_error"`
	Rule     Rule      `json:"rule"`
	Ruleset  string    `json:"ruleset"`
}

func ProtoToRuleError(proto *proto.RuleError) *RuleError {
	re := &RuleError{}
	re.Suggestion = proto.Suggestion
	re.Remediation = proto.Remediation
	re.Metadata = proto.Metadata
	re.Location = Range{
		Start: Position{
			Line:   proto.Location.Start.Line,
			Column: proto.Location.Start.Column,
		},
		End: Position{
			Line:   proto.Location.End.Line,
			Column: proto.Location.End.Column,
		},
	}
	return re
}
