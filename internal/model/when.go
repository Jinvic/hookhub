package model

import (
	"fmt"
	"math"
	"sync"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
)

var (
	whenEnvOnce sync.Once
	whenEnv     *cel.Env
	whenEnvErr  error
)

func getWhenEnv() (*cel.Env, error) {
	whenEnvOnce.Do(func() {
		whenEnv, whenEnvErr = cel.NewEnv(
			cel.Variable("source", cel.StringType),
			cel.Variable("type", cel.StringType),
			cel.Variable("subject", cel.StringType),
			cel.Variable("data", cel.MapType(cel.StringType, cel.DynType)),
			cel.Variable("vars", cel.MapType(cel.StringType, cel.StringType)),
		)
	})
	return whenEnv, whenEnvErr
}

// CompileWhen compiles a CEL when expression that must evaluate to bool.
// An empty expr returns a nil program (treated as always true).
func CompileWhen(expr string) (cel.Program, error) {
	if expr == "" {
		return nil, nil
	}
	env, err := getWhenEnv()
	if err != nil {
		return nil, err
	}
	ast, iss := env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}
	if !ast.OutputType().IsExactType(types.BoolType) {
		return nil, fmt.Errorf("when must return bool, got %v", ast.OutputType())
	}
	return env.Program(ast)
}

// EvalWhen evaluates a compiled when program against e and vars.
// A nil program returns true.
func EvalWhen(prg cel.Program, e *Event, vars map[string]string) (bool, error) {
	if prg == nil {
		return true, nil
	}
	if e == nil {
		return false, fmt.Errorf("nil event")
	}
	if vars == nil {
		vars = map[string]string{}
	}
	data := e.Data
	if data == nil {
		data = map[string]any{}
	} else {
		data = normalizeJSONNumbers(data).(map[string]any)
	}

	out, _, err := prg.Eval(map[string]any{
		"source":  e.Source,
		"type":    e.Type,
		"subject": e.Subject,
		"data":    data,
		"vars":    vars,
	})
	if err != nil {
		return false, err
	}
	b, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("when result type %T", out.Value())
	}
	return b, nil
}

// normalizeJSONNumbers converts whole-number float64 values (common after
// encoding/json into map[string]any) to int64 so CEL can use `== 2`.
func normalizeJSONNumbers(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, child := range x {
			out[k] = normalizeJSONNumbers(child)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, child := range x {
			out[i] = normalizeJSONNumbers(child)
		}
		return out
	case float64:
		if !math.IsNaN(x) && !math.IsInf(x, 0) && x == math.Trunc(x) &&
			x >= float64(math.MinInt64) && x <= float64(math.MaxInt64) {
			return int64(x)
		}
		return x
	default:
		return v
	}
}
