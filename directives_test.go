package graphql_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/gqlerrors"
	"github.com/tailor-platform/graphql/testutil"
)

var directivesTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"a": &graphql.Field{
				Type: graphql.String,
			},
			"b": &graphql.Field{
				Type: graphql.String,
			},
		},
	}),
})

var directivesTestData map[string]interface{} = map[string]interface{}{
	"a": func() interface{} { return "a" },
	"b": func() interface{} { return "b" },
}

func executeDirectivesTestQuery(t *testing.T, doc string) *graphql.Result {
	ast := testutil.TestParse(t, doc)
	ep := graphql.ExecuteParams{
		Schema: directivesTestSchema,
		AST:    ast,
		Root:   directivesTestData,
	}
	return testutil.TestExecute(t, ep)
}

func TestDirectives_DirectivesMustBeNamed(t *testing.T) {
	invalidDirective := graphql.NewDirective(graphql.DirectiveConfig{
		Locations: []string{
			graphql.DirectiveLocationField,
		},
	})
	_, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "TestType",
			Fields: graphql.Fields{
				"a": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
		Directives: []*graphql.Directive{invalidDirective},
	})
	actualErr := gqlerrors.FormatError(err)
	expectedErr := gqlerrors.FormatError(errors.New("Directive must be named."))
	if !testutil.EqualFormattedError(expectedErr, actualErr) {
		t.Fatalf("Expected error to be equal, got: %v", testutil.Diff(expectedErr, actualErr))
	}
}

func TestDirectives_DirectiveNameMustBeValid(t *testing.T) {
	invalidDirective := graphql.NewDirective(graphql.DirectiveConfig{
		Name: "123invalid name",
		Locations: []string{
			graphql.DirectiveLocationField,
		},
	})
	_, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "TestType",
			Fields: graphql.Fields{
				"a": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
		Directives: []*graphql.Directive{invalidDirective},
	})
	actualErr := gqlerrors.FormatError(err)
	expectedErr := gqlerrors.FormatError(errors.New(`Names must match /^[_a-zA-Z][_a-zA-Z0-9]*$/ but "123invalid name" does not.`))
	if !testutil.EqualFormattedError(expectedErr, actualErr) {
		t.Fatalf("Expected error to be equal, got: %v", testutil.Diff(expectedErr, actualErr))
	}
}

func TestDirectives_DirectiveNameMustProvideLocations(t *testing.T) {
	invalidDirective := graphql.NewDirective(graphql.DirectiveConfig{
		Name: "skip",
	})
	_, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "TestType",
			Fields: graphql.Fields{
				"a": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
		Directives: []*graphql.Directive{invalidDirective},
	})
	actualErr := gqlerrors.FormatError(err)
	expectedErr := gqlerrors.FormatError(errors.New(`Must provide locations for directive.`))
	if !testutil.EqualFormattedError(expectedErr, actualErr) {
		t.Fatalf("Expected error to be equal, got: %v", testutil.Diff(expectedErr, actualErr))
	}
}

func TestDirectives_DirectiveArgNamesMustBeValid(t *testing.T) {
	invalidDirective := graphql.NewDirective(graphql.DirectiveConfig{
		Name: "skip",
		Description: "Directs the executor to skip this field or fragment when the `if` " +
			"argument is true.",
		Args: graphql.FieldConfigArgument{
			"123if": &graphql.ArgumentConfig{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Skipped when true.",
			},
		},
		Locations: []string{
			graphql.DirectiveLocationField,
			graphql.DirectiveLocationFragmentSpread,
			graphql.DirectiveLocationInlineFragment,
		},
	})
	_, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "TestType",
			Fields: graphql.Fields{
				"a": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
		Directives: []*graphql.Directive{invalidDirective},
	})
	actualErr := gqlerrors.FormatError(err)
	expectedErr := gqlerrors.FormatError(errors.New(`Names must match /^[_a-zA-Z][_a-zA-Z0-9]*$/ but "123if" does not.`))
	if !testutil.EqualFormattedError(expectedErr, actualErr) {
		t.Fatalf("Expected error to be equal, got: %v", testutil.Diff(expectedErr, actualErr))
	}
}

func TestDirectivesWorksWithoutDirectives(t *testing.T) {
	query := `{ a, b }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnScalarsIfTrueIncludesScalar(t *testing.T) {
	query := `{ a, b @include(if: true) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnScalarsIfFalseOmitsOnScalar(t *testing.T) {
	query := `{ a, b @include(if: false) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnScalarsUnlessFalseIncludesScalar(t *testing.T) {
	query := `{ a, b @skip(if: false) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnScalarsUnlessTrueOmitsScalar(t *testing.T) {
	query := `{ a, b @skip(if: true) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnFragmentSpreadsIfFalseOmitsFragmentSpread(t *testing.T) {
	query := `
        query Q {
          a
          ...Frag @include(if: false)
        }
        fragment Frag on TestType {
          b
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnFragmentSpreadsIfTrueIncludesFragmentSpread(t *testing.T) {
	query := `
        query Q {
          a
          ...Frag @include(if: true)
        }
        fragment Frag on TestType {
          b
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnFragmentSpreadsUnlessFalseIncludesFragmentSpread(t *testing.T) {
	query := `
        query Q {
          a
          ...Frag @skip(if: false)
        }
        fragment Frag on TestType {
          b
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnFragmentSpreadsUnlessTrueOmitsFragmentSpread(t *testing.T) {
	query := `
        query Q {
          a
          ...Frag @skip(if: true)
        }
        fragment Frag on TestType {
          b
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnInlineFragmentIfFalseOmitsInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... on TestType @include(if: false) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnInlineFragmentIfTrueIncludesInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... on TestType @include(if: true) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnInlineFragmentUnlessFalseIncludesInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... on TestType @skip(if: false) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnInlineFragmentUnlessTrueIncludesInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... on TestType @skip(if: true) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnAnonymousInlineFragmentIfFalseOmitsAnonymousInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... @include(if: false) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnAnonymousInlineFragmentIfTrueIncludesAnonymousInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... @include(if: true) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnAnonymousInlineFragmentUnlessFalseIncludesAnonymousInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... @skip(if: false) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksOnAnonymousInlineFragmentUnlessTrueIncludesAnonymousInlineFragment(t *testing.T) {
	query := `
        query Q {
          a
          ... @skip(if: true) {
            b
          }
        }
	`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksWithSkipAndIncludeDirectives_IncludeAndNoSkip(t *testing.T) {
	query := `{ a, b @include(if: true) @skip(if: false) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
			"b": "b",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksWithSkipAndIncludeDirectives_IncludeAndSkip(t *testing.T) {
	query := `{ a, b @include(if: true) @skip(if: true) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksWithSkipAndIncludeDirectives_NoIncludeAndSkip(t *testing.T) {
	query := `{ a, b @include(if: false) @skip(if: true) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

func TestDirectivesWorksWithSkipAndIncludeDirectives_NoIncludeOrSkip(t *testing.T) {
	query := `{ a, b @include(if: false) @skip(if: false) }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"a": "a",
		},
	}
	result := executeDirectivesTestQuery(t, query)
	if !testutil.EqualResults(expected, result) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result))
	}
}

// The same two fields under both modes, so the @skip / @include tests below can
// pin the opt-out column as a regression guard.
var directivesNonSpecTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType",
		Fields: graphql.Fields{
			"a": &graphql.Field{Type: graphql.String},
			"b": &graphql.Field{Type: graphql.String},
		},
	}),
	NonSpecArgumentHandling: true,
})

// runDirectiveModes executes doc through graphql.Do under both modes. Results are
// compared as the serialised data map plus the first error, so a test can pin an
// expected rejection as well as an expected shape.
func runDirectiveModes(t *testing.T, doc string, vars map[string]interface{}, wantNonSpec, wantSpec string) {
	t.Helper()
	run := func(schema graphql.Schema) string {
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  doc,
			VariableValues: vars,
			RootObject:     directivesTestData,
		})
		data, _ := json.Marshal(result.Data)
		if len(result.Errors) > 0 {
			return string(data) + " ERROR: " + result.Errors[0].Message
		}
		return string(data)
	}
	if got := run(directivesNonSpecTestSchema); got != wantNonSpec {
		t.Errorf("non-spec mode mismatch\n  got:  %s\n  want: %s", got, wantNonSpec)
	}
	if got := run(directivesTestSchema); got != wantSpec {
		t.Errorf("spec mode mismatch\n  got:  %s\n  want: %s", got, wantSpec)
	}
}

// Spec §6.3.2 CollectFields tests @skip and @include structurally rather than
// coercing their arguments: @skip drops the selection when if is true, and
// @include keeps it only when if is true. An if that is neither — a variable
// carrying null, say — is simply "not true", and no error is raised.
//
// The one document that reaches this at runtime declares a variable with a
// non-null default, which §5.8.5 allows at the non-null if argument, and then
// supplies null for it.
func TestDirectives_NullIfArgument_IsNotTrue(t *testing.T) {
	vars := map[string]interface{}{"x": nil}

	t.Run("@skip with a null if keeps the selection", func(t *testing.T) {
		// Non-spec mode never sees the null: the variable's default replaces it.
		runDirectiveModes(t, `query Q($x: Boolean = true){ a @skip(if: $x) b }`, vars,
			`{"b":"b"}`, `{"a":"a","b":"b"}`)
	})
	t.Run("@include with a null if drops the selection", func(t *testing.T) {
		runDirectiveModes(t, `query Q($x: Boolean = true){ a @include(if: $x) b }`, vars,
			`{"a":"a","b":"b"}`, `{"b":"b"}`)
	})
	t.Run("a true value still behaves normally", func(t *testing.T) {
		v := map[string]interface{}{"x": true}
		runDirectiveModes(t, `query Q($x: Boolean = true){ a @skip(if: $x) b }`, v,
			`{"b":"b"}`, `{"b":"b"}`)
		runDirectiveModes(t, `query Q($x: Boolean = true){ a @include(if: $x) b }`, v,
			`{"a":"a","b":"b"}`, `{"a":"a","b":"b"}`)
	})
}
