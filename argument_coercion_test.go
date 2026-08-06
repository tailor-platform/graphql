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

// Serialises the "input" argument so tests can tell an absent input-object
// field apart from one present as null.
func probeObjectArgs(p graphql.ResolveParams) (interface{}, error) {
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
}

var coercionProbeInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"a": &graphql.InputObjectFieldConfig{Type: graphql.String},
		"b": &graphql.InputObjectFieldConfig{Type: graphql.String},
	},
})

// Same shape as coercionProbeInputObject, but field "a" declares a default so
// tests can pin down how a default interacts with absent / explicit null.
var coercionProbeDefaultInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeDefaultInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"a": &graphql.InputObjectFieldConfig{Type: graphql.String, DefaultValue: "FIELDDEF"},
		"b": &graphql.InputObjectFieldConfig{Type: graphql.String},
	},
})

var coercionProbeNestedInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeNestedInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"inner": &graphql.InputObjectFieldConfig{Type: coercionProbeInputObject},
	},
})

// Two levels deep, with the default declared on the innermost field, so the
// recursive coercion paths are exercised rather than just the top level.
var coercionProbeNestedDefaultInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeNestedDefaultInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"inner": &graphql.InputObjectFieldConfig{Type: coercionProbeDefaultInputObject},
	},
})

// Three levels deep: proves the rule keeps holding as recursion gets deeper.
var coercionProbeDeepInputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CoercionProbeDeepInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"level2": &graphql.InputObjectFieldConfig{Type: coercionProbeNestedDefaultInputObject},
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
		"probeArgDefault": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"a": &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "ARGDEF"},
			},
			Resolve: probeArgs,
		},
		// Non-null argument carrying a default: spec §5.4.2.1 says it is
		// optional, so omitting it must validate and resolve to the default.
		"probeNonNullDefault": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"a": &graphql.ArgumentConfig{
					Type:         graphql.NewNonNull(graphql.String),
					DefaultValue: "NNDEF",
				},
			},
			Resolve: probeArgs,
		},
		"probeList": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"a": &graphql.ArgumentConfig{Type: graphql.NewList(graphql.String)},
			},
			Resolve: probeArgs,
		},
		"probeObject": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeInputObject},
			},
			Resolve: probeObjectArgs,
		},
		"probeObjectDefault": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeDefaultInputObject},
			},
			Resolve: probeObjectArgs,
		},
		"probeNested": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeNestedInputObject},
			},
			Resolve: probeObjectArgs,
		},
		"probeNestedDefault": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeNestedDefaultInputObject},
			},
			Resolve: probeObjectArgs,
		},
		"probeDeep": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeDeepInputObject},
			},
			Resolve: probeObjectArgs,
		},
		"probeObjectList": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: graphql.NewList(coercionProbeDefaultInputObject)},
			},
			Resolve: probeArgs,
		},
		// Same argument as probeObject, but serialised with probeArgs so tests
		// can tell an absent input-object argument from one that is null.
		"probeObjectRaw": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: coercionProbeInputObject},
			},
			Resolve: probeArgs,
		},
	},
})

// The same probe types under both coercion modes. SpecCompliantArgumentCoercion
// is opt-in, so the zero-valued config is the behaviour shipped before this
// change and must stay byte-for-byte identical.
var coercionProbeLegacySchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: coercionProbeType,
})

var coercionProbeSpecSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:                         coercionProbeType,
	SpecCompliantArgumentCoercion: true,
})

func execProbe(t *testing.T, schema graphql.Schema, field, doc string, vars map[string]interface{}) string {
	t.Helper()
	parsed := testutil.TestParse(t, doc)
	result := testutil.TestExecute(t, graphql.ExecuteParams{
		Schema: schema,
		AST:    parsed,
		Args:   vars,
	})
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	data, _ := result.Data.(map[string]interface{})
	got, _ := data[field].(string)
	return got
}

// runProbe asserts both coercion modes agree — the cases the flag does not
// change.
func runProbe(t *testing.T, field, doc string, vars map[string]interface{}, want string) {
	t.Helper()
	runProbeModes(t, field, doc, vars, want, want)
}

// runProbeModes pins down a case where the flag changes the outcome: the legacy
// column is the regression guard, the spec column is the fix.
func runProbeModes(t *testing.T, field, doc string, vars map[string]interface{}, wantLegacy, wantSpec string) {
	t.Helper()
	if got := execProbe(t, coercionProbeLegacySchema, field, doc, vars); got != wantLegacy {
		t.Errorf("legacy mode mismatch\n  got:  %s\n  want: %s", got, wantLegacy)
	}
	if got := execProbe(t, coercionProbeSpecSchema, field, doc, vars); got != wantSpec {
		t.Errorf("spec mode mismatch\n  got:  %s\n  want: %s", got, wantSpec)
	}
}

func TestArgumentCoercion_ScalarVariable_PreservesThreeStates(t *testing.T) {
	doc := `query Probe($a: String, $b: String) { probe(a: $a, b: $b) }`

	t.Run("variable omitted -> argument absent", func(t *testing.T) {
		runProbeModes(t, "probe", doc,
			map[string]interface{}{"a": "x"},
			`{"a":"x","b":"null","keys":["a","b"]}`,
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
		runProbeModes(t, "probeObject", doc,
			map[string]interface{}{"a": "x"},
			`{"keys":["a","b"],"obj":{"a":"x","b":null}}`,
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

func TestArgumentCoercion_ScalarArgument_OmittedFromQueryStaysAbsent(t *testing.T) {
	// The argument is not written in the query at all: CoerceArgumentValues
	// leaves hasValue false and there is no default, so nothing is added.
	runProbe(t, "probe", `{ probe(a: "x") }`, nil,
		`{"a":"x","keys":["a"]}`)
}

// Spec: CoerceArgumentValues (§6.4.1). The argument default applies only when
// the caller supplied no value at all. An explicit null is a supplied value and
// must survive as null rather than fall back to the default.
func TestArgumentCoercion_ArgumentDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	doc := `query Probe($a: String) { probeArgDefault(a: $a) }`

	t.Run("argument omitted from query -> default", func(t *testing.T) {
		runProbe(t, "probeArgDefault", `{ probeArgDefault }`, nil,
			`{"a":"ARGDEF","keys":["a"]}`)
	})
	t.Run("variable omitted -> default", func(t *testing.T) {
		runProbe(t, "probeArgDefault", doc,
			map[string]interface{}{},
			`{"a":"ARGDEF","keys":["a"]}`)
	})
	t.Run("variable explicitly null -> null, not the default", func(t *testing.T) {
		runProbeModes(t, "probeArgDefault", doc,
			map[string]interface{}{"a": nil},
			`{"a":"ARGDEF","keys":["a"]}`,
			`{"a":"null","keys":["a"]}`)
	})
	t.Run("variable with value -> value", func(t *testing.T) {
		runProbe(t, "probeArgDefault", doc,
			map[string]interface{}{"a": "v"},
			`{"a":"v","keys":["a"]}`)
	})
}

// Spec: input object field defaults (§3.10 Input Coercion), reached through an
// object literal written in the query document.
func TestArgumentCoercion_InputObjectLiteralFieldDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	doc := `query Probe($a: String) { probeObjectDefault(input: {a: $a}) }`

	t.Run("field omitted from literal -> default", func(t *testing.T) {
		runProbe(t, "probeObjectDefault", `{ probeObjectDefault(input: {}) }`, nil,
			`{"keys":["a"],"obj":{"a":"FIELDDEF"}}`)
	})
	t.Run("nested variable omitted -> default", func(t *testing.T) {
		runProbeModes(t, "probeObjectDefault", doc,
			map[string]interface{}{},
			`{"keys":["a"],"obj":{"a":null}}`,
			`{"keys":["a"],"obj":{"a":"FIELDDEF"}}`)
	})
	t.Run("nested variable explicitly null -> null, not the default", func(t *testing.T) {
		runProbe(t, "probeObjectDefault", doc,
			map[string]interface{}{"a": nil},
			`{"keys":["a"],"obj":{"a":null}}`)
	})
	t.Run("nested variable with value -> value", func(t *testing.T) {
		runProbe(t, "probeObjectDefault", doc,
			map[string]interface{}{"a": "v"},
			`{"keys":["a"],"obj":{"a":"v"}}`)
	})
}

// The whole input object arrives as one variable, so presence is decided by
// whether the key exists in the supplied JSON object (coerceValue path).
func TestArgumentCoercion_WholeObjectVariable_PreservesThreeStates(t *testing.T) {
	doc := `query Probe($in: CoercionProbeInput) { probeObject(input: $in) }`

	t.Run("key absent in object -> field absent", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": "x"}},
			`{"keys":["a"],"obj":{"a":"x"}}`)
	})
	t.Run("key explicitly null -> field present as null", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": "x", "b": nil}},
			`{"keys":["a","b"],"obj":{"a":"x","b":null}}`)
	})
	t.Run("key with value -> field present with value", func(t *testing.T) {
		runProbe(t, "probeObject", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": "x", "b": "y"}},
			`{"keys":["a","b"],"obj":{"a":"x","b":"y"}}`)
	})
}

func TestArgumentCoercion_WholeObjectVariableFieldDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	doc := `query Probe($in: CoercionProbeDefaultInput) { probeObjectDefault(input: $in) }`

	t.Run("key absent in object -> default", func(t *testing.T) {
		runProbe(t, "probeObjectDefault", doc,
			map[string]interface{}{"in": map[string]interface{}{}},
			`{"keys":["a"],"obj":{"a":"FIELDDEF"}}`)
	})
	t.Run("key explicitly null -> null, not the default", func(t *testing.T) {
		runProbeModes(t, "probeObjectDefault", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": nil}},
			`{"keys":["a"],"obj":{"a":"FIELDDEF"}}`,
			`{"keys":["a"],"obj":{"a":null}}`)
	})
	t.Run("key with value -> value", func(t *testing.T) {
		runProbe(t, "probeObjectDefault", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": "v"}},
			`{"keys":["a"],"obj":{"a":"v"}}`)
	})
}

// Spec: CoerceVariableValues (§6.1.2). Same rule one level up — the variable
// default applies only when the caller supplied no value for that variable.
func TestArgumentCoercion_VariableDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	doc := `query Probe($a: String = "VARDEF") { probe(a: $a) }`

	t.Run("variable omitted -> variable default", func(t *testing.T) {
		runProbe(t, "probe", doc,
			map[string]interface{}{},
			`{"a":"VARDEF","keys":["a"]}`)
	})
	t.Run("variable explicitly null -> null, not the default", func(t *testing.T) {
		runProbeModes(t, "probe", doc,
			map[string]interface{}{"a": nil},
			`{"a":"VARDEF","keys":["a"]}`,
			`{"a":"null","keys":["a"]}`)
	})
	t.Run("variable with value -> value", func(t *testing.T) {
		runProbe(t, "probe", doc,
			map[string]interface{}{"a": "v"},
			`{"a":"v","keys":["a"]}`)
	})
}

// A variable default and an argument default in the same position: the variable
// default wins because it makes the argument "supplied".
func TestArgumentCoercion_VariableDefaultTakesPrecedenceOverArgumentDefault(t *testing.T) {
	doc := `query Probe($a: String = "VARDEF") { probeArgDefault(a: $a) }`

	t.Run("variable omitted -> variable default, not argument default", func(t *testing.T) {
		runProbe(t, "probeArgDefault", doc,
			map[string]interface{}{},
			`{"a":"VARDEF","keys":["a"]}`)
	})
	t.Run("variable explicitly null -> null, neither default", func(t *testing.T) {
		runProbeModes(t, "probeArgDefault", doc,
			map[string]interface{}{"a": nil},
			`{"a":"VARDEF","keys":["a"]}`,
			`{"a":"null","keys":["a"]}`)
	})
}

func TestArgumentCoercion_ListArgument_PreservesThreeStates(t *testing.T) {
	doc := `query Probe($a: [String]) { probeList(a: $a) }`

	t.Run("variable omitted -> argument absent", func(t *testing.T) {
		runProbeModes(t, "probeList", doc,
			map[string]interface{}{},
			`{"a":"null","keys":["a"]}`,
			`{"keys":[]}`)
	})
	t.Run("variable explicitly null -> argument present as null", func(t *testing.T) {
		runProbe(t, "probeList", doc,
			map[string]interface{}{"a": nil},
			`{"a":"null","keys":["a"]}`)
	})
	t.Run("variable with value -> argument present with value", func(t *testing.T) {
		runProbe(t, "probeList", doc,
			map[string]interface{}{"a": []interface{}{"x"}},
			`{"a":["x"],"keys":["a"]}`)
	})
	t.Run("unprovided variable inside a list literal -> null item", func(t *testing.T) {
		runProbe(t, "probeList", `query Probe($v: String) { probeList(a: ["x", $v]) }`,
			map[string]interface{}{},
			`{"a":["x",null],"keys":["a"]}`)
	})
}

func TestArgumentCoercion_NestedInputObject_PreservesAbsentAndNull(t *testing.T) {
	t.Run("literal: unprovided variable in nested object stays absent", func(t *testing.T) {
		runProbeModes(t, "probeNested",
			`query Probe($a: String) { probeNested(input: {inner: {a: $a}}) }`,
			map[string]interface{}{},
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`,
			`{"keys":["inner"],"obj":{"inner":{}}}`)
	})
	t.Run("literal: explicit null in nested object stays null", func(t *testing.T) {
		runProbe(t, "probeNested",
			`query Probe($a: String) { probeNested(input: {inner: {a: $a}}) }`,
			map[string]interface{}{"a": nil},
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`)
	})
	t.Run("whole variable: absent nested key stays absent", func(t *testing.T) {
		runProbe(t, "probeNested",
			`query Probe($in: CoercionProbeNestedInput) { probeNested(input: $in) }`,
			map[string]interface{}{"in": map[string]interface{}{"inner": map[string]interface{}{}}},
			`{"keys":["inner"],"obj":{"inner":{}}}`)
	})
	t.Run("whole variable: explicit null nested key stays null", func(t *testing.T) {
		runProbe(t, "probeNested",
			`query Probe($in: CoercionProbeNestedInput) { probeNested(input: $in) }`,
			map[string]interface{}{"in": map[string]interface{}{"inner": map[string]interface{}{"a": nil}}},
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`)
	})
}

// The nested object itself — not one of its fields — is the thing that is
// absent or null.
func TestArgumentCoercion_NestedInputObject_ObjectValuedFieldAbsentVsNull(t *testing.T) {
	litDoc := `query Probe($innerVar: CoercionProbeInput) { probeNested(input: {inner: $innerVar}) }`
	varDoc := `query Probe($in: CoercionProbeNestedInput) { probeNested(input: $in) }`

	t.Run("literal: unprovided object variable -> field absent", func(t *testing.T) {
		runProbeModes(t, "probeNested", litDoc,
			map[string]interface{}{},
			`{"keys":["inner"],"obj":{"inner":null}}`,
			`{"keys":[],"obj":{}}`)
	})
	t.Run("literal: explicitly null object variable -> field present as null", func(t *testing.T) {
		runProbe(t, "probeNested", litDoc,
			map[string]interface{}{"innerVar": nil},
			`{"keys":["inner"],"obj":{"inner":null}}`)
	})
	t.Run("whole variable: object key absent -> field absent", func(t *testing.T) {
		runProbe(t, "probeNested", varDoc,
			map[string]interface{}{"in": map[string]interface{}{}},
			`{"keys":[],"obj":{}}`)
	})
	t.Run("whole variable: object key explicitly null -> field present as null", func(t *testing.T) {
		runProbe(t, "probeNested", varDoc,
			map[string]interface{}{"in": map[string]interface{}{"inner": nil}},
			`{"keys":["inner"],"obj":{"inner":null}}`)
	})
}

// The input-object argument itself is absent or null, one level above the
// object's fields.
func TestArgumentCoercion_InputObjectArgument_AbsentVsNull(t *testing.T) {
	doc := `query Probe($in: CoercionProbeInput) { probeObjectRaw(input: $in) }`

	t.Run("variable omitted -> argument absent", func(t *testing.T) {
		runProbeModes(t, "probeObjectRaw", doc,
			map[string]interface{}{},
			`{"input":"null","keys":["input"]}`,
			`{"keys":[]}`)
	})
	t.Run("variable explicitly null -> argument present as null", func(t *testing.T) {
		runProbe(t, "probeObjectRaw", doc,
			map[string]interface{}{"in": nil},
			`{"input":"null","keys":["input"]}`)
	})
	t.Run("variable with value -> argument present with value", func(t *testing.T) {
		runProbe(t, "probeObjectRaw", doc,
			map[string]interface{}{"in": map[string]interface{}{"a": "x"}},
			`{"input":{"a":"x"},"keys":["input"]}`)
	})
}

// The default lives on a field one level down, so this only passes if the
// absent/null rule is applied by the recursive step and not just at the top.
func TestArgumentCoercion_NestedFieldDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	litDoc := `query Probe($a: String) { probeNestedDefault(input: {inner: {a: $a}}) }`
	varDoc := `query Probe($in: CoercionProbeNestedDefaultInput) { probeNestedDefault(input: $in) }`

	t.Run("literal: nested field omitted -> default", func(t *testing.T) {
		runProbe(t, "probeNestedDefault",
			`{ probeNestedDefault(input: {inner: {}}) }`, nil,
			`{"keys":["inner"],"obj":{"inner":{"a":"FIELDDEF"}}}`)
	})
	t.Run("literal: nested variable omitted -> default", func(t *testing.T) {
		runProbeModes(t, "probeNestedDefault", litDoc,
			map[string]interface{}{},
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`,
			`{"keys":["inner"],"obj":{"inner":{"a":"FIELDDEF"}}}`)
	})
	t.Run("literal: nested variable explicitly null -> null, not the default", func(t *testing.T) {
		runProbe(t, "probeNestedDefault", litDoc,
			map[string]interface{}{"a": nil},
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`)
	})
	t.Run("whole variable: nested key absent -> default", func(t *testing.T) {
		runProbe(t, "probeNestedDefault", varDoc,
			map[string]interface{}{"in": map[string]interface{}{"inner": map[string]interface{}{}}},
			`{"keys":["inner"],"obj":{"inner":{"a":"FIELDDEF"}}}`)
	})
	t.Run("whole variable: nested key explicitly null -> null, not the default", func(t *testing.T) {
		runProbeModes(t, "probeNestedDefault", varDoc,
			map[string]interface{}{"in": map[string]interface{}{"inner": map[string]interface{}{"a": nil}}},
			`{"keys":["inner"],"obj":{"inner":{"a":"FIELDDEF"}}}`,
			`{"keys":["inner"],"obj":{"inner":{"a":null}}}`)
	})
}

func TestArgumentCoercion_DeeplyNestedFieldDefault_AppliesOnlyWhenValueAbsent(t *testing.T) {
	litDoc := `query Probe($a: String) { probeDeep(input: {level2: {inner: {a: $a}}}) }`
	varDoc := `query Probe($in: CoercionProbeDeepInput) { probeDeep(input: $in) }`

	t.Run("literal: three levels down, variable omitted -> default", func(t *testing.T) {
		runProbeModes(t, "probeDeep", litDoc,
			map[string]interface{}{},
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":null}}}}`,
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":"FIELDDEF"}}}}`)
	})
	t.Run("literal: three levels down, explicitly null -> null, not the default", func(t *testing.T) {
		runProbe(t, "probeDeep", litDoc,
			map[string]interface{}{"a": nil},
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":null}}}}`)
	})
	t.Run("whole variable: three levels down, key absent -> default", func(t *testing.T) {
		runProbe(t, "probeDeep", varDoc,
			map[string]interface{}{"in": map[string]interface{}{
				"level2": map[string]interface{}{"inner": map[string]interface{}{}},
			}},
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":"FIELDDEF"}}}}`)
	})
	t.Run("whole variable: three levels down, explicitly null -> null, not the default", func(t *testing.T) {
		runProbeModes(t, "probeDeep", varDoc,
			map[string]interface{}{"in": map[string]interface{}{
				"level2": map[string]interface{}{"inner": map[string]interface{}{"a": nil}},
			}},
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":"FIELDDEF"}}}}`,
			`{"keys":["level2"],"obj":{"level2":{"inner":{"a":null}}}}`)
	})
}

// Input objects inside a list: the recursive step runs per element, so each
// element must keep its own absent / null / value state.
func TestArgumentCoercion_ListOfInputObjects_PreservesPerElementState(t *testing.T) {
	t.Run("literal: element field omitted -> default", func(t *testing.T) {
		runProbeModes(t, "probeObjectList",
			`query Probe($a: String) { probeObjectList(input: [{a: $a}]) }`,
			map[string]interface{}{},
			`{"input":[{"a":null}],"keys":["input"]}`,
			`{"input":[{"a":"FIELDDEF"}],"keys":["input"]}`)
	})
	t.Run("literal: element field explicitly null -> null, not the default", func(t *testing.T) {
		runProbe(t, "probeObjectList",
			`query Probe($a: String) { probeObjectList(input: [{a: $a}]) }`,
			map[string]interface{}{"a": nil},
			`{"input":[{"a":null}],"keys":["input"]}`)
	})
	t.Run("whole variable: per-element null and absent are independent", func(t *testing.T) {
		runProbeModes(t, "probeObjectList",
			`query Probe($in: [CoercionProbeDefaultInput]) { probeObjectList(input: $in) }`,
			map[string]interface{}{"in": []interface{}{
				map[string]interface{}{"a": nil},
				map[string]interface{}{},
			}},
			`{"input":[{"a":"FIELDDEF"},{"a":"FIELDDEF"}],"keys":["input"]}`,
			`{"input":[{"a":null},{"a":"FIELDDEF"}],"keys":["input"]}`)
	})
}

// Spec §5.4.2.1, end to end through graphql.Do so document validation runs too.
// This relaxation is not gated by SpecCompliantArgumentCoercion: it only lets
// queries through that previously failed, so it cannot break a working one.
func TestArgumentCoercion_NonNullArgumentWithDefault_IsOptionalInBothModes(t *testing.T) {
	for _, tc := range []struct {
		mode   string
		schema graphql.Schema
	}{
		{"legacy", coercionProbeLegacySchema},
		{"spec", coercionProbeSpecSchema},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			result := graphql.Do(graphql.Params{
				Schema:        tc.schema,
				RequestString: `{ probeNonNullDefault }`,
			})
			if len(result.Errors) > 0 {
				t.Fatalf("unexpected errors: %v", result.Errors)
			}
			data, _ := result.Data.(map[string]interface{})
			got, _ := data["probeNonNullDefault"].(string)
			want := `{"a":"NNDEF","keys":["a"]}`
			if got != want {
				t.Fatalf("probe mismatch\n  got:  %s\n  want: %s", got, want)
			}
		})
	}
}
