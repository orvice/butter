package a2ui

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/interrupt"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Form limits (issue #350).
const (
	MaxFormFields       = 20
	MaxTextLength       = 2000
	MaxChoiceOptions    = 50
	maxFieldNameLen     = 64
	maxLabelRunes       = 200
	maxHintRunes        = 500
	maxFormTitleRunes   = 200
	formMetadataKey     = "butter_a2ui_form"
	formSubmitEventName = "butter.submitForm"
	formRevision        = 1
	// AnsweredRevision is the revision of a form's answered marker.
	AnsweredRevision = 2
)

// FieldType is how one form field is answered.
type FieldType string

const (
	FieldText         FieldType = "text"
	FieldSingleChoice FieldType = "single_choice"
)

// Option is one fixed choice of a single-choice field.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Field is one form field, frozen as it was configured when the form was
// shown.
type Field struct {
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	Hint      string    `json:"hint,omitempty"`
	Required  bool      `json:"required,omitempty"`
	Type      FieldType `json:"type"`
	MaxLength int       `json:"maxLength,omitempty"`
	Options   []Option  `json:"options,omitempty"`
}

var fieldNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// HasForm reports whether a node's form config asks for a form at all. An
// empty config keeps the plain-text question.
func HasForm(cfg *agentsv1.HumanInputForm) bool {
	return cfg != nil && (strings.TrimSpace(cfg.GetTitle()) != "" || len(cfg.GetFields()) > 0)
}

// ValidateFormConfig checks a Human Input node's form config. It runs when
// the agent is saved and again when the Workflow Agent is built, so an
// invalid form never reaches the runtime.
func ValidateFormConfig(cfg *agentsv1.HumanInputForm) error {
	if !HasForm(cfg) {
		return nil
	}
	if utf8.RuneCountInString(cfg.GetTitle()) > maxFormTitleRunes {
		return fmt.Errorf("form title is longer than %d characters", maxFormTitleRunes)
	}
	fields := cfg.GetFields()
	if len(fields) == 0 {
		return errors.New("a form needs at least one field")
	}
	if len(fields) > MaxFormFields {
		return fmt.Errorf("a form may have at most %d fields, got %d", MaxFormFields, len(fields))
	}
	seen := make(map[string]bool, len(fields))
	for i, f := range fields {
		name := f.GetName()
		where := fmt.Sprintf("form field %d", i+1)
		if name != "" {
			where = fmt.Sprintf("form field %q", name)
		}
		if name == "" {
			return fmt.Errorf("%s: name is required", where)
		}
		if len(name) > maxFieldNameLen || !fieldNamePattern.MatchString(name) {
			return fmt.Errorf("%s: name must be up to %d letters, digits or underscores, starting with a letter or underscore", where, maxFieldNameLen)
		}
		if seen[name] {
			return fmt.Errorf("%s: duplicate field name", where)
		}
		seen[name] = true
		if strings.TrimSpace(f.GetLabel()) == "" {
			return fmt.Errorf("%s: label is required", where)
		}
		if utf8.RuneCountInString(f.GetLabel()) > maxLabelRunes {
			return fmt.Errorf("%s: label is longer than %d characters", where, maxLabelRunes)
		}
		if utf8.RuneCountInString(f.GetHint()) > maxHintRunes {
			return fmt.Errorf("%s: hint is longer than %d characters", where, maxHintRunes)
		}
		switch f.GetType() {
		case agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT:
			if f.GetMaxLength() < 0 || f.GetMaxLength() > MaxTextLength {
				return fmt.Errorf("%s: max_length must be between 0 and %d", where, MaxTextLength)
			}
			if len(f.GetOptions()) > 0 {
				return fmt.Errorf("%s: a text field has no options", where)
			}
		case agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE:
			if f.GetMaxLength() != 0 {
				return fmt.Errorf("%s: max_length applies to text fields only", where)
			}
			if err := validateOptions(f.GetOptions()); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		default:
			return fmt.Errorf("%s: type must be TEXT or SINGLE_CHOICE", where)
		}
	}
	return nil
}

func validateOptions(options []*agentsv1.HumanInputFormOption) error {
	if len(options) == 0 {
		return errors.New("a single-choice field needs at least one option")
	}
	if len(options) > MaxChoiceOptions {
		return fmt.Errorf("a single-choice field may have at most %d options, got %d", MaxChoiceOptions, len(options))
	}
	seen := make(map[string]bool, len(options))
	for i, o := range options {
		if strings.TrimSpace(o.GetValue()) == "" {
			return fmt.Errorf("option %d: value is required", i+1)
		}
		if utf8.RuneCountInString(o.GetValue()) > maxLabelRunes {
			return fmt.Errorf("option %q: value is longer than %d characters", o.GetValue(), maxLabelRunes)
		}
		if seen[o.GetValue()] {
			return fmt.Errorf("option %q: duplicate value", o.GetValue())
		}
		seen[o.GetValue()] = true
		if strings.TrimSpace(o.GetLabel()) == "" {
			return fmt.Errorf("option %q: label is required", o.GetValue())
		}
		if utf8.RuneCountInString(o.GetLabel()) > maxLabelRunes {
			return fmt.Errorf("option %q: label is longer than %d characters", o.GetValue(), maxLabelRunes)
		}
	}
	return nil
}

// FieldsFromConfig converts a validated form config into frozen fields.
func FieldsFromConfig(cfg *agentsv1.HumanInputForm) []Field {
	fields := make([]Field, 0, len(cfg.GetFields()))
	for _, f := range cfg.GetFields() {
		field := Field{
			Name:     f.GetName(),
			Label:    f.GetLabel(),
			Hint:     f.GetHint(),
			Required: f.GetRequired(),
		}
		switch f.GetType() {
		case agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE:
			field.Type = FieldSingleChoice
			for _, o := range f.GetOptions() {
				field.Options = append(field.Options, Option{Value: o.GetValue(), Label: o.GetLabel()})
			}
		default:
			field.Type = FieldText
			field.MaxLength = int(f.GetMaxLength())
			if field.MaxLength == 0 {
				field.MaxLength = MaxTextLength
			}
		}
		fields = append(fields, field)
	}
	return fields
}

// Instructions describes the fields in plain text. It follows the question
// everywhere the form cannot render (Telegram, the classic chat, cron
// notifications, AG-UI clients without A2UI), so a person can still answer
// in words or as JSON text; the form's rules bind only form submissions.
func Instructions(fields []Field) string {
	var b strings.Builder
	b.WriteString("Please answer these fields (plain text, or a JSON object with these keys):")
	for _, f := range fields {
		b.WriteString("\n- ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(f.Label)
		var notes []string
		if f.Required {
			notes = append(notes, "required")
		} else {
			notes = append(notes, "optional")
		}
		switch f.Type {
		case FieldSingleChoice:
			choices := make([]string, 0, len(f.Options))
			for _, o := range f.Options {
				if o.Label != "" && o.Label != o.Value {
					choices = append(choices, fmt.Sprintf("%s (%s)", o.Value, o.Label))
				} else {
					choices = append(choices, o.Value)
				}
			}
			notes = append(notes, "one of: "+strings.Join(choices, ", "))
		case FieldText:
			notes = append(notes, fmt.Sprintf("up to %d characters", f.MaxLength))
		}
		b.WriteString(" (")
		b.WriteString(strings.Join(notes, "; "))
		b.WriteString(")")
		if f.Hint != "" {
			b.WriteString(" — ")
			b.WriteString(f.Hint)
		}
	}
	return b.String()
}

// Form binds one server-built form surface to the Interrupt it answers. It
// is created when a Human Input node pauses and persisted with that
// request-input event, so the surface, submit token, revision and Interrupt
// are fixed by the server — neither the model nor the client can choose
// them — and later config edits never change the rules of a form already
// shown.
type Form struct {
	SurfaceID   string  `json:"surface_id"`
	Token       string  `json:"token"`
	Revision    int     `json:"revision"`
	InterruptID string  `json:"interrupt_id"`
	Title       string  `json:"title"`
	Question    string  `json:"question"`
	Fields      []Field `json:"fields"`
}

// NewForm builds the binding for the Interrupt interruptID. question is the
// node's own question, without instructions.
func NewForm(interruptID, question string, cfg *agentsv1.HumanInputForm) Form {
	title := strings.TrimSpace(cfg.GetTitle())
	if title == "" {
		title = question
	}
	return Form{
		SurfaceID:   newSurfaceID("form"),
		Token:       strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", ""),
		Revision:    formRevision,
		InterruptID: interruptID,
		Title:       title,
		Question:    question,
		Fields:      FieldsFromConfig(cfg),
	}
}

// Attach persists f with the request-input event that opens its Interrupt.
// It rides CustomMetadata, which is stored with the event and never sent to
// a model.
func (f Form) Attach(ev *session.Event) {
	raw, _ := json.Marshal(f)
	if ev.CustomMetadata == nil {
		ev.CustomMetadata = map[string]any{}
	}
	ev.CustomMetadata[formMetadataKey] = string(raw)
}

// FormOf reads the form binding persisted on a request-input event.
func FormOf(ev *session.Event) (Form, bool) {
	if ev == nil || ev.CustomMetadata == nil {
		return Form{}, false
	}
	s, _ := ev.CustomMetadata[formMetadataKey].(string)
	var f Form
	if s == "" || json.Unmarshal([]byte(s), &f) != nil || f.SurfaceID == "" || f.InterruptID == "" {
		return Form{}, false
	}
	return f, true
}

// PendingForms returns the forms whose Interrupt is still unanswered,
// oldest first. Pending state is derived from session events alone
// (ADR-0002); the form metadata only adds presentation to an Interrupt
// that interrupt.Pending already reports.
func PendingForms(sess session.Session) []Form {
	pending := interrupt.Pending(sess)
	if len(pending) == 0 {
		return nil
	}
	forms := formsByInterrupt(sess)
	var out []Form
	for _, p := range pending {
		if f, ok := forms[p.InterruptID]; ok {
			out = append(out, f)
		}
	}
	return out
}

func formsByInterrupt(sess session.Session) map[string]Form {
	forms := map[string]Form{}
	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		if f, ok := FormOf(events.At(i)); ok {
			forms[f.InterruptID] = f
		}
	}
	return forms
}

// FormView is the client's view of a form binding: enough to render the
// fields and address a submission, and nothing a client could use to
// retarget one.
type FormView struct {
	InterruptID string  `json:"interruptId"`
	Token       string  `json:"token"`
	Revision    int     `json:"revision"`
	Title       string  `json:"title"`
	Question    string  `json:"question"`
	Fields      []Field `json:"fields"`
}

// View returns the client view of f.
func (f Form) View() *FormView {
	return &FormView{
		InterruptID: f.InterruptID,
		Token:       f.Token,
		Revision:    f.Revision,
		Title:       f.Title,
		Question:    f.Question,
		Fields:      f.Fields,
	}
}

// Fallback is the readable text of the form for clients that cannot render
// it: the question and the field instructions.
func (f Form) Fallback() string {
	return f.Question + "\n\n" + Instructions(f.Fields)
}

// ReadableAnswer renders an answer to f the way the dashboard shows a form
// submission in the conversation: the title, then "Label: value" per field,
// with a choice's label for its value and "—" for an empty value. An answer
// that is not a JSON object, such as a plain-text reply, is returned as is.
func (f Form) ReadableAnswer(answer string) string {
	var values map[string]any
	if json.Unmarshal([]byte(answer), &values) != nil {
		return answer
	}
	lines := []string{f.Title}
	for _, field := range f.Fields {
		value, _ := values[field.Name].(string)
		lines = append(lines, field.Label+": "+field.displayValue(value))
	}
	return strings.Join(lines, "\n")
}

func (field Field) displayValue(value string) string {
	if value == "" {
		return "—"
	}
	if field.Type == FieldSingleChoice {
		for _, o := range field.Options {
			if o.Value == value {
				return o.Label
			}
		}
	}
	return value
}

// Envelopes returns the messages that build the form surface: a card with
// the title, the question, one input per field, and the submit button. The
// data model holds only the client's local draft.
func (f Form) Envelopes() []Envelope {
	children := []any{"title"}
	components := []Component{
		{"id": "root", "component": "Card", "child": "form"},
		{"id": "title", "component": "Text", "text": f.Title, "variant": "h3"},
	}
	if f.Question != "" && f.Question != f.Title {
		children = append(children, "question")
		components = append(components, Component{"id": "question", "component": "Text", "text": f.Question})
	}
	values := map[string]any{}
	for _, field := range f.Fields {
		id := "field_" + field.Name
		children = append(children, id)
		c := Component{
			"id":    id,
			"name":  field.Name,
			"label": field.Label,
			"value": map[string]any{"path": "/values/" + field.Name},
		}
		if field.Hint != "" {
			c["hint"] = field.Hint
		}
		if field.Required {
			c["required"] = true
		}
		switch field.Type {
		case FieldSingleChoice:
			c["component"] = "ChoicePicker"
			c["variant"] = "mutuallyExclusive"
			options := make([]any, 0, len(field.Options))
			for _, o := range field.Options {
				options = append(options, map[string]any{"label": o.Label, "value": o.Value})
			}
			c["options"] = options
			values[field.Name] = []any{}
		default:
			c["component"] = "TextField"
			c["variant"] = "shortText"
			if field.MaxLength > 200 {
				c["variant"] = "longText"
			}
			c["maxLength"] = field.MaxLength
			values[field.Name] = ""
		}
		components = append(components, c)
	}
	children = append(children, "submit")
	components = append(components,
		Component{
			"id": "submit", "component": "Button", "child": "submit_label", "variant": "primary",
			"action": map[string]any{"event": map[string]any{
				"name":    formSubmitEventName,
				"context": map[string]any{"values": map[string]any{"path": "/values"}},
			}},
		},
		Component{"id": "submit_label", "component": "Text", "text": "Submit"},
	)
	components = append([]Component{components[0], {"id": "form", "component": "Column", "children": children}}, components[1:]...)
	return []Envelope{
		createEnvelope(f.SurfaceID),
		componentsEnvelope(f.SurfaceID, components),
		dataEnvelope(f.SurfaceID, "/", map[string]any{"values": values, "status": "pending"}),
	}
}

// AnsweredEnvelope marks the form answered (at AnsweredRevision): the client
// disables it so the workflow cannot be resumed twice from the same form.
func (f Form) AnsweredEnvelope() Envelope {
	return dataEnvelope(f.SurfaceID, "/status", "answered")
}

// SubmissionKey is the member of a resume payload that marks it as a form
// submission rather than a plain answer.
const SubmissionKey = "butterForm"

// Submission is a client's form submission, carried in the payload of an
// AG-UI resume entry.
type Submission struct {
	Version   string         `json:"version"`
	SurfaceID string         `json:"surfaceId"`
	Revision  int            `json:"revision"`
	Token     string         `json:"token"`
	Values    map[string]any `json:"values"`
}

// SubmissionOf recognizes a form submission in a resume payload. ok is false
// for a plain answer; a payload that carries the marker but is not a
// well-formed submission is an error.
func SubmissionOf(payload any) (Submission, bool, error) {
	m, isMap := payload.(map[string]any)
	if !isMap {
		return Submission{}, false, nil
	}
	raw, present := m[SubmissionKey]
	if !present {
		return Submission{}, false, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return Submission{}, true, errors.New("butterForm is not valid JSON")
	}
	var sub Submission
	if err := json.Unmarshal(encoded, &sub); err != nil {
		return Submission{}, true, errors.New("butterForm must be {version, surfaceId, revision, token, values}")
	}
	if sub.Version != Version {
		return Submission{}, true, fmt.Errorf("butterForm.version must be %q", Version)
	}
	if sub.SurfaceID == "" || sub.Token == "" || sub.Values == nil {
		return Submission{}, true, errors.New("butterForm requires surfaceId, revision, token and values")
	}
	return sub, true, nil
}

// SubmitErrorKind classifies a rejected submission.
type SubmitErrorKind int

const (
	// SubmitUnknown: no form in this session matches — unknown, forged, or
	// submitted from another context.
	SubmitUnknown SubmitErrorKind = iota
	// SubmitAnswered: the form's Interrupt was already answered.
	SubmitAnswered
	// SubmitStale: the submission addresses an older revision of the form.
	SubmitStale
	// SubmitInvalid: the values break the form's field rules.
	SubmitInvalid
)

// SubmitError explains why a submission was rejected. Nothing was consumed.
type SubmitError struct {
	Kind        SubmitErrorKind
	Message     string
	FieldErrors map[string]string
}

func (e *SubmitError) Error() string { return e.Message }

// Resolve checks a submission against the session and returns the answer to
// deliver to the workflow: the configured fields, in configured order, as a
// JSON object text. The Interrupt must be the one the form was built for and
// must still be pending; a rejection never falls back to answering another
// Interrupt.
func Resolve(sess session.Session, interruptID string, sub Submission) (string, error) {
	var form Form
	found := false
	if sess != nil {
		for _, f := range formsByInterrupt(sess) {
			if f.SurfaceID == sub.SurfaceID {
				form, found = f, true
				break
			}
		}
	}
	if !found || form.InterruptID != interruptID ||
		subtle.ConstantTimeCompare([]byte(form.Token), []byte(sub.Token)) != 1 {
		return "", &SubmitError{Kind: SubmitUnknown, Message: "unknown or expired form"}
	}
	if sub.Revision != form.Revision {
		return "", &SubmitError{Kind: SubmitStale, Message: "this form has changed; reload it and submit again"}
	}
	if !interrupt.PendingIDs(sess)[interruptID] {
		return "", &SubmitError{Kind: SubmitAnswered, Message: "this form was already submitted"}
	}
	answer, fieldErrors := form.Answer(sub.Values)
	if len(fieldErrors) > 0 {
		return "", &SubmitError{Kind: SubmitInvalid, Message: "some fields are invalid", FieldErrors: fieldErrors}
	}
	return answer, nil
}

// Answer validates values against the form's fields and encodes the answer.
// Every configured field appears, in configured order; an unanswered
// optional field is the empty string. Unknown fields and non-string values
// are errors.
func (f Form) Answer(values map[string]any) (string, map[string]string) {
	errs := map[string]string{}
	known := make(map[string]bool, len(f.Fields))
	for _, field := range f.Fields {
		known[field.Name] = true
	}
	for name := range values {
		if !known[name] {
			errs[name] = "not a field of this form"
		}
	}
	answers := make([]string, 0, len(f.Fields))
	for _, field := range f.Fields {
		raw, present := values[field.Name]
		value := ""
		if present && raw != nil {
			s, ok := raw.(string)
			if !ok {
				errs[field.Name] = "must be a string"
				answers = append(answers, "")
				continue
			}
			value = s
		}
		if field.Type == FieldText {
			value = strings.TrimSpace(value)
		}
		switch {
		case value == "" && field.Required:
			errs[field.Name] = "required"
		case value == "":
		case field.Type == FieldText && utf8.RuneCountInString(value) > field.MaxLength:
			errs[field.Name] = fmt.Sprintf("must be at most %d characters", field.MaxLength)
		case field.Type == FieldSingleChoice && !hasOption(field.Options, value):
			errs[field.Name] = "not one of the options"
		}
		answers = append(answers, value)
	}
	if len(errs) > 0 {
		return "", errs
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, field := range f.Fields {
		if i > 0 {
			b.WriteByte(',')
		}
		key, _ := json.Marshal(field.Name)
		val, _ := json.Marshal(answers[i])
		b.Write(key)
		b.WriteByte(':')
		b.Write(val)
	}
	b.WriteByte('}')
	return b.String(), nil
}

func hasOption(options []Option, value string) bool {
	for _, o := range options {
		if o.Value == value {
			return true
		}
	}
	return false
}
