// Package tooltest checks the function declarations ADK tools send to a
// model.
package tooltest

import (
	"encoding/json"
	"testing"

	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

// RequireParamDescriptions fails t unless every parameter of every tool
// carries a description in the declaration the model sees. ADK infers a
// function tool's parameter schema with jsonschema-go, which reads a field's
// description from its `jsonschema` struct tag only and silently drops any
// other tag.
func RequireParamDescriptions(t testing.TB, tools ...tool.Tool) {
	t.Helper()
	if len(tools) == 0 {
		t.Fatal("no tools to check")
	}
	for _, tl := range tools {
		fn, ok := tl.(interface {
			Declaration() *genai.FunctionDeclaration
		})
		if !ok {
			t.Errorf("%s: not a function tool", tl.Name())
			continue
		}
		raw, err := json.Marshal(fn.Declaration().ParametersJsonSchema)
		if err != nil {
			t.Errorf("%s: encode parameter schema: %v", tl.Name(), err)
			continue
		}
		var schema struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Errorf("%s: decode parameter schema: %v", tl.Name(), err)
			continue
		}
		if len(schema.Properties) == 0 {
			t.Errorf("%s: declares no parameters to check", tl.Name())
		}
		for name, prop := range schema.Properties {
			if prop.Description == "" {
				t.Errorf("%s: parameter %q has no description", tl.Name(), name)
			}
		}
	}
}
