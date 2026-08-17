package graphql_test

import (
	"testing"

	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/gqlerrors"
	"github.com/tailor-platform/graphql/testutil"
)

func TestValidate_VariableDefaultValuesOfCorrectType_VariablesWithNoDefaultValues(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query NullableValues($a: Int, $b: String, $c: ComplexInput) {
        dog { name }
      }
    `)
}
func TestValidate_VariableDefaultValuesOfCorrectType_RequiredVariablesWithoutDefaultValues(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query RequiredValues($a: Int!, $b: String!) {
        dog { name }
      }
    `)
}
func TestValidate_VariableDefaultValuesOfCorrectType_VariablesWithValidDefaultValues(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query WithDefaultValues(
        $a: Int = 1,
        $b: String = "ok",
        $c: ComplexInput = { requiredField: true, intField: 3 }
      ) {
        dog { name }
      }
    `)
}

// Spec §6.1.2 applies a variable's default before checking the non-null
// requirement, so a non-null variable may declare one and the default is
// reached. This test asserted the opposite before the coercion fix; the
// rejection now lives behind NonSpecArgumentHandling and is covered by
// TestValidate_VariableDefaultValuesOfCorrectType_NonSpecRejectsNonNullVariableDefault.
func TestValidate_VariableDefaultValuesOfCorrectType_NonNullVariablesWithDefaultValues(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query NonNullDefaultValues($a: Int! = 3, $b: String! = "default") {
        dog { name }
      }
    `)
}
func TestValidate_VariableDefaultValuesOfCorrectType_VariablesWithInvalidDefaultValues(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query InvalidDefaultValues(
        $a: Int = "one",
        $b: String = 4,
        $c: ComplexInput = "notverycomplex"
      ) {
        dog { name }
      }
    `,
		[]gqlerrors.FormattedError{
			testutil.RuleError(`Variable "$a" has invalid default value: "one".`+
				"\nExpected type \"Int\", found \"one\".",
				3, 19),
			testutil.RuleError(`Variable "$b" has invalid default value: 4.`+
				"\nExpected type \"String\", found 4.",
				4, 22),
			testutil.RuleError(
				`Variable "$c" has invalid default value: "notverycomplex".`+
					"\nExpected \"ComplexInput\", found not an object.",
				5, 28),
		})
}
func TestValidate_VariableDefaultValuesOfCorrectType_ComplexVariablesMissingRequiredField(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query MissingRequiredField($a: ComplexInput = {intField: 3}) {
        dog { name }
      }
    `,
		[]gqlerrors.FormattedError{
			testutil.RuleError(
				`Variable "$a" has invalid default value: {intField: 3}.`+
					"\nIn field \"requiredField\": Expected \"Boolean!\", found null.",
				2, 53),
		})
}
func TestValidate_VariableDefaultValuesOfCorrectType_ListVariablesWithInvalidItem(t *testing.T) {
	testutil.ExpectFailsRule(t, graphql.DefaultValuesOfCorrectTypeRule, `
      query InvalidItem($a: [String] = ["one", 2]) {
        dog { name }
      }
    `,
		[]gqlerrors.FormattedError{
			testutil.RuleError(
				`Variable "$a" has invalid default value: ["one", 2].`+
					"\nIn element #1: Expected type \"String\", found 2.",
				2, 40),
		})
}

func TestValidate_VariableDefaultValuesOfCorrectType_InvalidNonNull(t *testing.T) {
	testutil.ExpectPassesRule(t, graphql.DefaultValuesOfCorrectTypeRule, `query($g:e!){a}`)
}

// One nullable argument, so a non-null variable can be used at it, under
// whichever mode the caller asks for.
func nonNullVariableDefaultSchema(t *testing.T, nonSpec bool) graphql.Schema {
	t.Helper()
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"f": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"a": &graphql.ArgumentConfig{Type: graphql.String},
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

// Spec §6.1.2 applies a variable's default before checking the non-null
// requirement, and nothing in the Validation section forbids a non-null variable
// from declaring one.
func TestValidate_VariableDefaultValuesOfCorrectType_NonNullVariableMayHaveDefault(t *testing.T) {
	schema := nonNullVariableDefaultSchema(t, false)
	testutil.ExpectPassesRuleWithSchema(t, &schema, graphql.DefaultValuesOfCorrectTypeRule, `
      query Probe($a: String! = "VARDEF") {
        f(a: $a)
      }
    `)
}

// NonSpecArgumentHandling keeps rejecting it.
func TestValidate_VariableDefaultValuesOfCorrectType_NonSpecRejectsNonNullVariableDefault(t *testing.T) {
	schema := nonNullVariableDefaultSchema(t, true)
	testutil.ExpectFailsRuleWithSchema(t, &schema, graphql.DefaultValuesOfCorrectTypeRule, `
      query Probe($a: String! = "VARDEF") {
        f(a: $a)
      }
    `, []gqlerrors.FormattedError{
		testutil.RuleError(`Variable "$a" of type "String!" is required and will not use the default value. Perhaps you meant to use type "String".`, 2, 33),
	})
}
