package a2ui

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

// The butter-basic-v1 catalog. Component names and shapes follow A2UI's
// basic catalog where one exists (Column, Row, Text, Card, Divider,
// TextField, ChoicePicker, Button); KeyValue and Status are butter additions
// for result cards. The dashboard renders every component with its own
// design system (front/src/features/agui-chat/a2ui).
//
// Models may only produce the read-only subset validated below. TextField,
// ChoicePicker and Button exist solely in forms the server builds from a
// Human Input node, so a model can neither collect input nor define an
// action.

type propKind int

const (
	// propDynString is a literal string or a {"path": "/…"} data binding.
	propDynString propKind = iota
	propEnum
	propChild
	propChildren
)

type propSpec struct {
	kind     propKind
	required bool
	enum     []string
}

type componentSpec struct {
	props map[string]propSpec
	// modelAllowed marks the read-only subset render_ui accepts.
	modelAllowed bool
}

var (
	justifyValues = []string{"start", "center", "end", "spaceBetween", "spaceAround", "spaceEvenly", "stretch"}
	alignValues   = []string{"start", "center", "end", "stretch"}
)

var catalog = map[string]componentSpec{
	"Column": {modelAllowed: true, props: map[string]propSpec{
		"children": {kind: propChildren, required: true},
		"justify":  {kind: propEnum, enum: justifyValues},
		"align":    {kind: propEnum, enum: alignValues},
	}},
	"Row": {modelAllowed: true, props: map[string]propSpec{
		"children": {kind: propChildren, required: true},
		"justify":  {kind: propEnum, enum: justifyValues},
		"align":    {kind: propEnum, enum: alignValues},
	}},
	"Text": {modelAllowed: true, props: map[string]propSpec{
		"text":    {kind: propDynString, required: true},
		"variant": {kind: propEnum, enum: []string{"h1", "h2", "h3", "h4", "h5", "caption", "body"}},
	}},
	"Card": {modelAllowed: true, props: map[string]propSpec{
		"child": {kind: propChild, required: true},
	}},
	"Divider": {modelAllowed: true, props: map[string]propSpec{
		"axis": {kind: propEnum, enum: []string{"horizontal", "vertical"}},
	}},
	"KeyValue": {modelAllowed: true, props: map[string]propSpec{
		"label": {kind: propDynString, required: true},
		"value": {kind: propDynString, required: true},
	}},
	"Status": {modelAllowed: true, props: map[string]propSpec{
		"text": {kind: propDynString, required: true},
		"tone": {kind: propEnum, enum: []string{"neutral", "info", "success", "warning", "error"}},
	}},
	// Form components: built by the server only (Form.Envelopes), so they
	// carry no property rules here — a model is refused them by name.
	"TextField":    {},
	"ChoicePicker": {},
	"Button":       {},
}

// ModelComponentNames lists the read-only components a model may render.
func ModelComponentNames() []string {
	var names []string
	for name, spec := range catalog {
		if spec.modelAllowed {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

const (
	maxComponentIDLen = 64
	maxLiteralRunes   = 4000
)

var (
	componentIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	// htmlTagPattern spots raw markup: the renderer shows text as text, but
	// the catalog forbids HTML outright rather than rely on that.
	htmlTagPattern = regexp.MustCompile(`<\s*/?\s*[A-Za-z!][^>]*>`)
	// urlPattern spots URLs. No component takes a URL, and cards carry none
	// in their text either: links belong in the model's text answer.
	urlPattern = regexp.MustCompile(`(?i)\b[a-z][a-z0-9+.-]*://|\bjavascript:`)
)

// validateModelComponent checks one component a model sent: a known
// read-only component with only its own, well-typed properties.
func validateModelComponent(c Component) error {
	id, _ := c["id"].(string)
	if id == "" {
		return fmt.Errorf("every component needs a string \"id\"")
	}
	if len(id) > maxComponentIDLen || !componentIDPattern.MatchString(id) {
		return fmt.Errorf("component id %q must be 1-%d letters, digits, '-' or '_'", id, maxComponentIDLen)
	}
	name, _ := c["component"].(string)
	spec, known := catalog[name]
	if !known {
		return fmt.Errorf("component %q: unknown component %q; use one of %s", id, name, strings.Join(ModelComponentNames(), ", "))
	}
	if !spec.modelAllowed {
		return fmt.Errorf("component %q: %s is not available to render_ui; cards are read-only (use one of %s)", id, name, strings.Join(ModelComponentNames(), ", "))
	}
	for key := range c {
		if key == "id" || key == "component" || key == "weight" {
			continue
		}
		if _, ok := spec.props[key]; !ok {
			return fmt.Errorf("component %q: %s has no property %q", id, name, key)
		}
	}
	if w, ok := c["weight"]; ok {
		if _, isNum := w.(float64); !isNum {
			return fmt.Errorf("component %q: weight must be a number", id)
		}
	}
	for key, prop := range spec.props {
		v, present := c[key]
		if !present {
			if prop.required {
				return fmt.Errorf("component %q: %s requires %q", id, name, key)
			}
			continue
		}
		if err := validateProp(prop, v); err != nil {
			return fmt.Errorf("component %q: %s.%s %w", id, name, key, err)
		}
	}
	return nil
}

func validateProp(prop propSpec, v any) error {
	switch prop.kind {
	case propDynString:
		return validateDynString(v)
	case propEnum:
		s, ok := v.(string)
		if !ok || !slices.Contains(prop.enum, s) {
			return fmt.Errorf("must be one of %s", strings.Join(prop.enum, ", "))
		}
	case propChild:
		if s, ok := v.(string); !ok || s == "" {
			return fmt.Errorf("must be the id of a component")
		}
	case propChildren:
		list, ok := v.([]any)
		if !ok {
			return fmt.Errorf("must be an array of component ids (templates are not supported)")
		}
		for _, item := range list {
			if s, ok := item.(string); !ok || s == "" {
				return fmt.Errorf("must be an array of component ids (templates are not supported)")
			}
		}
	default:
		return fmt.Errorf("is not supported")
	}
	return nil
}

// validateDynString accepts a literal string or a data-model path binding.
// Function calls are not part of this catalog.
func validateDynString(v any) error {
	switch value := v.(type) {
	case string:
		return validateLiteral(value)
	case map[string]any:
		path, ok := value["path"].(string)
		if !ok || len(value) != 1 {
			return fmt.Errorf("must be a string or {\"path\": \"/…\"}")
		}
		if _, err := parsePointer(path); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("must be a string or {\"path\": \"/…\"}")
	}
}

func validateLiteral(s string) error {
	if utf8.RuneCountInString(s) > maxLiteralRunes {
		return fmt.Errorf("is longer than %d characters", maxLiteralRunes)
	}
	if htmlTagPattern.MatchString(s) {
		return fmt.Errorf("must be plain text; HTML is not allowed")
	}
	if urlPattern.MatchString(s) {
		return fmt.Errorf("must not contain a URL; put links in your text answer instead")
	}
	return nil
}

// childRefs returns the component IDs c references as children.
func childRefs(c Component) []string {
	spec := catalog[stringProp(c, "component")]
	var refs []string
	for key, prop := range spec.props {
		switch prop.kind {
		case propChild:
			if s, ok := c[key].(string); ok {
				refs = append(refs, s)
			}
		case propChildren:
			if list, ok := c[key].([]any); ok {
				for _, item := range list {
					if s, ok := item.(string); ok {
						refs = append(refs, s)
					}
				}
			}
		}
	}
	sort.Strings(refs)
	return refs
}

func stringProp(c Component, key string) string {
	s, _ := c[key].(string)
	return s
}
