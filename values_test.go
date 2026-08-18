package graphql

import (
	"math"
	"reflect"
	"testing"
)

func TestIsIterable(t *testing.T) {
	if !isIterable([]int{}) {
		t.Fatal("expected isIterable to return true for a slice, got false")
	}
	if !isIterable([]int{}) {
		t.Fatal("expected isIterable to return true for an array, got false")
	}
	if isIterable(1) {
		t.Fatal("expected isIterable to return false for an int, got true")
	}
	if isIterable(nil) {
		t.Fatal("expected isIterable to return false for nil, got true")
	}
}

func Test_coerceValue(t *testing.T) {
	t.Parallel()

	type input struct {
		ttype Input
		value any
	}
	testCases := map[string]struct {
		input    input
		expected any
	}{
		"null Input Object is coerced to nil": {
			input: input{
				ttype: NewInputObject(InputObjectConfig{
					Name: "InputObject",
				}),
				value: nil,
			},
			expected: nil,
		},
		"null field in Input Object is not omitted, and coerced to nil": {
			input: input{
				ttype: NewInputObject(InputObjectConfig{
					Name: "InputObject",
					Fields: InputObjectConfigFieldMap{
						"string": &InputObjectFieldConfig{
							Type: String,
						},
					},
				}),
				value: map[string]any{"string": nil},
			},
			expected: map[string]any{"string": nil},
		},
	}

	// None of these cases involve a default value, so both coercion modes must
	// agree on all of them.
	for name, tc := range testCases {
		name, tc := name, tc
		for _, nonSpec := range []bool{false, true} {
			nonSpec := nonSpec
			mode := "spec"
			if nonSpec {
				mode = "nonSpec"
			}
			t.Run(name+"/"+mode, func(t *testing.T) {
				t.Parallel()

				got := coerceValue(tc.input.ttype, tc.input.value, nonSpec)
				if !reflect.DeepEqual(tc.expected, got) {
					t.Errorf("unexpected result, expected: %v, got: %v", tc.expected, got)
				}
			})
		}
	}
}

// Spec §3.10 / §5.6.4: an input field is required only when its type is non-null
// AND it declares no default value. Pinned at the function level so the three
// states stay distinct: absent-with-default, absent-without-default, explicit null.
func Test_isValidInputValue_NonNullFieldWithDefault(t *testing.T) {
	withDefault := NewInputObject(InputObjectConfig{
		Name: "WithDefault",
		Fields: InputObjectConfigFieldMap{
			"a": &InputObjectFieldConfig{Type: NewNonNull(String), DefaultValue: "FIELDDEF"},
		},
	})
	withoutDefault := NewInputObject(InputObjectConfig{
		Name: "WithoutDefault",
		Fields: InputObjectConfigFieldMap{
			"a": &InputObjectFieldConfig{Type: NewNonNull(String)},
		},
	})

	if isValid, messages := isValidInputValue(map[string]interface{}{}, withDefault, false); !isValid {
		t.Errorf("spec mode: expected an omitted field with a default to be valid, got: %v", messages)
	}
	if isValid, _ := isValidInputValue(map[string]interface{}{}, withDefault, true); isValid {
		t.Error("non-spec mode: expected an omitted non-null field to be invalid")
	}
	for _, nonSpec := range []bool{false, true} {
		if isValid, _ := isValidInputValue(map[string]interface{}{"a": nil}, withDefault, nonSpec); isValid {
			t.Errorf("expected an explicit null to be invalid (nonSpec=%v)", nonSpec)
		}
		if isValid, _ := isValidInputValue(map[string]interface{}{}, withoutDefault, nonSpec); isValid {
			t.Errorf("expected an omitted field without a default to be invalid (nonSpec=%v)", nonSpec)
		}
	}
}

// A default that is nullish but not nil — a typed nil pointer, say — is not a
// default coercion can substitute: coerceValue and valueFromAST both skip it
// under isNullish. Validation has to agree, otherwise a non-null field passes
// validation and then goes missing from the coerced map.
func Test_isValidInputValue_NullishDefaultIsNotADefault(t *testing.T) {
	var nilString *string
	for _, tc := range []struct {
		name       string
		defaultVal interface{}
		wantValid  bool
	}{
		{"usable default", "FIELDDEF", true},
		{"no default", nil, false},
		{"typed nil pointer", nilString, false},
		{"NaN", math.NaN(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := NewInputObject(InputObjectConfig{
				Name: "NullishDefault" + tc.name,
				Fields: InputObjectConfigFieldMap{
					"a": &InputObjectFieldConfig{Type: NewNonNull(String), DefaultValue: tc.defaultVal},
				},
			})
			isValid, messages := isValidInputValue(map[string]interface{}{}, in, false)
			if isValid != tc.wantValid {
				t.Errorf("isValid = %v, want %v (messages: %v)", isValid, tc.wantValid, messages)
			}
		})
	}
}
