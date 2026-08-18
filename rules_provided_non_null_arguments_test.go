package graphql_test

import (
	"testing"

	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/gqlerrors"
	"github.com/tailor-platform/graphql/testutil"
)

func TestValidate_ProvidedNonNullArguments_IgnoresUnknownArguments(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
      {
        dog {
          isHousetrained(unknownArgument: true)
        }
      }
    `)
}

func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_ArgOnOptionalArg(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          dog {
            isHousetrained(atOtherHomes: true)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_NoArgOnOptionalArg(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          dog {
            isHousetrained
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_MultipleArgs(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleReqs(req1: 1, req2: 2)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_MultipleArgsReverseOrder(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleReqs(req2: 2, req1: 1)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_NoArgsOnMultipleOptional(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOpts
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_OneArgOnMultipleOptional(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOpts(opt1: 1)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_SecondArgOnMultipleOptional(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOpts(opt2: 1)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_MultipleReqsOnMixedList(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOptAndReq(req1: 3, req2: 4)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_MultipleReqsAndOneOptOnMixedList(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOptAndReq(req1: 3, req2: 4, opt1: 5)
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_ValidNonNullableValue_AllReqsAndOptsOnMixedList(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleOptAndReq(req1: 3, req2: 4, opt1: 5, opt2: 6)
          }
        }
    `)
}

func TestValidate_ProvidedNonNullArguments_InvalidNonNullableValue_MissingOneNonNullableArgument(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleReqs(req2: 2)
          }
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "multipleReqs" argument "req1" of type "Int!" is required but not provided.`, 4, 13),
	})
}
func TestValidate_ProvidedNonNullArguments_InvalidNonNullableValue_MissingMultipleNonNullableArguments(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleReqs
          }
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "multipleReqs" argument "req1" of type "Int!" is required but not provided.`, 4, 13),
		testutil.RuleError(`Field "multipleReqs" argument "req2" of type "Int!" is required but not provided.`, 4, 13),
	})
}
func TestValidate_ProvidedNonNullArguments_InvalidNonNullableValue_IncorrectValueAndMissingArgument(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          complicatedArgs {
            multipleReqs(req1: "one")
          }
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "multipleReqs" argument "req2" of type "Int!" is required but not provided.`, 4, 13),
	})
}

func TestValidate_ProvidedNonNullArguments_DirectiveArguments_IgnoresUnknownDirectives(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          dog @unknown
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_DirectiveArguments_WithDirectivesOfValidTypes(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          dog @include(if: true) {
            name
          }
          human @skip(if: false) {
            name
          }
        }
    `)
}
func TestValidate_ProvidedNonNullArguments_DirectiveArguments_WithDirectiveWithMissingTypes(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.ProvidedNonNullArgumentsRule, `
        {
          dog @include {
            name @skip
          }
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Directive "@include" argument "if" of type "Boolean!" is required but not provided.`, 3, 15),
		testutil.RuleError(`Directive "@skip" argument "if" of type "Boolean!" is required but not provided.`, 4, 18),
	})
}

// One field whose non-null argument declares a default value, under whichever
// mode the caller asks for.
func nonNullArgWithDefaultSchema(t *testing.T, nonSpec bool) graphql.Schema {
	t.Helper()
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"fieldWithDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"arg": &graphql.ArgumentConfig{
							Type:         graphql.NewNonNull(graphql.Boolean),
							DefaultValue: true,
						},
					},
				},
			},
		}),
		NonSpecArgumentHandling: nonSpec,
	})
	if err != nil {
		t.Fatalf("Unexpected error, got: %v", err)
	}
	return schema
}

// Spec §5.4.2.1: "An argument is required if the argument type is non-null and
// does not have a default value. Otherwise, the argument is optional."
// See graphql-go/graphql#739.
func TestValidate_ProvidedNonNullArguments_FieldArguments_NoErrorOnNonNullArgumentWithDefaultValue(t *testing.T) {
	schema := nonNullArgWithDefaultSchema(t, false)
	testutil.ExpectPassesRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          fieldWithDefault
        }
    `)
}

// NonSpecArgumentHandling restores the older, stricter reading, under which a
// non-null argument is required whether or not it declares a default.
func TestValidate_ProvidedNonNullArguments_FieldArguments_NonSpecErrorsOnNonNullArgumentWithDefaultValue(t *testing.T) {
	schema := nonNullArgWithDefaultSchema(t, true)
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          fieldWithDefault
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "fieldWithDefault" argument "arg" of type "Boolean!" is required but not provided.`, 3, 11),
	})
}

func TestValidate_ProvidedNonNullArguments_FieldArguments_StillErrorsOnNonNullArgumentWithoutDefaultValue(t *testing.T) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"fieldWithoutDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"arg": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Boolean),
						},
					},
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unexpected error, got: %v", err)
	}
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          fieldWithoutDefault
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "fieldWithoutDefault" argument "arg" of type "Boolean!" is required but not provided.`, 3, 11),
	})
}

// One directive whose non-null argument declares a default value, under
// whichever mode the caller asks for.
func nonNullDirectiveArgWithDefaultSchema(t *testing.T, nonSpec bool) graphql.Schema {
	t.Helper()
	deferDirective := graphql.NewDirective(graphql.DirectiveConfig{
		Name: "defer",
		Locations: []string{
			graphql.DirectiveLocationFragmentSpread,
			graphql.DirectiveLocationInlineFragment,
		},
		Args: graphql.FieldConfigArgument{
			"if": &graphql.ArgumentConfig{
				Type:         graphql.NewNonNull(graphql.Boolean),
				DefaultValue: true,
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"a": &graphql.Field{Type: graphql.String},
			},
		}),
		Directives:              []*graphql.Directive{deferDirective},
		NonSpecArgumentHandling: nonSpec,
	})
	if err != nil {
		t.Fatalf("Unexpected error, got: %v", err)
	}
	return schema
}

func TestValidate_ProvidedNonNullArguments_DirectiveArguments_NoErrorOnNonNullArgumentWithDefaultValue(t *testing.T) {
	schema := nonNullDirectiveArgWithDefaultSchema(t, false)
	testutil.ExpectPassesRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          ... on Query @defer {
            a
          }
        }
    `)
}

func TestValidate_ProvidedNonNullArguments_DirectiveArguments_NonSpecErrorsOnNonNullArgumentWithDefaultValue(t *testing.T) {
	schema := nonNullDirectiveArgWithDefaultSchema(t, true)
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          ... on Query @defer {
            a
          }
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Directive "@defer" argument "if" of type "Boolean!" is required but not provided.`, 3, 24),
	})
}

// A default coercion will not substitute — a typed nil pointer is nullish —
// does not make a non-null argument optional.
func TestValidate_ProvidedNonNullArguments_FieldArguments_NullishDefaultKeepsArgumentRequired(t *testing.T) {
	var nilString *string
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"fieldWithNullishDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"arg": &graphql.ArgumentConfig{
							Type:         graphql.NewNonNull(graphql.Boolean),
							DefaultValue: nilString,
						},
					},
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unexpected error, got: %v", err)
	}
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.ProvidedNonNullArgumentsRule, `
        {
          fieldWithNullishDefault
        }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Field "fieldWithNullishDefault" argument "arg" of type "Boolean!" is required but not provided.`, 3, 11),
	})
}
