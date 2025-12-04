package graphql

import (
	"github.com/tailor-platform/graphql/language/ast"
	"github.com/tailor-platform/graphql/language/printer"
	"strings"
)

func typeAsNode(tp Input) ast.Type {
	return ast.NewNamed(&ast.Named{
		Name: ast.NewName(&ast.Name{
			Value: tp.String(),
		}),
	})
}

func argumentAsNode(args []*Argument) (arguments []*ast.InputValueDefinition) {
	for _, arg := range args {
		arguments = append(arguments, ast.NewInputValueDefinition(&ast.InputValueDefinition{
			Name: ast.NewName(&ast.Name{
				Value: arg.Name(),
			}),
			Type:        typeAsNode(arg.Type),
			Description: ast.NewStringValue(&ast.StringValue{Value: arg.Description()}),
		}))
	}
	return arguments
}

func directivesAsNode(directives []*ObjectDirective) []*ast.Directive {
	var dirs []*ast.Directive
	for _, directive := range directives {
		var args []*ast.Argument
		for _, arg := range directive.Args {
			args = append(args, ast.NewArgument(&ast.Argument{
				Name: ast.NewName(&ast.Name{
					Value: arg.Name,
				}),
				Value: ast.NewStringValue(&ast.StringValue{
					Value: arg.Value.(string),
				}),
			}))
		}
		dirs = append(dirs, ast.NewDirective(&ast.Directive{
			Name: ast.NewName(&ast.Name{
				Value: directive.Directive.Name,
			}),
			Arguments: args,
		}))
	}
	return dirs
}

func objectAsNode(o *Object, options *SDLExportOptions) ast.Node {
	var fields []*ast.FieldDefinition
	for name, object := range o.Fields() {
		if options != nil && options.ExcludeQueryService &&
			o.Name() == "Query" && name == "_service" {
			continue
		}
		var directives []*ast.Directive
		for _, d := range object.Directives {
			var args []*ast.Argument
			for _, arg := range d.Args {
				args = append(args, ast.NewArgument(&ast.Argument{
					Name: ast.NewName(&ast.Name{
						Value: arg.Name,
					}),
					Value: ast.NewStringValue(ast.NewStringValue(&ast.StringValue{
						Value: arg.Value.(string),
					})),
				}))
			}

			directives = append(directives, ast.NewDirective(&ast.Directive{
				Name: ast.NewName(&ast.Name{
					Value: d.Directive.Name,
				}),
				Arguments: args,
			}))
		}

		fields = append(fields, ast.NewFieldDefinition(&ast.FieldDefinition{
			Name: ast.NewName(&ast.Name{
				Value: name,
			}),
			Description: ast.NewStringValue(&ast.StringValue{Value: object.Description}),
			Arguments:   argumentAsNode(object.Args),
			Directives:  directives,
			Type:        typeAsNode(object.Type),
		}))
	}

	node := ast.NewObjectDefinition(&ast.ObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: o.Name(),
		}),
		Description: ast.NewStringValue(&ast.StringValue{Value: o.Description()}),
		Directives:  directivesAsNode(o.directives),
		Fields:      fields,
	})
	if o.extend {
		return ast.NewTypeExtensionDefinition(&ast.TypeExtensionDefinition{Definition: node})
	}
	return node
}

func inputObjectAsNode(o *InputObject) *ast.InputObjectDefinition {
	var fields []*ast.InputValueDefinition
	for name, object := range o.Fields() {
		fields = append(fields, ast.NewInputValueDefinition(&ast.InputValueDefinition{
			Name: ast.NewName(&ast.Name{
				Value: name,
			}),
			Description: ast.NewStringValue(&ast.StringValue{Value: object.Description()}),
			Directives:  []*ast.Directive{},
			Type:        typeAsNode(object.Type),
		}))
	}
	return ast.NewInputObjectDefinition(&ast.InputObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: o.Name(),
		}),
		Description: ast.NewStringValue(&ast.StringValue{Value: o.Description()}),
		Fields:      fields,
	})
}

func enumAsNode(o *Enum) *ast.EnumDefinition {
	var enumValues []*ast.EnumValueDefinition
	for _, v := range o.values {
		enumValues = append(enumValues, ast.NewEnumValueDefinition(&ast.EnumValueDefinition{
			Name: ast.NewName(&ast.Name{
				Value: v.Name,
			}),
			Description: ast.NewStringValue(&ast.StringValue{Value: v.Description}),
		}))
	}
	return ast.NewEnumDefinition(&ast.EnumDefinition{
		Name: ast.NewName(&ast.Name{
			Value: o.Name(),
		}),
		Description: ast.NewStringValue(&ast.StringValue{Value: o.Description()}),
		Values:      enumValues,
	})
}

func scalarAsNode(o *Scalar) *ast.ScalarDefinition {
	return ast.NewScalarDefinition(&ast.ScalarDefinition{
		Name: ast.NewName(&ast.Name{
			Value: o.Name(),
		}),
		Description: ast.NewStringValue(&ast.StringValue{Value: o.Description()}),
	})
}

func unionAsNode(o *Union) *ast.UnionDefinition {
	var types []*ast.Named

	for _, t := range o.Types() {
		types = append(types, ast.NewNamed(&ast.Named{
			Name: ast.NewName(&ast.Name{
				Value: t.Name(),
			}),
		}))
	}

	return ast.NewUnionDefinition(&ast.UnionDefinition{
		Name: ast.NewName(&ast.Name{
			Value: o.Name(),
		}),
		Types: types,
	})
}

func typeAstNode(tp Type, options *SDLExportOptions) ast.Node {
	switch o := tp.(type) {
	case *InputObject:
		return inputObjectAsNode(o)
	case *Object:
		return objectAsNode(o, options)
	case *Enum:
		return enumAsNode(o)
	case *Scalar:
		return scalarAsNode(o)
	case *Union:
		return unionAsNode(o)
	}
	return nil
}

func BuildSDL(schema Schema, option *SDLExportOptions) string {
	doc := ast.Document{}
	for name, tp := range schema.typeMap {
		if option != nil {
			if option.ExcludeDoubleUnderscorePrefix && strings.HasPrefix(name, "__") {
				continue
			}
			if !option.IncludeBasicScalar {
				switch name {
				case "ID", "String", "Int", "Boolean":
					continue
				}
			}
			if option.ExcludeQueryService && name == "_Service" {
				continue
			}
		}
		doc.Definitions = append(doc.Definitions, typeAstNode(tp, option))
	}

	printed := printer.Print(ast.NewDocument(&doc))
	return printed.(string)

}

type SDLExportOptions struct {
	ExcludeDoubleUnderscorePrefix bool
	ExcludeQueryService           bool
	IncludeBasicScalar            bool
}
