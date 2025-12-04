package graphql

import (
	"github.com/stretchr/testify/assert"
	"github.com/tailor-platform/graphql/language/printer"
	"testing"
)

func TestBuildSDL(t *testing.T) {
	var types []Type
	enum1 := NewEnum(EnumConfig{
		Name:        "Enum1",
		Description: "enum description",
		Values: EnumValueConfigMap{
			"val1": &EnumValueConfig{
				Value:       "val1",
				Description: "desc val1",
			},
			"val2": &EnumValueConfig{
				Value:       "val2",
				Description: "desc val2",
			},
		},
	})

	keyDirective := NewDirective(DirectiveConfig{
		Name:        "key",
		Description: "key description",
		Locations: []string{
			DirectiveLocationField,
			DirectiveLocationFragmentSpread,
			DirectiveLocationInlineFragment,
		},
		Args: FieldConfigArgument{
			"fields": &ArgumentConfig{
				Type:        String,
				Description: "ddd",
			},
		},
	})

	serviceType := NewObject(ObjectConfig{
		Name: "_Service",
		Fields: Fields{
			"_sdl": &Field{
				Type:        String,
				Description: "gdl query desc",
			},
		},
	})

	queryType := NewObject(ObjectConfig{
		Name:        "Query",
		Description: "Query desc",
		Fields: Fields{
			"me": &Field{
				Type:        String,
				Description: "me query desc",
			},
			"_service": &Field{
				Type:        serviceType,
				Description: "for federation",
			},
		},
	})

	testType1 := NewObject(ObjectConfig{
		Name:        "TestType1",
		Description: "TestType1 desc",
		Fields: Fields{
			"str": &Field{
				Type:        String,
				Description: "TestType1 string desc",
				Directives: FieldDirectives{
					&ObjectDirective{
						Directive: ExternalDirective,
					},
				},
			},
			"enum1": &Field{
				Type: enum1,
			},
			"arr": &Field{
				Type:        NewList(Int),
				Description: "type1 arr[int] desc",
			},
		},
		Directives: []*ObjectDirective{
			{
				Directive: keyDirective,
				Args: []ObjectDirectiveArg{
					{
						Name:  "fields",
						Value: "id",
					},
				},
			},
		},
		Extend: true,
	})

	testType2 := NewObject(ObjectConfig{
		Name:        "TestType2",
		Description: "type of TestType2",
		Fields: Fields{
			"int": &Field{
				Type:        Int,
				Description: "TestType1 int desc",
			},
			"nest": &Field{
				Type:        testType1,
				Description: "TestType2 nest desc",
			},
		},
		Extend: true,
	})
	pingOutputType := NewObject(ObjectConfig{
		Name:        "PingOutputType",
		Description: "type of PingOutputType",
		Fields: Fields{
			"int": &Field{
				Type:        Int,
				Description: "TestType1 int desc",
			},
			"nest": &Field{
				Type:        testType1,
				Description: "TestType2 nest desc",
			},
		},
	})

	inputType1 := NewInputObject(InputObjectConfig{
		Name:        "InputType1",
		Description: "type of InputType1",
		Fields: InputObjectConfigFieldMap{
			"int": &InputObjectFieldConfig{
				Type:        Int,
				Description: "inputType1 int desc",
			},
			"arr": &InputObjectFieldConfig{
				Type:        NewList(Int),
				Description: "inputType1 arr[int] desc",
			},
		},
	})
	inputType2 := NewInputObject(InputObjectConfig{
		Name:        "InputType2",
		Description: "type of InputType2",
		Fields: InputObjectConfigFieldMap{
			"int": &InputObjectFieldConfig{
				Type:        Int,
				Description: "inputType2 int desc",
			},
			"nestedInput": &InputObjectFieldConfig{
				Type: inputType1,
			},
		},
	})

	datetime := NewScalar(ScalarConfig{
		Name:        "Datetime",
		Description: "type of Datetime",
		Serialize: func(value interface{}) interface{} {
			return nil
		},
	})

	types = append(types, serviceType)
	types = append(types, enum1)
	types = append(types, testType1)
	types = append(types, testType2)
	types = append(types, inputType1)
	types = append(types, inputType2)
	types = append(types, pingOutputType)
	types = append(types, datetime)

	t.Run("objectAsNode exclude Query._service", func(t *testing.T) {
		sdl := printer.Print(objectAsNode(queryType, &SDLExportOptions{
			ExcludeQueryService: false,
		}))
		assert.Contains(t, sdl, `"""Query desc"""
type Query {`)
		assert.Contains(t, sdl, `"""for federation"""
  _service: _Service`)
		assert.Contains(t, sdl, `"""me query desc"""
  me: String`)
		sdl = printer.Print(objectAsNode(queryType, &SDLExportOptions{
			ExcludeQueryService: true,
		}))
		assert.Contains(t, sdl, `"""Query desc"""
type Query {`)
		assert.Contains(t, sdl, `"""me query desc"""
  me: String`)
		assert.NotContains(t, sdl, `"""for federation"""
  _service: String`)
	})

	defaultOptions := SDLExportOptions{}
	t.Run("objectAsNode", func(t *testing.T) {
		sdl := printer.Print(objectAsNode(testType1, &defaultOptions))
		assert.True(t, testType1.extend)
		assert.Contains(t, sdl, `"""TestType1 desc"""
extend type TestType1 @key(fields: "id") {`)
		assert.Contains(t, sdl, `"""TestType1 string desc"""
  str: String @external`)
		assert.Contains(t, sdl, `enum1: Enum1`)
		assert.Contains(t, sdl, `"""type1 arr[int] desc"""
  arr: [Int]`)
	})

	t.Run("inputObjectAsNode", func(t *testing.T) {
		sdl := printer.Print(inputObjectAsNode(inputType1))
		assert.Contains(t, sdl, `"""type of InputType1"""
input InputType1 {`)
		assert.Contains(t, sdl, `"""inputType1 int desc"""
  int: Int`)
		assert.Contains(t, sdl, `"""inputType1 arr[int] desc"""
  arr: [Int]`)
	})

	t.Run("enumAsNode", func(t *testing.T) {
		sdl := printer.Print(enumAsNode(enum1))
		assert.Contains(t, sdl, `"""enum description"""
enum Enum1 {`)
		assert.Contains(t, sdl, `"""desc val1"""
  val1`)
		assert.Contains(t, sdl, `"""desc val2"""
  val2`)
	})

	t.Run("scalarAsNode", func(t *testing.T) {
		sdl := printer.Print(scalarAsNode(datetime))
		t.Log(sdl)
		assert.Equal(t, sdl, `"""type of Datetime"""
scalar Datetime`)
	})

	schemaConfig := SchemaConfig{
		Types: types,
		Query: NewObject(ObjectConfig{Name: "Query", Fields: Fields{
			"ping": &Field{
				Type: pingOutputType,
			},
			"_service": &Field{
				Type:        serviceType,
				Description: "for federation",
			},
		}}),
		Mutation: NewObject(ObjectConfig{Name: "Mutation", Fields: Fields{
			"ping": &Field{
				Type: pingOutputType,
				Args: FieldConfigArgument{
					"q": &ArgumentConfig{
						Type: Int,
					},
				},
			},
		}}),
	}
	schema, err := NewSchema(schemaConfig)
	assert.NoError(t, err)

	sdl := BuildSDL(schema, &SDLExportOptions{
		ExcludeDoubleUnderscorePrefix: true,
		IncludeBasicScalar:            false,
		ExcludeQueryService:           true,
	})

	assert.NotContains(t, sdl, "scalar String")
	assert.NotContains(t, sdl, "scalar Boolean")
	assert.NotContains(t, sdl, "scalar Int")
	assert.NotContains(t, sdl, "__")
	assert.NotContains(t, sdl, "_Service")
	assert.NotContains(t, sdl, "_service")
	assert.NotContains(t, sdl, "_sdl")

	sdl = BuildSDL(schema, &SDLExportOptions{
		ExcludeDoubleUnderscorePrefix: true,
		IncludeBasicScalar:            true,
		ExcludeQueryService:           false,
	})

	assert.Contains(t, sdl, "scalar String")
	assert.Contains(t, sdl, "scalar Boolean")
	assert.Contains(t, sdl, "scalar Int")
	assert.Contains(t, sdl, "_Service")
	assert.Contains(t, sdl, "_service")
	assert.Contains(t, sdl, "_sdl")

}
