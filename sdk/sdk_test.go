package sdk

import (
	"encoding/json"
	"testing"
)

func TestLintErrorWrapper(t *testing.T) {
	logLine := []byte(`{"data":{"lint_error":{"filepath":"/home/clintjedwards/Documents/hclvet/internal/testdata/test1.tf","line":"resource \"google_compute_instance\" \"example\" {","rule_error":{"suggestion":"Use a different resource name than example","remediation":"resource \"google_compute_instance\" \"\u003cnew_name\u003e\" {","location":{"start":{"line":1,"column":1},"end":{"line":1,"column":47}},"metadata":{"example":"Lorem ipsum dolor sit amet","severity":"warning"}},"rule":{"id":"89cd4","name":"No resource with name example","short":"Example is a poor name for a resource and might lead to naming collisions.","long":"\nThis is simply a test description of a resource that effectively alerts on nothingness.\nIn turn this is essentially a really long description so we can test that our descriptions\nwork properly and are displayed properly in the terminal.\n","link":"https://google.com","enabled":true},"ruleset":"example"}},"label":"error"}`)

	var newError LintErrorWrapper
	err := json.Unmarshal(logLine, &newError)
	if err != nil {
		t.Fatal(err)
	}

	if newError.Data.LintError == nil {
		t.Fatal("LintError which should be an object is nil")
	}
}
