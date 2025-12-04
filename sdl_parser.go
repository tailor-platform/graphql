package graphql

import (
	"fmt"

	"github.com/tailor-platform/graphql/language/ast"
	"github.com/tailor-platform/graphql/language/parser"
)

const (
	TypeNameID       = "ID"
	TypeNameString   = "String"
	TypeNameBoolean  = "Boolean"
	TypeNameBool     = "Bool"
	TypeNameInt      = "Int"
	TypeNameInteger  = "Integer"
	TypeNameFloat    = "Float"
	TypeNameDateTime = "Datetime"
)

type GraphqlParser struct {
	typeMap           map[string]Type
	typeFieldMap      map[string]Fields
	unionTypeMap      map[string][]*Object
	inputFieldMap     map[string]InputObjectConfigFieldMap
	fieldConfigArgMap map[string]FieldConfigArgument
	fieldDirectiveMap map[string]FieldDirectives
	directiveMap      map[string]*Directive
	sdlResolver       SDLResolver
}

func NewGraphqlParser(sdlResolver SDLResolver) GraphqlParser {
	return GraphqlParser{
		typeMap:           make(map[string]Type),
		typeFieldMap:      make(map[string]Fields),
		unionTypeMap:      make(map[string][]*Object),
		inputFieldMap:     make(map[string]InputObjectConfigFieldMap),
		fieldConfigArgMap: make(map[string]FieldConfigArgument),
		fieldDirectiveMap: make(map[string]FieldDirectives),
		directiveMap:      make(map[string]*Directive),
		sdlResolver:       sdlResolver,
	}
}

type TypeNameMapOption func(_type string) Type

func (g *GraphqlParser) astAtInputType(inputType ast.Type) (Type, error) {
	switch t := inputType.(type) {
	case *ast.Named:
		switch t.Name.Value {
		case TypeNameID:
			return ID, nil
		case TypeNameInteger, TypeNameInt:
			return Int, nil
		case TypeNameFloat:
			return Float, nil
		case TypeNameBoolean, TypeNameBool:
			return Boolean, nil
		case TypeNameDateTime:
			return DateTime, nil
		default:
			type_, err := g.asType(t)
			if err != nil {
				return nil, err
			}
			switch type_.(type) {
			case *InputObject, *Scalar, *List, *NonNull, *Union:
				return type_, nil
			default:
				return nil, fmt.Errorf("%s is not use InputType", t.Name.Value)
			}
		}
	case *ast.NonNull:
		type_, err := g.astAtInputType(t.Type)
		if err != nil {
			return nil, err
		}
		return NewNonNull(type_), nil
	case *ast.List:
		type_, err := g.astAtInputType(t.Type)
		if err != nil {
			return nil, err
		}
		return NewList(type_), nil
	default:
		return nil, fmt.Errorf("%s is not use InputType", inputType.String())
	}
}

func asString(value *ast.StringValue) string {
	val := ""
	if value != nil {
		val = value.Value
	}
	return val
}

func (g *GraphqlParser) asObjectDirectives(directives []*ast.Directive) (FieldDirectives, error) {
	var fieldDirectives FieldDirectives
	for _, d := range directives {
		if directive, ok := g.directiveMap[d.Name.Value]; ok {
			fieldDirectives = append(fieldDirectives, &ObjectDirective{
				Directive: directive,
			})
		} else {
			return nil, fmt.Errorf("directive %s is not found", directive.Name)
		}
	}
	return fieldDirectives, nil
}

func (g *GraphqlParser) asFieldConfigArgs(args []*ast.InputValueDefinition) (FieldConfigArgument, error) {
	fieldConfigArg := make(FieldConfigArgument)
	for _, arg := range args {
		type_, err := g.asType(arg.Type)
		if err != nil {
			return nil, err
		}
		var defaultValue interface{}
		if arg.DefaultValue != nil {
			defaultValue = arg.DefaultValue.GetValue()
		}
		fieldConfigArg[arg.Name.Value] = &ArgumentConfig{
			Type:         type_,
			DefaultValue: defaultValue,
			Description:  asString(arg.Description),
		}
	}
	return fieldConfigArg, nil
}

func (g *GraphqlParser) asType(type_ ast.Type) (Type, error) {
	switch t := type_.(type) {
	case *ast.Named:
		switch t.Name.Value {
		case TypeNameString:
			return String, nil
		case TypeNameID:
			return ID, nil
		case TypeNameBoolean, TypeNameBool:
			return Boolean, nil
		case TypeNameInt, TypeNameInteger:
			return Int, nil
		case TypeNameFloat:
			return Float, nil
		default:
			if tp, ok := g.typeMap[t.Name.Value]; ok {
				return tp, nil
			} else {
				return nil, fmt.Errorf("type %s is not found", t.Name.Value)
			}
		}
	case *ast.NonNull:
		tt, err := g.asType(t.Type)
		if err != nil {
			return nil, err
		}
		return NewNonNull(tt), nil
	case *ast.List:
		tt, err := g.asType(t.Type)
		if err != nil {
			return nil, err
		}
		return NewList(tt), nil
	default:
		return nil, fmt.Errorf("type %s is not found", t.String())
	}
}

func (g *GraphqlParser) asLocationStrins(names []*ast.Name) []string {
	locations := make([]string, len(names))
	for _, name := range names {
		locations = append(locations, name.Value)
	}
	return locations
}

func unisonResolver(p ResolveTypeParams) *Object {
	return nil
}

func (g *GraphqlParser) AstAsSchemaConfig(nodes []ast.Node, opts ...TypeNameMapOption) (*SchemaConfig, error) {
	var checkHasFields []ast.Node
	for _, def := range nodes {
		switch o := def.(type) {
		case *ast.ScalarDefinition:
			name := o.Name.Value
			for _, opt := range opts {
				if t := opt(name); t != nil {
					g.typeMap[name] = t
					goto skip
				}
			}
			g.typeMap[name] = NewScalar(ScalarConfig{
				Name:        name,
				Description: asString(o.Description),
				Serialize: func(value interface{}) interface{} {
					return nil
				},
			})
		case *ast.EnumDefinition:
			name := o.Name.Value
			values := make(EnumValueConfigMap)
			for _, v := range o.Values {
				values[v.Name.Value] = &EnumValueConfig{
					Value:       v.Name.Value,
					Description: asString(v.Description),
				}
			}
			g.typeMap[name] = NewEnum(EnumConfig{
				Name:        name,
				Description: asString(o.Description),
				Values:      values,
			})
		case *ast.UnionDefinition:
			name := o.Name.Value
			g.unionTypeMap[name] = make([]*Object, len(o.Types))
			g.typeMap[name] = NewUnion(UnionConfig{
				Name:        o.Name.Value,
				Description: asString(o.Description),
				Types:       g.unionTypeMap[name],
				ResolveType: unisonResolver,
			})
			checkHasFields = append(checkHasFields, o)
		case *ast.ObjectDefinition:
			name := o.Name.Value
			g.typeFieldMap[name] = Fields{}
			g.fieldDirectiveMap[name] = FieldDirectives{}
			g.typeMap[name] = NewObject(ObjectConfig{
				Name:        name,
				Description: asString(o.Description),
				Fields:      g.typeFieldMap[name],
				Directives:  g.fieldDirectiveMap[name],
			})
			if len(o.Fields) > 0 {
				checkHasFields = append(checkHasFields, o)
			}
		case *ast.TypeExtensionDefinition:
			name := o.Definition.Name.Value
			if _, ok := g.typeMap[name]; !ok {
				return nil, fmt.Errorf("type %s is not found", name)
			}
			if len(o.Definition.Fields) > 0 {
				checkHasFields = append(checkHasFields, o.Definition)
			}
		case *ast.InputObjectDefinition:
			name := o.Name.Value
			g.inputFieldMap[name] = InputObjectConfigFieldMap{}
			g.typeMap[name] = NewInputObject(InputObjectConfig{
				Name:        o.Name.Value,
				Fields:      g.inputFieldMap[name],
				Description: asString(o.Description),
			})
			if len(o.Fields) > 0 {
				checkHasFields = append(checkHasFields, o)
			}
		case *ast.DirectiveDefinition:
			name := o.Name.Value
			g.fieldConfigArgMap[name] = FieldConfigArgument{}
			locations := g.asLocationStrins(o.Locations)
			g.directiveMap[name] = NewDirective(DirectiveConfig{
				Name:        name,
				Args:        g.fieldConfigArgMap[name],
				Description: asString(o.Description),
				Locations:   locations,
			})
			if len(o.Arguments) > 0 {
				checkHasFields = append(checkHasFields, o)
			}
		}
	skip:
	}
	for _, def := range checkHasFields {
		switch o := def.(type) {
		case *ast.ObjectDefinition:
			name := o.Name.Value
			for _, field := range o.Fields {
				fieldName := field.Name.Value
				if t, ok := g.typeFieldMap[name]; ok {
					type_, err := g.asType(field.Type)
					if err != nil {
						return nil, err
					}
					args, err := g.asFieldConfigArgs(field.Arguments)
					if err != nil {
						return nil, err
					}
					directives, err := g.asObjectDirectives(field.Directives)
					if err != nil {
						return nil, err
					}
					t[fieldName] = &Field{
						Name:        fieldName,
						Args:        args,
						Type:        type_,
						Directives:  directives,
						Description: asString(field.Description),
						Resolve:     g.sdlResolver(name, fieldName),
					}
				} else {
					return nil, fmt.Errorf("type %s is not found", fieldName)
				}
			}
		case *ast.UnionDefinition:
			name := o.Name.Value
			for i, tp := range o.Types {
				type_, err := g.asType(tp)
				if err != nil {
					return nil, err
				}
				g.unionTypeMap[name][i] = type_.(*Object)
			}
		case *ast.InputObjectDefinition:
			name := o.Name.Value
			for _, field := range o.Fields {
				fieldName := field.Name.Value
				if i, ok := g.inputFieldMap[name]; ok {
					type_, err := g.asType(field.Type)
					if err != nil {
						return nil, err
					}
					i[fieldName] = &InputObjectFieldConfig{
						Type:        type_,
						Description: asString(field.Description),
					}
				} else {
					return nil, fmt.Errorf("input type %s is not found", fieldName)
				}
			}
		case *ast.DirectiveDefinition:
			name := o.Name.Value
			if f, ok := g.fieldConfigArgMap[name]; ok {
				for _, arg := range o.Arguments {
					argName := arg.Name.Value
					type_, err := g.asType(arg.Type)
					if err != nil {
						return nil, err
					}
					f[argName] = &ArgumentConfig{
						Type:         type_,
						DefaultValue: arg.DefaultValue.GetValue(),
						Description:  asString(arg.Description),
					}
				}
			} else {
				return nil, fmt.Errorf("directive %s is not found", name)
			}
		default:
			return nil, fmt.Errorf("%+v", o)
		}
	}
	schemaConfig := SchemaConfig{}
	for _, type_ := range g.typeMap {
		schemaConfig.Types = append(schemaConfig.Types, type_)
	}
	for _, directive := range g.directiveMap {
		schemaConfig.Directives = append(schemaConfig.Directives, directive)
	}

	if query, ok := g.typeMap["Query"].(*Object); ok {
		schemaConfig.Query = query
	}
	if mutation, ok := g.typeMap["Mutation"].(*Object); ok {
		schemaConfig.Mutation = mutation
	}
	if subscription, ok := g.typeMap["Subscription"].(*Object); ok {
		schemaConfig.Subscription = subscription
	}
	return &schemaConfig, nil
}

type SDLResolver func(typeName string, fieldName string) FieldResolveFn

func ParseSDL(sdl string, sdlResolver SDLResolver) (*Schema, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: sdl,
	})
	if err != nil {
		return nil, err
	}
	g := NewGraphqlParser(sdlResolver)
	schemaConfig, err := g.AstAsSchemaConfig(doc.Definitions)
	if err != nil {
		return nil, err
	}
	schema, err := NewSchema(*schemaConfig)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}
