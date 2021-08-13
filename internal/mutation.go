package internal

import (
	"context"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"os"
	"strings"
)

type Mutation struct {
	Id       int
	Filter   string
	Mutation string
}

var runtimeEnv = generateEnvAst()

func applyMutations(input map[string]interface{}, mutations []Mutation) ([]map[string]interface{}, error) {
	patches := make([]map[string]interface{}, 0)

	for _, mutation := range mutations {
		filterDoesMatch, err := checkFilter(input, mutation.Filter)
		if err != nil {
			return patches, err
		}
		if !filterDoesMatch {
			continue
		}
		p, err := generatePatches(input, mutation.Mutation)
		if err != nil {
			return patches, err
		}
		patches = append(patches, p...)
	}

	return patches, nil
}

func generatePatches(input map[string]interface{}, module string) ([]map[string]interface{}, error) {
	ret := make([]map[string]interface{}, 0)

	ctx := context.Background()
	query, err := rego.New(
		rego.Module("example.rego", module),
		rego.Query("data.mutation.mutate[x]"),
		rego.Runtime(runtimeEnv),
	).PrepareForEval(ctx)

	if err != nil {
		return ret, err
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil || len(results) == 0 {
		return ret, err
	}

	for _, res := range results {
		ret = append(ret, res.Bindings["x"].(map[string]interface{}))
	}

	return ret, nil
}

func checkFilter(input map[string]interface{}, module string) (bool, error) {
	ctx := context.Background()
	query, err := rego.New(
		rego.Module("example.rego", module),
		rego.Query("x = data.filter.matches"),
		rego.Runtime(runtimeEnv),
	).PrepareForEval(ctx)

	if err != nil {
		return false, err
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil || len(results) == 0 {
		return false, err
	}

	for _, result := range results {
		if !result.Bindings["x"].(bool) {
			return false, nil
		}
	}

	return true, nil
}

func generateEnvAst() *ast.Term {
	obj := ast.NewObject()
	env := ast.NewObject()
	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}
	obj.Insert(ast.StringTerm("env"), ast.NewTerm(env))

	return ast.NewTerm(obj)
}
