package mcpcodemode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/dop251/goja"
	"google.golang.org/adk/v2/agent"
)

// maxCallStack bounds the script's recursion depth.
const maxCallStack = 256

// result is what one script run produced.
type result struct {
	value     any
	truncated bool
	logs      []string
	toolCalls int
	err       error
}

// execute runs script against the catalog's tools. It never panics: whatever
// the script or a tool does ends up in result.err.
func execute(parent agent.Context, cat *catalog, limits Limits, script string) (res result) {
	runCtx, cancel := context.WithTimeout(parent, limits.Timeout)
	defer cancel()

	vm := goja.New()
	vm.SetMaxCallStackSize(maxCallStack)

	sb := &sandbox{
		vm:     vm,
		cat:    cat,
		limits: limits,
		agent:  deadlineContext{Context: parent, ctx: runCtx},
		ctx:    runCtx,
	}
	defer func() {
		res.logs = sb.logs
		res.toolCalls = int(sb.calls.Load())
		if r := recover(); r != nil {
			res.err = fmt.Errorf("script crashed: %v", r)
		}
	}()

	if err := sb.install(); err != nil {
		res.err = err
		return res
	}

	// Stop the script when it runs out of time or the turn is cancelled. A tool
	// call in flight sees the same context and returns on its own.
	stop := context.AfterFunc(runCtx, func() {
		reason := "cancelled"
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			reason = fmt.Sprintf("timed out after %s", limits.Timeout)
		}
		vm.Interrupt("script " + reason)
	})
	defer stop()

	out, err := vm.RunString("(function () {\n\"use strict\";\n" + script + "\n})()")
	if err != nil {
		res.err = describeScriptError(err)
		return res
	}
	res.value, res.truncated, res.err = sb.render(out)
	return res
}

func describeScriptError(err error) error {
	var interrupted *goja.InterruptedError
	if errors.As(err, &interrupted) {
		return fmt.Errorf("%v", interrupted.Value())
	}
	var exc *goja.Exception
	if errors.As(err, &exc) {
		return errors.New(strings.TrimSpace(exc.String()))
	}
	return err
}

type sandbox struct {
	vm     *goja.Runtime
	cat    *catalog
	limits Limits
	agent  agent.Context
	ctx    context.Context

	calls atomic.Int64

	logBytes int
	logs     []string

	stringify goja.Callable
	parse     goja.Callable
}

// install defines the script's globals. Everything else goja offers is
// plain ECMAScript: there is no require, fetch, process, or file access.
func (s *sandbox) install() error {
	vm := s.vm
	jsonObj := vm.Get("JSON").ToObject(vm)
	var ok bool
	if s.stringify, ok = goja.AssertFunction(jsonObj.Get("stringify")); !ok {
		return errors.New("sandbox has no JSON.stringify")
	}
	if s.parse, ok = goja.AssertFunction(jsonObj.Get("parse")); !ok {
		return errors.New("sandbox has no JSON.parse")
	}

	toolsObj := vm.NewObject()
	for _, e := range s.cat.entries {
		name := e.name
		if err := toolsObj.Set(name, func(call goja.FunctionCall) goja.Value {
			args, err := s.argsOf(call.Argument(0))
			if err != nil {
				panic(vm.NewGoError(fmt.Errorf("tools.%s: %w", name, err)))
			}
			out, err := s.invoke(name, args)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return s.toJS(out)
		}); err != nil {
			return err
		}
	}
	names := make([]string, len(s.cat.entries))
	for i, e := range s.cat.entries {
		names[i] = e.name
	}

	globals := map[string]any{
		"tools":     toolsObj,
		"ALL_TOOLS": s.toJS(names),
		"callTool": func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			args, err := s.argsOf(call.Argument(1))
			if err != nil {
				panic(vm.NewGoError(fmt.Errorf("callTool(%q): %w", name, err)))
			}
			out, err := s.invoke(name, args)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			return s.toJS(out)
		},
		"callTools": s.callTools,
		"searchTools": func(call goja.FunctionCall) goja.Value {
			found := s.cat.search(call.Argument(0).String())
			items := make([]map[string]any, len(found))
			for i, e := range found {
				items[i] = map[string]any{"name": e.name, "description": firstLine(e.description), "params": e.params}
			}
			return s.toJS(items)
		},
		"describeTool": func(call goja.FunctionCall) goja.Value {
			name := call.Argument(0).String()
			e, ok := s.cat.byName[name]
			if !ok {
				panic(vm.NewGoError(unknownTool(name)))
			}
			return s.toJS(map[string]any{"name": e.name, "description": e.description, "inputSchema": e.schema})
		},
	}
	for k, v := range globals {
		if err := vm.Set(k, v); err != nil {
			return err
		}
	}

	console := vm.NewObject()
	for _, level := range []string{"log", "info", "debug", "warn", "error"} {
		prefix := ""
		if level == "warn" || level == "error" {
			prefix = level + ": "
		}
		if err := console.Set(level, func(call goja.FunctionCall) goja.Value {
			s.log(prefix, call.Arguments)
			return goja.Undefined()
		}); err != nil {
			return err
		}
	}
	return vm.Set("console", console)
}

func unknownTool(name string) error {
	return fmt.Errorf("unknown tool %q; list tools with ALL_TOOLS or searchTools(query)", name)
}

// argsOf turns a script's argument into a tool's argument object.
func (s *sandbox) argsOf(v goja.Value) (map[string]any, error) {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return map[string]any{}, nil
	}
	// Round-trip through JSON so the tool sees plain Go values, never goja
	// wrappers, and a value JSON cannot hold is an error here.
	text, err := s.stringify(goja.Undefined(), v)
	if err != nil {
		return nil, fmt.Errorf("arguments are not JSON-serializable: %w", err)
	}
	if goja.IsUndefined(text) {
		return nil, errors.New("arguments must be an object")
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(text.String()), &args); err != nil {
		return nil, errors.New("arguments must be an object")
	}
	if args == nil {
		args = map[string]any{}
	}
	return args, nil
}

// reserve counts n tool calls against the script's budget.
func (s *sandbox) reserve(n int) error {
	if s.calls.Add(int64(n)) > int64(s.limits.MaxToolCalls) {
		s.calls.Add(-int64(n))
		return fmt.Errorf("tool call limit exceeded: a script may make at most %d tool calls", s.limits.MaxToolCalls)
	}
	return nil
}

// invoke calls one tool and returns its output as a JSON-compatible value.
func (s *sandbox) invoke(name string, args map[string]any) (any, error) {
	e, ok := s.cat.byName[name]
	if !ok {
		return nil, unknownTool(name)
	}
	if err := s.reserve(1); err != nil {
		return nil, err
	}
	return s.call(e, args)
}

// call runs a tool whose call was already counted. It is safe to run from
// several goroutines: it touches no goja value.
func (s *sandbox) call(e *entry, args map[string]any) (out any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %q crashed: %v", e.name, r)
		}
	}()
	if err := s.ctx.Err(); err != nil {
		return nil, fmt.Errorf("tool %q was not called: %w", e.name, err)
	}
	resp, err := e.run(s.agent, args)
	if err != nil {
		return nil, fmt.Errorf("tool %q failed: %w", e.name, err)
	}
	// An MCP tool answers {"output": ...}; hand the script the output itself.
	if v, ok := resp["output"]; ok && len(resp) == 1 {
		return v, nil
	}
	return resp, nil
}

// callTools runs a batch of calls concurrently and reports each outcome, so
// one failing call does not lose the others' results.
func (s *sandbox) callTools(call goja.FunctionCall) goja.Value {
	vm := s.vm
	arr, ok := call.Argument(0).Export().([]any)
	if !ok {
		panic(vm.NewTypeError("callTools expects an array of {name, args}"))
	}
	type job struct {
		e    *entry
		args map[string]any
	}
	jobs := make([]job, len(arr))
	for i, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			panic(vm.NewTypeError("callTools item %d must be an object {name, args}", i))
		}
		name, _ := m["name"].(string)
		e, ok := s.cat.byName[name]
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("callTools item %d: %w", i, unknownTool(name))))
		}
		args, err := s.argsOf(vm.ToValue(m["args"]))
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("callTools item %d (%s): %w", i, name, err)))
		}
		jobs[i] = job{e, args}
	}
	if err := s.reserve(len(jobs)); err != nil {
		panic(vm.NewGoError(err))
	}

	type outcome struct {
		out any
		err error
	}
	outcomes := make([]outcome, len(jobs))
	sem := make(chan struct{}, s.limits.MaxParallel)
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			out, err := s.call(j.e, j.args)
			outcomes[i] = outcome{out, err}
		}()
	}
	wg.Wait()

	items := make([]any, len(outcomes))
	for i, o := range outcomes {
		if o.err != nil {
			items[i] = map[string]any{"ok": false, "error": o.err.Error()}
		} else {
			items[i] = map[string]any{"ok": true, "result": o.out}
		}
	}
	return s.toJS(items)
}

// toJS converts a JSON-compatible Go value into a plain script value.
func (s *sandbox) toJS(v any) goja.Value {
	b, err := json.Marshal(v)
	if err != nil {
		// A tool output that is not JSON: hand over its text.
		return s.vm.ToValue(fmt.Sprint(v))
	}
	parsed, err := s.parse(goja.Undefined(), s.vm.ToValue(string(b)))
	if err != nil {
		return s.vm.ToValue(string(b))
	}
	return parsed
}

// render turns the script's return value into the Go value the model sees,
// bounded to MaxOutputBytes of JSON.
func (s *sandbox) render(v goja.Value) (value any, truncated bool, err error) {
	if v == nil || goja.IsUndefined(v) {
		return nil, false, nil
	}
	text, err := s.stringify(goja.Undefined(), v)
	if err != nil {
		return nil, false, fmt.Errorf("the returned value is not JSON-serializable: %w", err)
	}
	if goja.IsUndefined(text) {
		return nil, false, nil
	}
	body := text.String()
	if len(body) > s.limits.MaxOutputBytes {
		cut := body[:s.limits.MaxOutputBytes]
		for !utf8.ValidString(cut) {
			cut = cut[:len(cut)-1]
		}
		return fmt.Sprintf("%s… [output truncated: %d of %d bytes of JSON; return less]", cut, len(cut), len(body)), true, nil
	}
	var out any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		return nil, false, fmt.Errorf("the returned value is not JSON: %w", err)
	}
	return out, false, nil
}

// log records one console call within the log budget.
func (s *sandbox) log(prefix string, args []goja.Value) {
	if s.logBytes >= s.limits.MaxLogBytes {
		return
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = s.display(a)
	}
	line := prefix + strings.Join(parts, " ")
	if s.logBytes+len(line) > s.limits.MaxLogBytes {
		line = line[:max(0, s.limits.MaxLogBytes-s.logBytes)]
		for !utf8.ValidString(line) {
			line = line[:len(line)-1]
		}
		s.logs = append(s.logs, line+"… [logs truncated]")
		s.logBytes = s.limits.MaxLogBytes
		return
	}
	s.logBytes += len(line)
	s.logs = append(s.logs, line)
}

func (s *sandbox) display(v goja.Value) string {
	if v == nil || goja.IsUndefined(v) {
		return "undefined"
	}
	if _, ok := v.Export().(string); ok {
		return v.String()
	}
	if obj, ok := v.(*goja.Object); ok {
		if text, err := s.stringify(goja.Undefined(), obj); err == nil && !goja.IsUndefined(text) {
			return text.String()
		}
	}
	return v.String()
}
