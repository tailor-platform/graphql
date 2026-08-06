package graphql

import (
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
		for _, specCompliant := range []bool{false, true} {
			specCompliant := specCompliant
			mode := "legacy"
			if specCompliant {
				mode = "spec"
			}
			t.Run(name+"/"+mode, func(t *testing.T) {
				t.Parallel()

				got := coerceValue(tc.input.ttype, tc.input.value, specCompliant)
				if !reflect.DeepEqual(tc.expected, got) {
					t.Errorf("unexpected result, expected: %v, got: %v", tc.expected, got)
				}
			})
		}
	}
}
