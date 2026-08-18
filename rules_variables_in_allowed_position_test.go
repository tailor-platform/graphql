package graphql_test

import (
	"testing"

	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/gqlerrors"
	"github.com/tailor-platform/graphql/testutil"
)

func TestValidate_VariablesInAllowedPosition_BooleanToBoolean(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($booleanArg: Boolean)
      {
        complicatedArgs {
          booleanArgField(booleanArg: $booleanArg)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_BooleanToBooleanWithinFragment(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      fragment booleanArgFrag on ComplicatedArgs {
        booleanArgField(booleanArg: $booleanArg)
      }
      query Query($booleanArg: Boolean)
      {
        complicatedArgs {
          ...booleanArgFrag
        }
      }
    `)
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($booleanArg: Boolean)
      {
        complicatedArgs {
          ...booleanArgFrag
        }
      }
      fragment booleanArgFrag on ComplicatedArgs {
        booleanArgField(booleanArg: $booleanArg)
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_NonNullableBooleanToBoolean(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($nonNullBooleanArg: Boolean!)
      {
        complicatedArgs {
          booleanArgField(booleanArg: $nonNullBooleanArg)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_NonNullableBooleanToBooleanWithinFragment(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      fragment booleanArgFrag on ComplicatedArgs {
        booleanArgField(booleanArg: $nonNullBooleanArg)
      }

      query Query($nonNullBooleanArg: Boolean!)
      {
        complicatedArgs {
          ...booleanArgFrag
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_IntToNonNullableIntWithDefault(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($intArg: Int = 1)
      {
        complicatedArgs {
          nonNullIntArgField(nonNullIntArg: $intArg)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_ListOfStringToListOfString(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringListVar: [String])
      {
        complicatedArgs {
          stringListArgField(stringListArg: $stringListVar)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_ListOfNonNullableStringToListOfString(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringListVar: [String!])
      {
        complicatedArgs {
          stringListArgField(stringListArg: $stringListVar)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_StringToListOfStringInItemPosition(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringVar: String)
      {
        complicatedArgs {
          stringListArgField(stringListArg: [$stringVar])
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_NonNullableStringToListOfStringInItemPosition(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringVar: String!)
      {
        complicatedArgs {
          stringListArgField(stringListArg: [$stringVar])
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_ComplexInputToComplexInput(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($complexVar: ComplexInput)
      {
        complicatedArgs {
          complexArgField(complexArg: $complexVar)
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_ComplexInputToComplexInputInFieldPosition(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($boolVar: Boolean = false)
      {
        complicatedArgs {
          complexArgField(complexArg: {requiredArg: $boolVar})
        }
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_NonNullableBooleanToNonNullableBooleanInDirective(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($boolVar: Boolean!)
      {
        dog @include(if: $boolVar)
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_NonNullableBooleanToNonNullableBooleanInDirectiveInDirectiveWithDefault(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($boolVar: Boolean = false)
      {
        dog @include(if: $boolVar)
      }
    `)
}
func TestValidate_VariablesInAllowedPosition_IntToNonNullableInt(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($intArg: Int) {
        complicatedArgs {
          nonNullIntArgField(nonNullIntArg: $intArg)
        }
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$intArg" of type "Int" used in position `+
			`expecting type "Int!".`, 2, 19, 4, 45),
	})
}
func TestValidate_VariablesInAllowedPosition_IntToNonNullableIntWithinFragment(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      fragment nonNullIntArgFieldFrag on ComplicatedArgs {
        nonNullIntArgField(nonNullIntArg: $intArg)
      }

      query Query($intArg: Int) {
        complicatedArgs {
          ...nonNullIntArgFieldFrag
        }
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$intArg" of type "Int" used in position `+
			`expecting type "Int!".`, 6, 19, 3, 43),
	})
}
func TestValidate_VariablesInAllowedPosition_IntToNonNullableIntWithinNestedFragment(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      fragment outerFrag on ComplicatedArgs {
        ...nonNullIntArgFieldFrag
      }

      fragment nonNullIntArgFieldFrag on ComplicatedArgs {
        nonNullIntArgField(nonNullIntArg: $intArg)
      }

      query Query($intArg: Int) {
        complicatedArgs {
          ...outerFrag
        }
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$intArg" of type "Int" used in position `+
			`expecting type "Int!".`, 10, 19, 7, 43),
	})
}
func TestValidate_VariablesInAllowedPosition_StringOverBoolean(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringVar: String) {
        complicatedArgs {
          booleanArgField(booleanArg: $stringVar)
        }
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$stringVar" of type "String" used in position `+
			`expecting type "Boolean".`, 2, 19, 4, 39),
	})
}
func TestValidate_VariablesInAllowedPosition_StringToListOfString(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringVar: String) {
        complicatedArgs {
          stringListArgField(stringListArg: $stringVar)
        }
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$stringVar" of type "String" used in position `+
			`expecting type "[String]".`, 2, 19, 4, 45),
	})
}
func TestValidate_VariablesInAllowedPosition_BooleanToNonNullableBooleanInDirective(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($boolVar: Boolean) {
        dog @include(if: $boolVar)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$boolVar" of type "Boolean" used in position `+
			`expecting type "Boolean!".`, 2, 19, 3, 26),
	})
}
func TestValidate_VariablesInAllowedPosition_StringToNonNullableBooleanInDirective(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.VariablesInAllowedPositionRule, `
      query Query($stringVar: String) {
        dog @include(if: $stringVar)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$stringVar" of type "String" used in position `+
			`expecting type "Boolean!".`, 2, 19, 3, 26),
	})
}

// Two non-null arguments, one declaring a default and one not, under whichever
// mode the caller asks for.
func nonNullArgDefaultPositionSchema(t *testing.T, nonSpec bool) graphql.Schema {
	t.Helper()
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"withDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"a": &graphql.ArgumentConfig{
							Type:         graphql.NewNonNull(graphql.String),
							DefaultValue: "NNDEF",
						},
					},
				},
				"withoutDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"a": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
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

// Spec §5.8.5: hasLocationDefaultValue permits a nullable variable at a non-null
// location that declares a default.
func TestValidate_VariablesInAllowedPosition_StringToNonNullStringWithArgumentDefault(t *testing.T) {
	schema := nonNullArgDefaultPositionSchema(t, false)
	testutil.ExpectPassesRuleWithSchema(t, &schema, graphql.VariablesInAllowedPositionRule, `
      query Probe($x: String) {
        withDefault(a: $x)
      }
    `)
}

// Without a default on either side the usage stays rejected.
func TestValidate_VariablesInAllowedPosition_StringToNonNullStringWithoutArgumentDefault(t *testing.T) {
	schema := nonNullArgDefaultPositionSchema(t, false)
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.VariablesInAllowedPositionRule, `
      query Probe($x: String) {
        withoutDefault(a: $x)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$x" of type "String" used in position `+
			`expecting type "String!".`, 2, 19, 3, 27),
	})
}

// NonSpecArgumentHandling only ever considered the variable's own default, so it
// keeps rejecting the usage.
func TestValidate_VariablesInAllowedPosition_NonSpecIgnoresArgumentDefault(t *testing.T) {
	schema := nonNullArgDefaultPositionSchema(t, true)
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.VariablesInAllowedPositionRule, `
      query Probe($x: String) {
        withDefault(a: $x)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$x" of type "String" used in position `+
			`expecting type "String!".`, 2, 19, 3, 24),
	})
}

// Spec §5.8.5 hasLocationDefaultValue: a default coercion will not substitute
// does not permit a nullable variable at a non-null location. Validation is
// static, so this holds whatever values the request later supplies.
func TestValidate_VariablesInAllowedPosition_NullishArgumentDefaultIsNotADefault(t *testing.T) {
	var nilString *string
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"withNullishDefault": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"a": &graphql.ArgumentConfig{
							Type:         graphql.NewNonNull(graphql.String),
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
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.VariablesInAllowedPositionRule, `
      query Probe($x: String) {
        withNullishDefault(a: $x)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$x" of type "String" used in position `+
			`expecting type "String!".`, 2, 19, 3, 31),
	})
}
