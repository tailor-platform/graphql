package graphql_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/testutil"
)

// Serialises p.Args so tests can tell "absent", "null", and "value" apart.
func probeArgs(p graphql.ResolveParams) (interface{}, error) {
	keys := make([]string, 0, len(p.Args))
	for k := range p.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := map[string]interface{}{"keys": keys}
	for _, k := range keys {
		v := p.Args[k]
		if v == nil {
			out[k] = "null"
		} else {
			out[k] = v
		}
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

var coercionProbeInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"a": &graphql.InputObjectFieldConfig{Type: graphql.String},
		"b": &graphql.InputObjectFieldConfig{Type: graphql.String},
	},
})

var coercionProbeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CoercionProbeQuery",
	Fields: graphql.Fields{
		"probe": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"a": &graphql.ArgumentConfig{Type: graphql.String},
				"b": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: probeArgs,
		},
		"probeObject": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeInputObject},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				obj, _ := p.Args["input"].(map[string]interface{})
				keys := make([]string, 0, len(obj))
				for k := range obj {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				b, _ := json.Marshal(map[string]interface{}{
					"keys": keys,
					"obj":  obj,
				})
				return string(b), nil
			},
		},
	},
})

var coercionProbeSchema, _ = graphql.NewSchema(graphql.SchemaConfig{Query: coercionProbeType})

func runProbe(t *testing.T, field, doc string, vars map[string]interface{}, want string) {
	t.Helper()
	parsed := testutil.TestParse(t, doc)
	result := testutil.TestExecute(t, graphql.ExecuteParams{
		Schema: coercionProbeSchema,
		AST:    parsed,
		Args:   vars,
	})
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	data, _ := result.Data.(map[string]interface{})
	got, _ := data[field].(string)
	if got != want {
		t.Fatalf("probe mismatch\n  got:  %s\n  want: %s", got, want)
	}
}

func TestArgumentCoercion_ScalarVariable_PreservesThreeStates(t *testing.T) {
	doc := `query Probe($a: String, $b: String) { probe(a: $a, b: $b) }`

	t.Run("variable omitted -> argument absent", func(t *testing.T) {
		runProbe(t, "probe", doc,
			map[string]interface{}{"a": "x"},
			`{"a":"x","keys":["a"]}`)
	})
	t.Run("variable explicitly null -> argument present as null", func(t *testing.T) {
		runProbe(t, "probe", doc,
			map[string]interface{}{"a": "x", "b": nil},
			`{"a":"x","b":"null","keys":["a","b"]}`)
	})
	t.Run("variable with value -> argument present with value", func(t *testing.T) {
		runProbe(t, "probe", doc,
			map[string]interface{}{"a": "x", "b": "y"},
			`{"a":"x","b":"y","keys":["a","b"]}`)
	})
}

func TestArgumentCoercion_InputObjectVariable_PreservesThreeStates(t *testing.T) {
	doc := `query Probe($a: String, $b: String) { probeObject(input: {a: $a, b: $b}) }`

	t.Run("nested variable omitted -> field absent in object", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"a": "x"},
			`{"keys":["a"],"obj":{"a":"x"}}`)
	})
	t.Run("nested variable explicitly null -> field present as null", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"a": "x", "b": nil},
			`{"keys":["a","b"],"obj":{"a":"x","b":null}}`)
	})
	t.Run("nested variable with value -> field present with value", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"a": "x", "b": "y"},
			`{"keys":["a","b"],"obj":{"a":"x","b":"y"}}`)
	})
}

func TestArgumentCoercion_InputObjectLiteral_OmittedFieldStaysAbsent(t *testing.T) {
	doc := `{ probeObject(input: {a: "x"}) }`
	runProbe(t, "probeObject", doc, nil,
		`{"keys":["a"],"obj":{"a":"x"}}`)
}
